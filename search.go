package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var _ = fmt.Println

var CHECKMATE_FLAG byte = 0x80
var DRAW_FLAG byte = 0x40
var CHECK_FLAG byte = 0x20
var THREEFOLD_REP_FLAG byte = 0x10

type SearchStats struct {
	leafNodes    uint
	branchNodes  uint
	qBranchNodes uint
	hashCutoffs  uint
	cutoffs      uint
}

type SearchResult struct {
	move  Move
	value int
	flags byte
	time  int64
	depth uint
	stats SearchStats
	pv    string
}

type Variation struct {
	move     [MAX_MOVES]Move
	numMoves uint
}

func (result *SearchResult) IsCheckmate() bool {
	return result.flags&CHECKMATE_FLAG == CHECKMATE_FLAG
}

func (result *SearchResult) IsDraw() bool {
	return result.flags&DRAW_FLAG == DRAW_FLAG
}

type SearchConfig struct {
	move          Move
	isDebug       bool
	debugMoves    string
	startingDepth uint
}

type ExternalSearchConfig struct {
	isDebug         bool
	debugMoves      string
	onlySearchDepth uint
}

func Search(boardState *BoardState, depth uint) SearchResult {
	return SearchWithConfig(boardState, depth, ExternalSearchConfig{})
}

func SearchWithConfig(boardState *BoardState, depth uint, config ExternalSearchConfig) SearchResult {
	startTime := time.Now().UnixNano()

	stats := SearchStats{}
	variation := Variation{}
	score := searchAlphaBeta(boardState, &stats, &variation, depth, 0, -INFINITY, INFINITY, SearchConfig{
		isDebug:       config.isDebug,
		debugMoves:    config.debugMoves,
		startingDepth: depth,
	}, MoveSizeHint{})

	result := SearchResult{}
	if boardState.offsetToMove == BLACK_OFFSET {
		score = -score
	}
	absScore := score
	if score < 0 {
		absScore = -score
	}
	if absScore > CHECKMATE_SCORE-100 {
		result.flags = CHECKMATE_FLAG
	}
	result.value = score
	result.time = (time.Now().UnixNano() - startTime) / 10000000
	result.move = variation.move[0]
	result.pv = MoveArrayToPrettyString(variation.move[0:variation.numMoves], boardState)
	result.stats = stats
	result.depth = depth

	return result
}

// searchAlphaBeta runs an alpha-beta search over the boardState
//
// - depth, when positive, is the number of levels until quiescent search begins.
//   when negative it is the number of levels deep in quiescent search.
// - currentDepth is the total depth of the search tree, including quiescent levels.
//  - alpha: minimum score that player to move can achieve
//  - black: maximum score that opponent can achieve
//
func searchAlphaBeta(
	boardState *BoardState,
	searchStats *SearchStats,
	variation *Variation,
	depthLeft uint,
	currentDepth uint,
	alpha int,
	beta int,
	searchConfig SearchConfig,
	hint MoveSizeHint,
) int {
	var line Variation

	isDebug := searchConfig.isDebug

	var hashMove = make([]Move, 0)
	if entry := ProbeTranspositionTable(boardState); entry != nil {
		if entry.depth >= depthLeft {
			switch entry.entryType {
			case TT_EXACT:
				return entry.score
			case TT_FAIL_HIGH:
				if entry.score >= beta {
					return beta
				}
			case TT_FAIL_LOW:
				if entry.score <= alpha {
					return alpha
				}
			}
		}
		move := entry.move
		if _, err := boardState.IsMoveLegal(move); err == nil {
			hashMove = append(hashMove, move)
		}
	}

	// TODO: probably don't check this
	if boardState.IsThreefoldRepetition() {
		return 0
	}

	if depthLeft == 0 || boardState.shouldAbort {
		score := getLeafResult(boardState, searchConfig, searchStats)
		StoreTranspositionTable(boardState, Move{}, score, TT_EXACT, depthLeft)

		return score
	}

	searchStats.branchNodes++

	// We'll generate the other moves after we test the hash move
	// 0 = hash
	// 1 = captures
	// 2 = promotions
	// 3 = checks
	// 4 = moves
	var moveOrdering [5][]Move
	moveOrdering[0] = hashMove

	hasLegalMove := false

	bestScore := -INFINITY
	var bestMove Move

	for i := 0; i < len(moveOrdering); i++ {
		for _, move := range moveOrdering[i] {
			if move.IsCastle() && !boardState.TestCastleLegality(move) {
				continue
			}

			boardState.ApplyMove(move)

			if boardState.IsInCheck(oppositeColorOffset(boardState.offsetToMove)) {
				boardState.UnapplyMove(move)
				continue
			}
			hasLegalMove = true
			searchConfig.move = move
			searchConfig.isDebug = false

			score := -searchAlphaBeta(boardState, searchStats, &line, depthLeft-1, currentDepth+1, -beta, -alpha,
				searchConfig, hint)
			boardState.UnapplyMove(move)

			if score > beta {
				StoreTranspositionTable(boardState, move, score, TT_FAIL_HIGH, depthLeft)
				return score
			}

			if score > bestScore {
				bestScore = score
				bestMove = move
				if bestScore > alpha {
					alpha = score
					variation.move[0] = move
					copy(variation.move[1:], line.move[0:line.numMoves])
					variation.numMoves = line.numMoves + 1
				}
			}

			if isDebug && (strings.Contains(MoveToPrettyString(move, boardState), searchConfig.debugMoves) ||
				searchConfig.debugMoves == "*") {
				fmt.Printf("[%d; %s] value=%d alpha=%d beta=%d pv=%s\n", depthLeft, MoveToString(move), score,
					alpha, beta, MoveArrayToString(line.move[0:line.numMoves]))
			}
		}

		if i == 0 {
			var listing MoveListing
			// add the other moves now that we're done with hash move
			listing, hint = GenerateMoveListing(boardState, hint)

			moveOrdering[1] = listing.captures
			moveOrdering[2] = listing.promotions
			moveOrdering[3] = boardState.FilterChecks(moveOrdering[4])
			moveOrdering[4] = listing.moves
		}
	}

	// IF WE HAD NO LEGAL MOVES, GAME IS OVER

	if !hasLegalMove {
		score := getNoLegalMoveResult(boardState, currentDepth, searchConfig)
		StoreTranspositionTable(boardState, Move{}, score, TT_EXACT, depthLeft)

		return score
	}

	var ttEntryType int
	if bestScore < alpha {
		// never raised alpha
		ttEntryType = TT_FAIL_LOW
	} else {
		// we raised alpha, so we have an exact match
		ttEntryType = TT_EXACT
	}

	StoreTranspositionTable(boardState, bestMove, bestScore, ttEntryType, depthLeft)

	return bestScore
}

