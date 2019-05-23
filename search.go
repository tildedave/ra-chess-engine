package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var CHECKMATE_FLAG byte = 0x80
var DRAW_FLAG byte = 0x40
var CHECK_FLAG byte = 0x20
var THREEFOLD_REP_FLAG byte = 0x10

type SearchStats struct {
	leafnodes              uint
	branchnodes            uint
	qbranchnodes           uint
	hashcutoffs            uint
	cutoffs                uint
	qcutoffs               uint
	qcapturesfiltered      uint
	transpositiontablehits uint
}

type SearchResult struct {
	move  Move
	value int
	flags byte
	time  time.Duration
	depth uint
	stats SearchStats
	pv    string
}

type Variation struct {
	move     [MAX_MOVES]Move
	numMoves uint
}

func CopyVariation(line *Variation, restLine *Variation) {
	copy(line.move[0:], restLine.move[0:restLine.numMoves])
	line.numMoves = restLine.numMoves
}

func CopyVariationWithMove(line *Variation, move Move, restLine *Variation) {
	line.move[0] = move
	copy(line.move[1:], restLine.move[0:restLine.numMoves])
	line.numMoves = restLine.numMoves + 1
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
	isDebug       bool
	debugMoves    string
	searchToDepth uint
}

func Search(boardState *BoardState, depth uint) SearchResult {
	return SearchWithConfig(boardState, depth, ExternalSearchConfig{})
}

