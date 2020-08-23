package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoubledPawnBitboard(t *testing.T) {
	bitboard := SetBitboard(SetBitboard(0, SQUARE_D4), SQUARE_E5)
	assert.Equal(t, uint64(0), GetDoubledPawnBitboard(bitboard))

	bitboard = SetBitboard(SetBitboard(0, SQUARE_D4), SQUARE_E4)
	assert.Equal(t, uint64(0), GetDoubledPawnBitboard(bitboard))

	bitboard = SetBitboard(SetBitboard(0, SQUARE_D4), SQUARE_D3)
	assert.Equal(t, bitboard, GetDoubledPawnBitboard(bitboard))

	bitboard = SetBitboard(SetBitboard(SetBitboard(0, SQUARE_A4), SQUARE_A3), SQUARE_A2)
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
	bitboard := SetBitboard(SetBitboard(0, SQUARE_D4), SQUARE_H4)
	assert.Equal(t, bitboard, GetPawnRankBitboard(bitboard, RANK_4))
}
