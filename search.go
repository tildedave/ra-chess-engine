package main

import (
	"fmt"
	"sort"
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
	alpha   int
	beta    int
	move    Move
	isDebug bool
}

type ExternalSearchConfig struct {
	isDebug bool
}

func Search(boardState *BoardState, depth uint) SearchResult {
	return SearchWithConfig(boardState, depth, ExternalSearchConfig{})
}

func SearchWithConfig(boardState *BoardState, depth uint, config ExternalSearchConfig) SearchResult {
	startTime := time.Now().UnixNano()

	result := searchAlphaBeta(boardState, depth, SearchConfig{
		alpha:   -INFINITY,
		beta:    INFINITY,
		isDebug: config.isDebug,
	})
	result.time = (time.Now().UnixNano() - startTime) / 10000000

	return result
}

func searchAlphaBeta(boardState *BoardState, depth uint, searchConfig SearchConfig) SearchResult {
	if depth == 0 {
		return getTerminalResult(boardState, searchConfig)
	}

	var nodes uint

	moves := GenerateMoves(boardState)
	var bestResult *SearchResult

	// [Ordering] prioritize captures first
	// Future: hash move (transposition table), killer move
	sort.Slice(moves, func(i int, j int) bool {
		return (moves[i].flags & 0xC0) > (moves[j].flags & 0xC0)
	})

	isDebug := searchConfig.isDebug
	for _, move := range moves {
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

			result := searchAlphaBeta(boardState, searchDepth, searchConfig)

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

	if bestResult == nil {
		return getNoLegalMoveResult(boardState, searchConfig)
	}

	bestResult.nodes = nodes
	bestResult.pv = MoveToPrettyString(bestResult.move, boardState) + " " + bestResult.pv
	return *bestResult
}

func getTerminalResult(boardState *BoardState, searchConfig SearchConfig) SearchResult {
	// TODO(perf): use an incremental evaluation state passed in as an argument

	e := Eval(boardState)
	return SearchResult{
		value: e.value(),
		move:  searchConfig.move,
		nodes: 1,
		pv:    "",
	}
}

func getNoLegalMoveResult(boardState *BoardState, searchConfig SearchConfig) SearchResult {
	if boardState.IsInCheck(boardState.whiteToMove) {
		score := CHECKMATE_SCORE
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
