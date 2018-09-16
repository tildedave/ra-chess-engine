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
	depth uint
	time  int64
}

type SearchConfig struct {
	depth uint
	alpha int
	beta  int
	move  Move
}

func Search(boardState *BoardState, depth uint) SearchResult {
	startTime := time.Now().UnixNano()
	result := searchAlphaBeta(boardState, SearchConfig{
		depth: depth,
		alpha: -INFINITY,
		beta:  INFINITY,
	})
	result.time = (time.Now().UnixNano() - startTime) / 10000000

	return result
}

func searchAlphaBeta(boardState *BoardState, searchConfig SearchConfig) SearchResult {
	if searchConfig.depth == 0 {
		return getTerminalResult(boardState, searchConfig)
	}

	var nodes uint = 0

	moves := GenerateMoves(boardState)
	var bestResult *SearchResult = nil
	searchConfig.depth -= 1

	for _, move := range moves {
		if move.IsCastle() && !boardState.TestCastleLegality(move) {
			continue
		}

		boardState.ApplyMove(move)
		if !boardState.IsInCheck(!boardState.whiteToMove) {
			searchConfig.move = move

			result := searchAlphaBeta(boardState, searchConfig)
			result.move = move
			nodes += result.nodes

			if bestResult == nil {
				bestResult = &result
			} else {
				if (!boardState.whiteToMove && result.value > bestResult.value) || // white move, maximize score
					(boardState.whiteToMove && result.value < bestResult.value) { // black move, minimize score
					bestResult = &result
				}
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
	bestResult.depth += 1
	return *bestResult
}

func getTerminalResult(boardState *BoardState, searchConfig SearchConfig) SearchResult {
	// TODO(perf): use an incremental evaluation state passed in as an argument

	e := Eval(boardState)
	return SearchResult{
		value: e.value(),
		move:  searchConfig.move,
		nodes: 1,
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
		}
	}

	// Stalemate
	return SearchResult{
		value: 0,
		flags: STALEMATE_FLAG,
	}

}

func SearchResultToString(result SearchResult) string {
	return fmt.Sprintf("%s (value=%d, nodes=%d)", MoveToString(result.move), result.value, result.nodes)
}
