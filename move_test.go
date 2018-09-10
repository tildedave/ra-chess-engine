package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"unsafe"
)

func TestCreateMove(t *testing.T) {
	var m = CreateMove(31, 51)

	assert.Equal(t, uintptr(3), unsafe.Sizeof(m))
	assert.Equal(t, Move{from: 31, to: 51}, m)
}

func TestCreateCapture(t *testing.T) {
	var m = CreateCapture(31, 51)

	assert.Equal(t, Move{from: 31, to: 51, flags: 0x80}, m)
}

func TestMoveToString(t *testing.T) {
	assert.Equal(t, "a2xa4", MoveToString(CreateCapture(SQUARE_A2, SQUARE_A4)))
	assert.Equal(t, "a2-a4", MoveToString(CreateMove(SQUARE_A2, SQUARE_A4)))
	assert.Equal(t, "O-O", MoveToString(CreateKingsideCastle(25, 27)))
	assert.Equal(t, "O-O-O", MoveToString(CreateQueensideCastle(25, 23)))
}

func TestMoveToPrettyString(t *testing.T) {
	var boardState BoardState = CreateEmptyBoardState()

	boardState.board[SQUARE_A2] = WHITE_MASK | PAWN_MASK
	boardState.board[SQUARE_F5] = WHITE_MASK | KNIGHT_MASK
	boardState.board[SQUARE_G7] = BLACK_MASK | ROOK_MASK

	assert.Equal(t, "a2xa4", MoveToPrettyString(CreateCapture(SQUARE_A2, SQUARE_A4), boardState))
	assert.Equal(t, "a4", MoveToPrettyString(CreateMove(SQUARE_A2, SQUARE_A4), boardState))
	assert.Equal(t, "Nxg7", MoveToPrettyString(CreateCapture(SQUARE_F5, SQUARE_G7), boardState))
}
