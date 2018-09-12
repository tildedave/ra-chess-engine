package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var _ = fmt.Println

func TestEvalEmptyBoard(t *testing.T) {
	testBoard := CreateEmptyBoardState()
	boardEval := Eval(&testBoard)

	assert.Equal(t, 0, boardEval.material)
}

func TestEvalPawn(t *testing.T) {
	testBoard := CreateEmptyBoardState()
	testBoard.board[SQUARE_A2] = WHITE_MASK | PAWN_MASK

	boardEval := Eval(&testBoard)

	assert.Equal(t, 100, boardEval.material)
}

func TestEvalPawnAgainstBishop(t *testing.T) {
	testBoard := CreateEmptyBoardState()
	testBoard.board[SQUARE_A2] = WHITE_MASK | PAWN_MASK
	testBoard.board[SQUARE_A3] = BLACK_MASK | BISHOP_MASK

	boardEval := Eval(&testBoard)

	assert.Equal(t, -200, boardEval.material)
}

func TestEvalStartingPosition(t *testing.T) {
	testBoard := CreateInitialBoardState()
	boardEval := Eval(&testBoard)

	assert.Equal(t, boardEval.material, 0)
}
