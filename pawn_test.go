package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoubledPawnBitboard(t *testing.T) {
	bitboard := SetBitboardMultiple(0, SQUARE_D4, SQUARE_E5)
	assert.Equal(t, uint64(0), GetDoubledPawnBitboard(bitboard))

	bitboard = SetBitboardMultiple(0, SQUARE_D4, SQUARE_E4)
	assert.Equal(t, uint64(0), GetDoubledPawnBitboard(bitboard))

	bitboard = SetBitboardMultiple(0, SQUARE_D4, SQUARE_D3)
	assert.Equal(t, bitboard, GetDoubledPawnBitboard(bitboard))

	bitboard = SetBitboardMultiple(0, SQUARE_A4, SQUARE_A3, SQUARE_A2)
	assert.Equal(t, bitboard, GetDoubledPawnBitboard(bitboard))
}

func TestPassedPawnBitboard(t *testing.T) {
	bitboard := SetBitboard(0, SQUARE_D4)
	otherBitboard := SetBitboard(0, SQUARE_D7)

	assert.Equal(t, uint64(0), GetPassedPawnBitboard(bitboard, otherBitboard, WHITE_OFFSET))
	assert.Equal(t, uint64(0), GetPassedPawnBitboard(otherBitboard, bitboard, BLACK_OFFSET))
	assert.Equal(t, bitboard, GetPassedPawnBitboard(bitboard, 0, WHITE_OFFSET))

	bitboard = SetBitboard(0, SQUARE_A2)
	otherBitboard = SetBitboard(0, SQUARE_H4)

	assert.Equal(t, bitboard, GetPassedPawnBitboard(bitboard, otherBitboard, WHITE_OFFSET))
	assert.Equal(t, otherBitboard, GetPassedPawnBitboard(otherBitboard, bitboard, BLACK_OFFSET))
}

func TestGetPawnRankBitboard(t *testing.T) {
	bitboard := SetBitboardMultiple(0, SQUARE_D4, SQUARE_H4)
	assert.Equal(t, bitboard, GetPawnRankBitboard(bitboard, RANK_4))
}

func TestGetIsolatedPawnBitboard(t *testing.T) {
	bitboard := SetBitboardMultiple(0, SQUARE_H5, SQUARE_G4)
	assert.Equal(t, uint64(0), GetIsolatedPawnBitboard(bitboard))

	bitboard = SetBitboard(0, SQUARE_D4)
	assert.Equal(t, bitboard, GetIsolatedPawnBitboard(bitboard))

	// From wikipedia
	boardState, _ := CreateBoardStateFromFENString("8/8/8/PP2P2P/2P3P1/4P3/8/8 w - - 0 1")
	pawnBoard := boardState.bitboards.piece[PAWN_MASK]

	assert.Equal(t,
		SetBitboardMultiple(0, SQUARE_E3, SQUARE_E5),
		GetIsolatedPawnBitboard(pawnBoard))
	assert.Equal(t,
		SetBitboardMultiple(0, SQUARE_A5, SQUARE_B5, SQUARE_C4, SQUARE_G4, SQUARE_H5),
		pawnBoard^GetIsolatedPawnBitboard(pawnBoard))
}
