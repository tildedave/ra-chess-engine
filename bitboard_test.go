package main

import (
	"math/bits"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLegacySquareToBitboardSquare(t *testing.T) {
	assert.Equal(t, byte(0), legacySquareToBitboardSquare(SQUARE_A1))
	assert.Equal(t, byte(7), legacySquareToBitboardSquare(SQUARE_H1))
	assert.Equal(t, byte(8), legacySquareToBitboardSquare(SQUARE_A2))
	assert.Equal(t, byte(15), legacySquareToBitboardSquare(SQUARE_H2))
	assert.Equal(t, byte(56), legacySquareToBitboardSquare(SQUARE_A8))
	assert.Equal(t, byte(63), legacySquareToBitboardSquare(SQUARE_H8))
}

func TestSetUnsetBitboard(t *testing.T) {
	bitboard := SetBitboard(0, SQUARE_D4)
	bitboard2 := UnsetBitboard(bitboard, SQUARE_D3)
	bitboard3 := UnsetBitboard(bitboard, SQUARE_D4)

	assert.Equal(t, bitboard, bitboard2)
	assert.True(t, IsBitboardSet(bitboard, SQUARE_D4))
	assert.False(t, IsBitboardSet(bitboard2, SQUARE_D3))
	assert.Equal(t, uint64(0), bitboard3)
}

func TestInitialBitboards(t *testing.T) {
	boardState := CreateInitialBoardState()

	assert.Equal(t, 16, bits.OnesCount64(boardState.bitboards.color[WHITE_OFFSET]))
	assert.Equal(t, 16, bits.OnesCount64(boardState.bitboards.color[BLACK_OFFSET]))
	assert.Equal(t, 16, bits.OnesCount64(boardState.bitboards.piece[PAWN_MASK]))
	assert.Equal(t, 4, bits.OnesCount64(boardState.bitboards.piece[KNIGHT_MASK]))
	assert.Equal(t, 4, bits.OnesCount64(boardState.bitboards.piece[BISHOP_MASK]))
	assert.Equal(t, 4, bits.OnesCount64(boardState.bitboards.piece[ROOK_MASK]))
	assert.Equal(t, 2, bits.OnesCount64(boardState.bitboards.piece[KING_MASK]))
	assert.Equal(t, 2, bits.OnesCount64(boardState.bitboards.piece[QUEEN_MASK]))
}
