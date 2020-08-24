package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _ = fmt.Println

func TestEvalEmptyBoard(t *testing.T) {
	testBoard := CreateEmptyBoardState()
	boardEval := Eval(&testBoard)

	assert.Equal(t, 0, boardEval.material)
}

func TestEvalPawn(t *testing.T) {
	testBoard := CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|PAWN_MASK)

	boardEval := Eval(&testBoard)

	assert.Equal(t, 100, boardEval.material)
}

func TestEvalPawnAgainstBishop(t *testing.T) {
	testBoard := CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A3, BLACK_MASK|BISHOP_MASK)

	boardEval := Eval(&testBoard)

	assert.Equal(t, -205, boardEval.material)
}

func TestEvalPassedPawns(t *testing.T) {
	testBoard := CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|PAWN_MASK)

	boardEval := Eval(&testBoard)
	fmt.Println(boardEval)
}

func TestEvalStartingPosition(t *testing.T) {
	testBoard := CreateInitialBoardState()
	boardEval := Eval(&testBoard)

	assert.Equal(t, boardEval.material, 0)
}

func TestEvalStartingPositionCenterControl(t *testing.T) {
	testBoard := CreateInitialBoardState()
	testBoard.SetPieceAtSquare(SQUARE_E2, 0x00)
	testBoard.SetPieceAtSquare(SQUARE_E4, WHITE_MASK|PAWN_MASK)
	boardEval := Eval(&testBoard)

	assert.Equal(t, boardEval.material, 0)
}

func TestEvalKingSafety(t *testing.T) {
	testBoard := CreateInitialBoardState()

	testBoard.SetPieceAtSquare(SQUARE_E1, 0x00)
	testBoard.SetPieceAtSquare(SQUARE_F1, WHITE_MASK|ROOK_MASK)
	testBoard.SetPieceAtSquare(SQUARE_G1, WHITE_MASK|KING_MASK)

	boardEval := Eval(&testBoard)

	assert.Equal(t, boardEval.kingPosition, KING_PAWN_COVER_EVAL_SCORE*3+KING_IN_CENTER_EVAL_SCORE)
}
