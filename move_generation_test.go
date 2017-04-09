package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var _ = fmt.Println

func TestMoveGenerationWorks(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.board[SQUARE_A2] = WHITE_MASK | KING_MASK
	testBoard.board[SQUARE_A3] = WHITE_MASK | PAWN_MASK
	testBoard.board[SQUARE_A1] = BLACK_MASK | PAWN_MASK

	moves := GenerateMoves(&testBoard)

	var movesFromKing []Move
	numCaptures := 0
	for _, move := range moves {
		if move.from == SQUARE_A2 {
			movesFromKing = append(movesFromKing, move)
			if move.IsCapture() {
				numCaptures += 1
			}
		}
	}

	assert.Equal(t, 4, len(movesFromKing))
	assert.Equal(t, 1, numCaptures)

	assert.Equal(t, 1, 1)
}
