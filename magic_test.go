package main

import (
	"math/bits"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRookMask(t *testing.T) {
	assert.Equal(t, uint64(0x8080876080800), RookMask(BB_SQUARE_D4))
	assert.Equal(t, 10, bits.OnesCount64(RookMask(BB_SQUARE_D4)))
	assert.Equal(t, uint64(0x101010101017e), RookMask(BB_SQUARE_A1))
	assert.Equal(t, 12, bits.OnesCount64(RookMask(BB_SQUARE_A1)))
	assert.Equal(t, uint64(0x202020202027c), RookMask(BB_SQUARE_B1))
	assert.Equal(t, 11, bits.OnesCount64(RookMask(BB_SQUARE_B1)))
	assert.Equal(t, uint64(0x7e80808080808000), RookMask(BB_SQUARE_H8))
	assert.Equal(t, 12, bits.OnesCount64(RookMask(BB_SQUARE_H8)))
}

func TestBishopMask(t *testing.T) {
	assert.Equal(t, uint64(0x40221400142200), BishopMask(BB_SQUARE_D4))
	assert.Equal(t, 9, bits.OnesCount64(BishopMask(BB_SQUARE_D4)))
	assert.Equal(t, uint64(0x40201008040200), BishopMask(BB_SQUARE_A1))
	assert.Equal(t, 6, bits.OnesCount64(BishopMask(BB_SQUARE_A1)))
	assert.Equal(t, uint64(0x402010080400), BishopMask(BB_SQUARE_B1))
	assert.Equal(t, 5, bits.OnesCount64(BishopMask(BB_SQUARE_B1)))
	assert.Equal(t, uint64(0x2040810204000), BishopMask(BB_SQUARE_H1))
	assert.Equal(t, 6, bits.OnesCount64(BishopMask(BB_SQUARE_H1)))
	assert.Equal(t, uint64(0x4020100a0000), BishopMask(BB_SQUARE_C2))
	assert.Equal(t, 5, bits.OnesCount64(BishopMask(BB_SQUARE_C2)))
}

func TestBishopMoveBoard(t *testing.T) {
	var occupancies uint64
	occupancies = SetBitboard(occupancies, BB_SQUARE_B2)
	occupancies = SetBitboard(occupancies, BB_SQUARE_B6)
	occupancies = SetBitboard(occupancies, BB_SQUARE_E5)

	assert.Equal(t, uint64(0x400142000), BishopMoveBoard(BB_SQUARE_D4, occupancies))
}

func TestRookMoveBoard(t *testing.T) {
	var occupancies uint64
	occupancies = SetBitboard(occupancies, BB_SQUARE_D1)
	occupancies = SetBitboard(occupancies, BB_SQUARE_G3)
	occupancies = SetBitboard(occupancies, BB_SQUARE_D6)

	assert.Equal(t, uint64(0x808360800), RookMoveBoard(BB_SQUARE_D3, occupancies))

}
