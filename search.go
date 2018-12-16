package main

import (
	"fmt"
	"time"
)

var _ = fmt.Println

var CHECKMATE_FLAG byte = 0x80
var STALEMATE_FLAG byte = 0x40
var CHECK_FLAG byte = 0x20
var THREEFOLD_REP_FLAG byte = 0x10

type SearchResult struct {
	move  Move
	value int
	flags byte
	nodes uint
	time  int64
	depth uint
	pv    string
}

type SearchConfig struct {
	alpha         int
	beta          int
	move          Move
	isDebug       bool
	startingDepth uint
}

type ExternalSearchConfig struct {
	isDebug bool
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
		startingDepth: depth,
	})
	result.time = (time.Now().UnixNano() - startTime) / 10000000

	return result
}

func searchAlphaBeta(boardState *BoardState, depth uint, currentDepth uint, searchConfig SearchConfig) SearchResult {
	if depth == 0 {
		return getTerminalResult(boardState, searchConfig)
	}

	isDebug := searchConfig.isDebug

	entry := ProbeTranspositionTable(boardState)
	var hashMove = make([]Move, 0)
	if entry != nil {
		if entry.depth >= depth {
			// I guess move might be bogus so we shouldn't just trust this
			return *entry.result
		}

		move := entry.result.move
		// Hash move might be bogus due to hash collision, so we need to validate
		// that it's a legit move (later)
		hashMove = append(hashMove, move)
	}

	var nodes uint
	listing := GenerateMoveListing(boardState)
	var bestResult *SearchResult

	for i, moveList := range [][]Move{hashMove, listing.promotions, listing.captures, listing.moves} {
		for _, move := range moveList {
			if i == 0 {
				// validate hash move is one of our generated moves
				// TODO: is this the best way?  we could just generate the moves from the square in
				// the hash move prior to generating the full move listing.  feels like this would be
				// effective at pruning the search tree and also save a bunch of time computing the full
				// move listing.

				isLegal := false
			CheckHashMoveLegality:
				for _, legalMoveListing := range [][]Move{listing.promotions, listing.captures, listing.moves} {
					for _, legalMove := range legalMoveListing {
						if move == legalMove {
							isLegal = true
							break CheckHashMoveLegality
						}
					}
				}

				if !isLegal {
					break
				}
			}

			if move.IsCastle() && !boardState.TestCastleLegality(move) {
				continue
			}

			boardState.ApplyMove(move)

			if !boardState.IsInCheck(!boardState.whiteToMove) {
				searchConfig.move = move
				searchConfig.isDebug = false

				searchDepth := depth - 1
				if move.IsCapture() || boardState.IsInCheck(boardState.whiteToMove) {
					searchDepth++
				}

				result := searchAlphaBeta(boardState, searchDepth, currentDepth+1, searchConfig)

				result.move = move
				result.depth++
				nodes += result.nodes

				if bestResult == nil {
					bestResult = &result
				} else {
					if (!boardState.whiteToMove && result.value > bestResult.value) || // white move, maximize score
						(boardState.whiteToMove && result.value < bestResult.value) { // black move, minimize score
						bestResult = &result
					}
				}

				if isDebug {
					fmt.Printf("[%d; %s] result=%d\n", depth, MoveToString(move), result.value)
				}

				if !boardState.whiteToMove {
					searchConfig.alpha = Max(searchConfig.alpha, bestResult.value)
				} else {
					searchConfig.beta = Min(searchConfig.beta, bestResult.value)
				}
			}

			boardState.UnapplyMove(move)

			if searchConfig.alpha >= searchConfig.beta {
				break
			}
		}
	}

	if bestResult == nil {
		result := getNoLegalMoveResult(boardState, depth, searchConfig)
		bestResult = &result
	} else {
		bestResult.pv = MoveToPrettyString(bestResult.move, boardState) + " " + bestResult.pv
	}

	bestResult.nodes = nodes
	StoreTranspositionTable(boardState, bestResult, depth)

	return *bestResult
}

func getTerminalResult(boardState *BoardState, searchConfig SearchConfig) SearchResult {
	// TODO(perf): use an incremental evaluation state passed in as an argument

	e, hasMatingMaterial := Eval(boardState)
	if !hasMatingMaterial {
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

func getNoLegalMoveResult(boardState *BoardState, depth uint, searchConfig SearchConfig) SearchResult {
	if boardState.IsInCheck(boardState.whiteToMove) {
		// moves to mate = startingDepth - depth
		movesToMate := searchConfig.startingDepth - depth
		score := CHECKMATE_SCORE - int(movesToMate)
		if boardState.whiteToMove {
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
		pv:    "1/2-1/2",
	}

}

func SearchResultToString(result SearchResult) string {
	return fmt.Sprintf("%s (value=%d, nodes=%d)", MoveToString(result.move), result.value, result.nodes)
}
