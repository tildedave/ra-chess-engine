package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var _ = fmt.Println

var CHECKMATE_FLAG byte = 0x80
var STALEMATE_FLAG byte = 0x40
var CHECK_FLAG byte = 0x20
var THREEFOLD_REP_FLAG byte = 0x10

type SearchResult struct {
	move        Move
	value       int
	flags       byte
	nodes       uint
	hashCutoffs uint
	cutoffs     uint
	time        int64
	depth       uint
	pv          string
}

func (result *SearchResult) IsCheckmate() bool {
	return result.flags&CHECKMATE_FLAG == CHECKMATE_FLAG
}

func (result *SearchResult) IsStalemate() bool {
	return result.flags&STALEMATE_FLAG == STALEMATE_FLAG
}

const (
	SEARCH_PHASE_INITIAL   = iota
	SEARCH_PHASE_QUIESCENT = iota
)

const QUIESCENT_SEARCH_DEPTH = 3

type SearchConfig struct {
	alpha         int
	beta          int
	move          Move
	isDebug       bool
	debugMoves    string
	startingDepth uint
	phase         int
}

type ExternalSearchConfig struct {
	isDebug    bool
	debugMoves string
}

func Search(boardState *BoardState, depth uint) SearchResult {
	return SearchWithConfig(boardState, depth, ExternalSearchConfig{})
}

func SearchWithConfig(boardState *BoardState, depth uint, config ExternalSearchConfig) SearchResult {
	startTime := time.Now().UnixNano()

	result := searchAlphaBeta(boardState, depth, 0, SearchConfig{
		alpha:         -INFINITY,
		beta:          INFINITY,
		isDebug:       config.isDebug,
		debugMoves:    config.debugMoves,
		startingDepth: depth,
		phase:         SEARCH_PHASE_INITIAL,
	}, MoveSizeHint{})
	result.time = (time.Now().UnixNano() - startTime) / 10000000

	return result
}

func searchAlphaBeta(
	boardState *BoardState,
	depth uint,
	currentDepth uint,
	searchConfig SearchConfig,
	hint MoveSizeHint,
) SearchResult {
	// fmt.Printf("phase=%d depth left=%d current depth=%d\n", searchConfig.phase, depth, currentDepth)
	if (depth == 0 && searchConfig.phase == SEARCH_PHASE_QUIESCENT) || boardState.shouldAbort {
		return getTerminalResult(boardState, searchConfig)
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

	var nodes uint
	var hashCutoffs uint
	var cutoffs uint

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

				result := searchAlphaBeta(boardState, searchDepth, currentDepth+1, searchConfig, hint)

				result.move = move
				result.depth++
				nodes += result.nodes
				hashCutoffs += result.hashCutoffs
				cutoffs += result.cutoffs

				if bestResult == nil {
					bestResult = &result
				} else {
					if (boardState.offsetToMove == BLACK_OFFSET && result.value > bestResult.value) || // white move, maximize score
						(boardState.offsetToMove == WHITE_OFFSET && result.value < bestResult.value) { // black move, minimize score
						bestResult = &result
					}
				}

				if isDebug && (strings.Contains(MoveToString(move), searchConfig.debugMoves) ||
					strings.Contains(bestResult.pv, searchConfig.debugMoves)) {
					fmt.Printf("[%d; %s] value=%d, result=%s, pv=%s\n", depth,
						MoveToString(move), result.value, SearchResultToString(result), result.pv)
				}

				if boardState.offsetToMove == BLACK_OFFSET {
					searchConfig.alpha = Max(searchConfig.alpha, bestResult.value)
				} else {
					searchConfig.beta = Min(searchConfig.beta, bestResult.value)
				}
			}

			boardState.UnapplyMove(move)

			if searchConfig.alpha >= searchConfig.beta {
				if i == 0 {
					hashCutoffs++
				}
				cutoffs++
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
			result = getTerminalResult(boardState, searchConfig)
		}
		bestResult = &result
	} else {
		separator := " "
		if phase != searchConfig.phase {
			separator = " <Q> "
		}
		bestResult.pv = MoveToPrettyString(bestResult.move, boardState) + separator + bestResult.pv
	}

	bestResult.nodes = nodes
	bestResult.hashCutoffs = hashCutoffs
	bestResult.cutoffs = cutoffs
	StoreTranspositionTable(boardState, bestResult, depth, searchConfig.phase)

	return *bestResult
}

func getTerminalResult(boardState *BoardState, searchConfig SearchConfig) SearchResult {
	// TODO(perf): use an incremental evaluation state passed in as an argument

	e := Eval(boardState)
	if !e.hasMatingMaterial {
		return SearchResult{
			value: 0,
			flags: STALEMATE_FLAG,
			pv:    "1/2-1/2 (Insufficient mating material)",
		}
	}

	return SearchResult{
		value: e.value(),
		move:  searchConfig.move,
		nodes: 1,
		pv:    "",
	}
}

func getNoLegalMoveResult(boardState *BoardState, currentDepth uint, searchConfig SearchConfig) SearchResult {
	if boardState.IsInCheck(boardState.offsetToMove) {
		// moves to mate = currentDepth
		score := CHECKMATE_SCORE - int(currentDepth) + 1
		if boardState.offsetToMove == WHITE_OFFSET {
			score = -score
		}

		return SearchResult{
			value: score,
			flags: CHECKMATE_FLAG,
			pv:    "#",
		}
	}

	// Stalemate
	return SearchResult{
		value: 0,
		flags: STALEMATE_FLAG,
		pv:    "1/2-1/2 (no legal moves)",
	}

}

func SearchResultToString(result SearchResult) string {
	return fmt.Sprintf("%s (value=%s, depth=%d, nodes=%d, cutoffs=%d, hash cutoffs=%d, pv=%s)",
		MoveToString(result.move),
		SearchValueToString(result),
		result.depth,
		result.nodes,
		result.cutoffs,
		result.hashCutoffs,
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

	if result.IsStalemate() {
		return fmt.Sprintf("Stalemate")
	}

	return strconv.Itoa(result.value)
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
