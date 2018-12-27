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

	result := searchAlphaBeta(boardState, depth, 0, -INFINITY, INFINITY, SearchConfig{
		isDebug:       config.isDebug,
		debugMoves:    config.debugMoves,
		startingDepth: depth,
		phase:         SEARCH_PHASE_INITIAL,
	}, MoveSizeHint{})
	result.time = (time.Now().UnixNano() - startTime) / 10000000
	if boardState.offsetToMove == BLACK_OFFSET {
		result.value = -result.value
	}

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
	depth uint,
	currentDepth uint,
	alpha int,
	beta int,
	searchConfig SearchConfig,
	hint MoveSizeHint,
) SearchResult {
	// fmt.Printf("phase=%d depth left=%d current depth=%d\n", searchConfig.phase, depth, currentDepth)
	if boardState.IsThreefoldRepetition() {
		return SearchResult{
			value: 0,
			flags: DRAW_FLAG,
			pv:    "1/2-1/2 (Threefold Repetition)",
		}
	}

	if (depth == 0 && searchConfig.phase == SEARCH_PHASE_QUIESCENT) || boardState.shouldAbort {
		return getLeafResult(boardState, searchConfig)
	}

	isDebug := searchConfig.isDebug

	entry := ProbeTranspositionTable(boardState)
	var hashMove = make([]Move, 0)
	if entry != nil {
		move := entry.result.move
		if _, err := boardState.IsMoveLegal(move); err == nil {
			if entry.depth >= depth && entry.searchPhase <= searchConfig.phase {
				return *entry.result
			}

			hashMove = append(hashMove, move)
		}
	}

	var searchStats SearchStats
	searchStats.branchNodes++
	if searchConfig.phase == SEARCH_PHASE_QUIESCENT {
		searchStats.qBranchNodes++
	}

	var bestResult *SearchResult

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

FindBestMove:
	for i := 0; i < len(moveOrdering); i++ {
		for _, move := range moveOrdering[i] {
			if move.IsCastle() && !boardState.TestCastleLegality(move) {
				continue
			}

			boardState.ApplyMove(move)
			if !boardState.IsInCheck(oppositeColorOffset(boardState.offsetToMove)) {
				searchConfig.move = move
				searchConfig.isDebug = false

				result := searchAlphaBeta(boardState, searchDepth, currentDepth+1, -beta, -alpha, searchConfig, hint)
				boardState.UnapplyMove(move)

				result.value = -result.value
				result.move = move
				result.depth++
				addSearchStats(&searchStats, result.stats)

				if bestResult == nil {
					bestResult = &result
				} else if result.value > bestResult.value {
					bestResult = &result
				}

				if isDebug && (strings.Contains(MoveToPrettyString(move, boardState), searchConfig.debugMoves) ||
					strings.Contains(result.pv, searchConfig.debugMoves) ||
					searchConfig.debugMoves == "*") {
					fmt.Printf("[%d; %s] value=%d, result=%s, pv=%s\n", depth,
						MoveToString(move, boardState), result.value, SearchResultToString(result), result.pv)
				}

				alpha = Max(alpha, result.value)
			} else {
				boardState.UnapplyMove(move)
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

	if bestResult == nil {
		var result SearchResult
		if phase == SEARCH_PHASE_INITIAL {
			result = getNoLegalMoveResult(boardState, currentDepth, searchConfig)
		} else {
			// This means we don't check for checkmates/stalemates in quiescent search
			// and just evaluate the board.
			result = getLeafResult(boardState, searchConfig)
		}
		bestResult = &result
	} else {
		separator := " "
		if phase != searchConfig.phase {
			separator = " <Q> "
		}
		bestResult.pv = MoveToPrettyString(bestResult.move, boardState) + separator + bestResult.pv
	}

	bestResult.stats = searchStats
	StoreTranspositionTable(boardState, bestResult, depth, searchConfig.phase)

	return *bestResult
}

func addSearchStats(searchStats *SearchStats, add SearchStats) {
	searchStats.cutoffs += add.cutoffs
	searchStats.leafNodes += add.leafNodes
	searchStats.branchNodes += add.branchNodes
	searchStats.qBranchNodes += add.qBranchNodes
	searchStats.hashCutoffs += searchStats.hashCutoffs
}

func getLeafResult(boardState *BoardState, searchConfig SearchConfig) SearchResult {
	// TODO(perf): use an incremental evaluation state passed in as an argument

	e := Eval(boardState)
	var res SearchResult
	if !e.hasMatingMaterial {
		res = SearchResult{
			value: 0,
			flags: DRAW_FLAG,
			pv:    "1/2-1/2 (Insufficient mating material)",
		}
	} else {
		res = SearchResult{
			value: e.value(),
			move:  searchConfig.move,
			pv:    "",
		}
	}

	res.stats.leafNodes = 1
	return res
}

func getNoLegalMoveResult(boardState *BoardState, currentDepth uint, searchConfig SearchConfig) SearchResult {
	if boardState.IsInCheck(boardState.offsetToMove) {
		// moves to mate = currentDepth
		return SearchResult{
			value: -(CHECKMATE_SCORE - int(currentDepth) + 1),
			flags: CHECKMATE_FLAG,
			pv:    "#",
		}
	}

	// Stalemate
	return SearchResult{
		value: 0,
		flags: DRAW_FLAG,
		pv:    "1/2-1/2 (no legal moves)",
	}

}

func SearchResultToString(result SearchResult) string {
	return fmt.Sprintf("%s-%s (value=%s, depth=%d, stats=%s, pv=%s)",
		SquareToAlgebraicString(result.move.from),
		SquareToAlgebraicString(result.move.to),
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
