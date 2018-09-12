package main

import (
	"fmt"
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
	nodes int
	line  []Move
}

type SearchConfig struct {
	depth uint
	debug bool
	move  Move
	line  []Move
}

func search(boardState *BoardState, depth uint) SearchResult {
	// TODO(search): alpha-beta search instead of minimax
	return searchMinimax(boardState, SearchConfig{depth: depth, debug: false})
}

func searchMinimax(boardState *BoardState, searchConfig SearchConfig) SearchResult {
	if searchConfig.depth == 0 {
		return getTerminalResult(boardState, searchConfig)
	}

	var nodes = 0

	moves := GenerateMoves(boardState)
	var bestResult *SearchResult = nil
	searchConfig.depth -= 1

	for _, move := range moves {
		if move.IsCastle() && !boardState.TestCastleLegality(move) {
			continue
		}

		boardState.ApplyMove(move)
		if !boardState.IsInCheck(!boardState.whiteToMove) {
			searchConfig.line = append(searchConfig.line, move)
			searchConfig.move = move

			nodes += 1
			result := searchMinimax(boardState, searchConfig)
			result.move = move
			searchConfig.line = searchConfig.line[:len(searchConfig.line)-1]

			if bestResult == nil {
				bestResult = &result
			} else {
				if (!boardState.whiteToMove && result.value > bestResult.value) || // white move, maximize score
					(boardState.whiteToMove && result.value < bestResult.value) { // black move, minimize score
					bestResult = &result
				}
			}
		}

		boardState.UnapplyMove(move)

	}

	if bestResult == nil {
		return getNoLegalMoveResult(boardState, searchConfig)
	}

	bestResult.nodes += nodes
	return *bestResult
}

func getTerminalResult(boardState *BoardState, searchConfig SearchConfig) SearchResult {
	// TODO(perf): use an incremental evaluation state passed in as an argument

	e := Eval(boardState)
	return SearchResult{
		value: e.forBoardState(boardState).value(),
		move:  searchConfig.move,
		line:  searchConfig.line,
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
			line:  searchConfig.line,
		}
	}

	// Stalemate
	return SearchResult{
		value: 0,
		flags: STALEMATE_FLAG,
		line:  searchConfig.line,
	}

}

func SearchResultToString(result SearchResult) string {
	return fmt.Sprintf("%s (value=%d, nodes=%d)", MoveToString(result.move), result.value, result.nodes)
}