func getLeafResult(boardState *BoardState, searchConfig SearchConfig, searchStats *SearchStats) int {
	// TODO(perf): use an incremental evaluation state passed in as an argument
	searchStats.leafNodes++

	e := Eval(boardState)
	if !e.hasMatingMaterial {
		return 0
	}

	return e.value()
}

func getNoLegalMoveResult(boardState *BoardState, currentDepth uint, searchConfig SearchConfig) int {
	if boardState.IsInCheck(boardState.offsetToMove) {
		// moves to mate = currentDepth
		return -(CHECKMATE_SCORE - int(currentDepth) + 1)
	}

	// Stalemate
	return 0

}

<<<<<<< HEAD
func SearchResultToString(result SearchResult) string {
	return fmt.Sprintf("%s-%s (value=%s, depth=%d, stats=%s, pv=%s)",
		SquareToAlgebraicString(result.move.from),
		SquareToAlgebraicString(result.move.to),
		SearchValueToString(result),
=======
func (result *SearchResult) String() string {
	return fmt.Sprintf("%s (value=%s, depth=%d, stats=%s, pv=%s)",
		MoveToString(result.move),
		SearchValueToString(*result),
>>>>>>> Search: fix alpha/TT return/store value
		result.depth,
		SearchStatsToString(result.stats),
		result.pv)
}

func SearchValueToString(result SearchResult) string {
	if result.IsCheckmate() {
		score := result.value
		if score < 0 {
			score = -score
		}
		movesToCheckmate := CHECKMATE_SCORE - score + 1
		return fmt.Sprintf("Mate(%d)", movesToCheckmate)
	}

	if result.IsDraw() {
		return fmt.Sprintf("Draw")
	}

	return strconv.Itoa(result.value)
}

func SearchStatsToString(stats SearchStats) string {
	return fmt.Sprintf("[nodes=%d, leafNodes=%d, branchNodes=%d, qBranchNodes=%d, cutoffs=%d, hash cutoffs=%d]",
		stats.Nodes(),
		stats.leafNodes,
		stats.branchNodes,
		stats.qBranchNodes,
		stats.cutoffs,
		stats.hashCutoffs)
}

func (stats *SearchStats) Nodes() uint {
	return stats.branchNodes + stats.leafNodes + stats.qBranchNodes
}

// Used to determine if we should extend search
func (m Move) IsQuiescentPawnPush(boardState *BoardState) bool {
	movePiece := boardState.board[m.from]
	if movePiece&0x0F != PAWN_MASK {
		return false
	}

	rank := Rank(m.to)
	return (movePiece == WHITE_MASK|PAWN_MASK && rank == 7) || (rank == 2)
}
