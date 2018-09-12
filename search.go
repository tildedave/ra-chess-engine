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
		// TODO(perf): use an incremental evaluation state passed in as an argument

		e := Eval(boardState)
		return SearchResult{
			value: e.forBoardState(boardState).value(),
			move:  searchConfig.move,
			line:  searchConfig.line,
		}
	}

	var nodes = 0

	moves := GenerateMoves(boardState)
	var bestResult *SearchResult = nil
	searchConfig.depth -= 1

	for _, move := range moves {
		if move.IsCastle() && !boardState.TestCastleLegality(move) {
			continue
		}

		var whiteToMove = boardState.whiteToMove
		boardState.ApplyMove(move)
		if !boardState.IsInCheck(!boardState.whiteToMove) {
			searchConfig.line = append(searchConfig.line, move)
			searchConfig.move = move

			nodes += 1
			result := searchMinimax(boardState, searchConfig)
			result.move = move
			searchConfig.line = searchConfig.line[:len(searchConfig.line)-1]

			if bestResult == nil {
				if searchConfig.debug {
					fmt.Printf("New best move found - by default %s (depth=%d)\n", SearchResultToString(result), searchConfig.depth)
				}
				bestResult = &result
			} else {
				if searchConfig.debug {
					fmt.Printf("%s (best=%s)\n", SearchResultToString(result), SearchResultToString(*bestResult))
				}

				if (whiteToMove && result.value > bestResult.value) || (!whiteToMove && result.value < bestResult.value) {
					if searchConfig.debug {
						fmt.Println("New best move found!")
					}
					bestResult = &result
				}
			}
		}

		boardState.UnapplyMove(move)

	}

	if bestResult == nil {
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

	bestResult.nodes += nodes
	return *bestResult
}

func SearchResultToString(result SearchResult) string {
	return fmt.Sprintf("%s (value=%d, nodes=%d)", MoveToString(result.move), result.value, result.nodes)
}