func SearchWithConfig(boardState *BoardState, depth uint, config ExternalSearchConfig) SearchResult {
	startTime := time.Now()

	stats := SearchStats{}
	variation := Variation{}
	score := searchAlphaBeta(boardState, &stats, &variation, int(depth), 0, -INFINITY, INFINITY, SearchConfig{
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
	result.time = time.Now().Sub(startTime)
	result.move = variation.move[0]
	result.pv, _ = MoveArrayToPrettyString(variation.move[0:variation.numMoves], boardState)
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
	depthLeft int,
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
				searchStats.transpositiontablehits++
				return entry.score
			case TT_FAIL_HIGH:
				if entry.score >= beta {
					searchStats.transpositiontablehits++
					return beta
				}
			case TT_FAIL_LOW:
				if entry.score <= alpha {
					searchStats.transpositiontablehits++
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

	if boardState.shouldAbort {
		score := getLeafResult(boardState, searchStats)
		StoreTranspositionTable(boardState, Move{}, score, TT_EXACT, depthLeft)

		return score
	}

	if depthLeft == 0 {
		score := searchQuiescent(boardState, searchStats, &line, depthLeft, currentDepth, alpha, beta, searchConfig, hint)

		ttEntryType := TT_FAIL_LOW
		if score >= beta {
			ttEntryType = TT_FAIL_HIGH
		}

		if score > alpha {
			ttEntryType = TT_EXACT
			CopyVariation(variation, &line)
		}

		StoreTranspositionTable(boardState, Move{}, score, ttEntryType, depthLeft)
		return score
	}

	searchStats.branchnodes++

	// We'll generate the other moves after we test the hash move
	// 0 = hash
	// 1 = captures
	// 2 = promotions
	// 3 = checks
	// 4 = moves
	var moveOrdering [5][]Move
	moveOrdering[0] = hashMove

	hasLegalMove := false

	bestScore := -INFINITY + 1
	var bestMove Move

	for i := 0; i < len(moveOrdering); i++ {
		for _, move := range moveOrdering[i] {
			if move.IsCastle() && !boardState.TestCastleLegality(move) {
				continue
			}

			debugMode := isDebug && (strings.Contains(MoveToPrettyString(move, boardState), searchConfig.debugMoves) ||
				searchConfig.debugMoves == "*")

			offset := boardState.offsetToMove
			boardState.ApplyMove(move)
			if boardState.IsInCheck(offset) {
				boardState.UnapplyMove(move)
				continue
			}
			hasLegalMove = true
			searchConfig.move = move
			var nodesStarting uint
			var hashKey uint64

			if debugMode {
				nodesStarting = searchStats.Nodes()
				hashKey = boardState.hashKey
			}
			searchConfig.isDebug = false

			score := -searchAlphaBeta(boardState, searchStats, &line, depthLeft-1, currentDepth+1, -beta, -alpha,
				searchConfig, hint)

			boardState.UnapplyMove(move)

			if debugMode {
				var moveLine Variation
				CopyVariationWithMove(&moveLine, move, &line)
				str := MoveArrayToXboardString(moveLine.move[0:moveLine.numMoves])
				entry := boardState.transpositionTable[hashKey]
				var entryType string
				if entry != nil {
					entryType = EntryTypeToString(entry.entryType)
				}
				fmt.Printf("[%d; %s] value=%d (alpha=%d, beta=%d) nodes=%d pv=%s result=%s\n",
					depthLeft, MoveToString(move, boardState), score, alpha, beta, searchStats.Nodes()-nodesStarting, str,
					entryType)
			}

			if score > beta {
				StoreTranspositionTable(boardState, move, score, TT_FAIL_HIGH, depthLeft)
				searchStats.cutoffs++
				if i == 0 {
					searchStats.hashcutoffs++
				}
				return score
			}

			if score > bestScore {
				bestScore = score
				bestMove = move
				if bestScore > alpha {
					alpha = score
					CopyVariationWithMove(variation, move, &line)
				}
			}
		}

		if i == 0 {
			var listing MoveListing
			// add the other moves now that we're done with hash move
			listing, hint = GenerateMoveListing(boardState, hint, true)

			moveOrdering[1] = listing.captures
			moveOrdering[2] = listing.promotions
			moveOrdering[3] = boardState.FilterChecks(moveOrdering[4])
			moveOrdering[4] = listing.moves
		}
	}

	// IF WE HAD NO LEGAL MOVES, GAME IS OVER

	if !hasLegalMove {
		score := getNoLegalMoveResult(boardState, currentDepth)
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

func searchQuiescent(
	boardState *BoardState,
	searchStats *SearchStats,
	variation *Variation,
	// depthLeft will always be negative
	depthLeft int,
	currentDepth uint,
	alpha int,
	beta int,
	searchConfig SearchConfig,
	hint MoveSizeHint,
) int {
	var line Variation
	var bestMove Move
	bestScore := -INFINITY + 1

	// Evaluate the board to see what the position is without making any quiescent moves.
	score := getLeafResult(boardState, searchStats)
	if score >= beta {
		StoreTranspositionTable(boardState, bestMove, bestScore, TT_FAIL_HIGH, depthLeft)
		return beta
	}
	bestScore = score

	if score >= alpha {
		alpha = score
	}
	if boardState.shouldAbort {
		return score
	}

	moveListing, hint := GenerateQuiescentMoveListing(boardState, hint)
	var moveOrdering [5][]Move
	moveOrdering[0] = boardState.FilterSEECaptures(moveListing.captures)
	searchStats.qcapturesfiltered += uint(len(moveListing.captures) - len(moveOrdering[0]))
	searchStats.qbranchnodes++

	for _, moves := range moveOrdering {
		ourOffset := boardState.offsetToMove
		for _, move := range moves {
			boardState.ApplyMove(move)
			if boardState.IsInCheck(ourOffset) {
				boardState.UnapplyMove(move)
				continue
			}

			score := -searchQuiescent(boardState, searchStats, &line, depthLeft-1, currentDepth+1, -beta, -alpha, searchConfig, hint)
			boardState.UnapplyMove(move)

			if score >= beta {
				searchStats.cutoffs++
				searchStats.qcutoffs++
				StoreTranspositionTable(boardState, bestMove, score, TT_FAIL_HIGH, depthLeft)

				return beta
			}

			if score > bestScore {
				bestScore = score
				bestMove = move
				if score > alpha {
					alpha = score
					variation.move[0] = move
					copy(variation.move[1:], line.move[0:line.numMoves])
					variation.numMoves = line.numMoves + 1
				}
			}
		}
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

func getLeafResult(boardState *BoardState, searchStats *SearchStats) int {
	// TODO(perf): use an incremental evaluation state passed in as an argument
	searchStats.leafnodes++

	e := Eval(boardState)
	if !e.hasMatingMaterial {
		return 0
	}

	return e.value()
}

func getNoLegalMoveResult(boardState *BoardState, currentDepth uint) int {
	if boardState.IsInCheck(boardState.offsetToMove) {
		// moves to mate = currentDepth
		return -(CHECKMATE_SCORE - int(currentDepth) + 1)
	}

	// Stalemate
	return 0

}

func (result *SearchResult) String() string {
	return fmt.Sprintf("%s (time=%s, value=%s, depth=%d, stats=%s, pv=%s)",
		MoveToXboardString(result.move),
		result.time.String(),
		SearchValueToString(*result),
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
	return fmt.Sprintf("[nodes=%d, transpositiontablehits=%d, leafnodes=%d, branchnodes=%d, qbranchnodes=%d, cutoffs=%d, hash cutoffs=%d, qcutoffs=%d, qcapturesfiltered=%d]",
		stats.Nodes(),
		stats.transpositiontablehits,
		stats.leafnodes,
		stats.branchnodes,
		stats.qbranchnodes,
		stats.cutoffs,
		stats.hashcutoffs,
		stats.qcutoffs,
		stats.qcapturesfiltered)
}

func (stats *SearchStats) Nodes() uint {
	return stats.branchnodes + stats.leafnodes + stats.qbranchnodes
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
