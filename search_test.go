package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var _ = fmt.Println

// Checkmate is Queen C3 to C1
func CreateMateInOneBoard() BoardState {
	boardState := CreateEmptyBoardState()
	boardState.board[SQUARE_A1] = BLACK_MASK | KING_MASK
	boardState.board[SQUARE_C3] = WHITE_MASK | QUEEN_MASK
	boardState.board[SQUARE_B4] = WHITE_MASK | KNIGHT_MASK
	boardState.board[SQUARE_H8] = WHITE_MASK | KING_MASK
	generateBoardLookupInfo(&boardState)

	return boardState
}

func TestSearchStartingPosition(t *testing.T) {
	boardState := CreateInitialBoardState()

	result := search(&boardState, 4)

	assert.Equal(t, 0, result.value)
}

func TestSearchMateInOne(t *testing.T) {
	boardState := CreateMateInOneBoard()

	result := search(&boardState, 2)

	assert.Equal(t, CHECKMATE_SCORE, result.value)
	assert.Equal(t, Move{from: SQUARE_C3, to: SQUARE_C1}, result.move)
}

func TestSearchMateInOneBlack(t *testing.T) {
	boardState := CreateMateInOneBoard()
	FlipBoardColors(&boardState)

	result := search(&boardState, 2)
	assert.Equal(t, -CHECKMATE_SCORE, result.value)
	assert.Equal(t, Move{from: SQUARE_C3, to: SQUARE_C1}, result.move)
}

func TestSearchAvoidMateInOne(t *testing.T) {
	boardState := CreateMateInOneBoard()
	boardState.whiteToMove = false
	boardState.board[SQUARE_H3] = BLACK_MASK | ROOK_MASK

	result := search(&boardState, 1)

	assert.True(t, result.move.IsCapture())
	assert.Equal(t, Move{from: SQUARE_H3, to: SQUARE_C3, flags: CAPTURE_MASK}, result.move)
}

func TestSearchPawnPromotion(t *testing.T) {
	boardState := CreateEmptyBoardState()
	boardState.board[SQUARE_A7] = WHITE_MASK | PAWN_MASK
	boardState.board[SQUARE_A8] = BLACK_MASK | KING_MASK
	boardState.board[SQUARE_A6] = WHITE_MASK | KING_MASK
	generateBoardLookupInfo(&boardState)

	result := search(&boardState, 10)

	assert.False(t, result.move.IsCapture())
}
