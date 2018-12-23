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

func (result *SearchResult) IsCheckmate() bool {
	return result.flags&CHECKMATE_FLAG == CHECKMATE_FLAG
}

func (result *SearchResult) IsDraw() bool {
	return result.flags&DRAW_FLAG == DRAW_FLAG
}

const (
	SEARCH_PHASE_INITIAL   = iota
	SEARCH_PHASE_QUIESCENT = iota
)

const QUIESCENT_SEARCH_DEPTH = 3

type SearchConfig struct {
	move          Move
	isDebug       bool
	debugMoves    string
	startingDepth uint
	phase         int
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
	score := searchAlphaBeta(boardState, &stats, depth, 0, -INFINITY, INFINITY, SearchConfig{
		isDebug:       config.isDebug,
		debugMoves:    config.debugMoves,
		startingDepth: depth,
		phase:         SEARCH_PHASE_INITIAL,
	}, MoveSizeHint{})

	result := SearchResult{}
	if boardState.offsetToMove == BLACK_OFFSET {
		score = -score
	}
	result.value = score
	result.time = (time.Now().UnixNano() - startTime) / 10000000
	e := ProbeTranspositionTable(boardState)
	result.move = e.move
	result.pv = MoveToPrettyString(result.move, boardState)
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
	depth uint,
	currentDepth uint,
	alpha int,
	beta int,
	searchConfig SearchConfig,
	hint MoveSizeHint,
) int {
	// TODO: probably don't check this
	if boardState.IsThreefoldRepetition() {
		return 0
	}

	if (depth == 0 && searchConfig.phase == SEARCH_PHASE_QUIESCENT) || boardState.shouldAbort {
		return getLeafResult(boardState, searchConfig, searchStats)
	}

	isDebug := searchConfig.isDebug

	entry := ProbeTranspositionTable(boardState)
	var hashMove = make([]Move, 0)
	if entry != nil {
		move := entry.move
		if _, err := boardState.IsMoveLegal(move); err == nil {
			if entry.depth >= depth && entry.searchPhase <= searchConfig.phase {
				return entry.score
			}

			hashMove = append(hashMove, move)
		}
	}

	searchStats.branchNodes++
	if searchConfig.phase == SEARCH_PHASE_QUIESCENT {
		searchStats.qBranchNodes++
	}

	// We'll generate the other moves after we test the hash move
	// 0 = hash
	// 1 = captures
	// 2 = promotions
	// 3 = checks
	// 4 = moves
	var moveOrdering [5][]Move
	moveOrdering[0] = hashMove

	// phase will change below
	phase := searchConfig.phase
	searchDepth := depth - 1
	if searchDepth == 0 && phase == SEARCH_PHASE_INITIAL {
		searchConfig.phase = SEARCH_PHASE_QUIESCENT
		// just do this for now
		searchDepth = QUIESCENT_SEARCH_DEPTH
	}

	hasLegalMove := false

	// Stupid temporary variable for quiescent search checkmate detection
	var allMoves []Move
	bestScore := -INFINITY
	var bestMove Move

FindBestMove:
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

			score := -searchAlphaBeta(boardState, searchStats, searchDepth, currentDepth+1, -beta, -alpha, searchConfig, hint)
			boardState.UnapplyMove(move)

			if score > beta {
				StoreTranspositionTable(boardState, move, beta, TT_FAIL_HIGH, currentDepth, searchConfig.phase)
				return beta
			}

			if score > bestScore {
				bestScore = score
				bestMove = move
				if bestScore > alpha {
					alpha = score
				}
			}

			if isDebug && (strings.Contains(MoveToPrettyString(move, boardState), searchConfig.debugMoves) ||
				searchConfig.debugMoves == "*") {
				fmt.Printf("[%d; %s] value=%d\n", depth, MoveToString(move), score)
			}

			if alpha >= beta {
				if i == 0 {
					searchStats.hashCutoffs++
				}
				searchStats.cutoffs++
				break FindBestMove
			}
		}

		if i == 0 {
			var listing MoveListing
			// add the other moves now that we're done with hash move
			listing, hint = GenerateMoveListing(boardState, hint)

			allMoves = listing.moves
			// You need to be allowed to leave check in quiescent phase
			if phase == SEARCH_PHASE_QUIESCENT && !boardState.IsInCheck(boardState.offsetToMove) {
				listing.moves = boardState.FilterChecks(listing.moves)
			}

			moveOrdering[1] = listing.captures
			moveOrdering[2] = listing.promotions
			moveOrdering[4] = listing.moves
		}

		// Quiescent search already only returns checks in listing.moves
		if i == 2 && phase == SEARCH_PHASE_INITIAL {
			moveOrdering[3] = boardState.FilterChecks(moveOrdering[4])
		}
	}

	// IF WE HAD NO LEGAL MOVES, GAME IS OVER
	if !hasLegalMove {
		if phase == SEARCH_PHASE_INITIAL {
			bestScore = getNoLegalMoveResult(boardState, currentDepth, searchConfig)
		} else {
			// qsearch
			for _, move := range allMoves {
				boardState.ApplyMove(move)
				inCheck := boardState.IsInCheck(oppositeColorOffset(boardState.offsetToMove))
				boardState.UnapplyMove(move)
				if !inCheck {
					hasLegalMove = true
					break
				}
			}

			if hasLegalMove {
				// some quiescent junk here
				bestScore = getLeafResult(boardState, searchConfig, searchStats)
			} else {
				bestScore = getNoLegalMoveResult(boardState, currentDepth, searchConfig)
			}
		}
	}

	var ttEntryType int
	if bestScore < alpha {
		// never raised alpha
		ttEntryType = TT_FAIL_LOW
	} else {
		// we raised alpha so we have an exact match
		ttEntryType = TT_EXACT
	}

	StoreTranspositionTable(boardState, bestMove, bestScore, ttEntryType, depth, searchConfig.phase)

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

func SearchResultToString(result SearchResult) string {
	return fmt.Sprintf("%s (value=%s, depth=%d, stats=%s, pv=%s)",
		MoveToString(result.move),
		SearchValueToString(result),
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
		movesToCheckmate := CHECKMATE_SCORE - score
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
