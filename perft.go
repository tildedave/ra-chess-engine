package main

import (
	"fmt"
)

var _ = fmt.Println

func Perft(boardState *BoardState, depth int) uint64 {
	var moveCount uint64 = 0
	if depth == 0 {
		return 1
	}

	moves := GenerateMoves(boardState)

	for _, move := range moves {
		// testMoveLegality(boardState, move)
		// fmt.Println(MoveToString(move))
		boardState.ApplyMove(move)
		if !boardState.IsInCheck(!boardState.whiteToMove) {
			moveCount += Perft(boardState, depth-1)
		}
		boardState.UnapplyMove(move)
	}

	return moveCount
}

func testMoveLegality(boardState *BoardState, move Move) {
	legal, err := boardState.IsMoveLegal(move)
	if !legal {
		fmt.Println(err)
		fmt.Println(boardState.ToString())
		fmt.Println(MoveToString(move))
		panic("Illegal move")
	}
}
