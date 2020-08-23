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

	bitboard = SetBitboard(SetBitboard(SetBitboard(0, SQUARE_A4), SQUARE_A3), A2)
	assert.Equal(t, bitboard, GetDoubledPawnBitboard(bitboard))
}
