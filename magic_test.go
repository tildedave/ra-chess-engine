package main

import (
	"fmt"
	"math/bits"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertMovePresent(t *testing.T, moves []Move, fromSq byte, toSq byte) {
	found := false
	for _, move := range moves {
		if move.from == fromSq && move.to == toSq {
			found = true
		}
	}

	assert.True(t, found, fmt.Sprintf("Move from=%d to=%d was not present", fromSq, toSq))
}

func assertMoveNotPresent(t *testing.T, moves []Move, fromSq byte, toSq byte) {
	found := false
	for _, move := range moves {
		if move.from == fromSq && move.to == toSq {
			found = true
		}
	}

	assert.False(t, found, fmt.Sprintf("Move from=%d to=%d was present", fromSq, toSq))
}

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

	assert.Equal(t, uint64(0x21400142240), BishopMoveBoard(BB_SQUARE_D4, occupancies))
}

func TestRookOccupancies(t *testing.T) {
	assert.Equal(t, 1024, len(GenerateRookOccupancies(BB_SQUARE_D4, false)))
	assert.Equal(t, 4096, len(GenerateRookOccupancies(BB_SQUARE_A1, false)))
	assert.Equal(t, 4096, len(GenerateRookOccupancies(BB_SQUARE_H1, false)))

	assert.Equal(t, 16384, len(GenerateRookOccupancies(BB_SQUARE_D4, true)))
	assert.Equal(t, 16384, len(GenerateRookOccupancies(BB_SQUARE_A1, true)))
	assert.Equal(t, 16384, len(GenerateRookOccupancies(BB_SQUARE_H1, true)))
}

func TestBishopOccupancies(t *testing.T) {
	assert.Equal(t, 64, len(GenerateBishopOccupancies(BB_SQUARE_A1, false)))
	assert.Equal(t, 128, len(GenerateBishopOccupancies(BB_SQUARE_A1, true)))
	assert.Equal(t, 32, len(GenerateBishopOccupancies(BB_SQUARE_D1, false)))
	assert.Equal(t, 128, len(GenerateBishopOccupancies(BB_SQUARE_D1, true)))
}

func TestRookMoveBoard(t *testing.T) {
	var occupancies uint64
	occupancies = SetBitboard(occupancies, BB_SQUARE_D1)
	occupancies = SetBitboard(occupancies, BB_SQUARE_G3)
	occupancies = SetBitboard(occupancies, BB_SQUARE_D6)

	assert.Equal(t, uint64(0x80808770808), RookMoveBoard(BB_SQUARE_D3, occupancies))
}

func TestGenerateRookSlidingMoves(t *testing.T) {
	magics, err := inputMagicFile("rook-magics.json")
	if err != nil {
		panic(err)
	}

	moves := make(map[uint16][]Move)
	a1Magic := magics[BB_SQUARE_A1]
	GenerateRookSlidingMoves(BB_SQUARE_A1, a1Magic, moves)

	var bitboard uint64
	bitboard = SetBitboard(bitboard, BB_SQUARE_A5)
	bitboard = SetBitboard(bitboard, BB_SQUARE_G1)

	key := uint16(((bitboard & a1Magic.Mask) * a1Magic.Magic) >> (64 - a1Magic.Bits))
	assert.Equal(t, 10, len(moves[key]))
	assertMovePresent(t, moves[key], BB_SQUARE_A1, BB_SQUARE_A2)
	assertMovePresent(t, moves[key], BB_SQUARE_A1, BB_SQUARE_A3)
	assertMovePresent(t, moves[key], BB_SQUARE_A1, BB_SQUARE_A4)
	assertMovePresent(t, moves[key], BB_SQUARE_A1, BB_SQUARE_A5)
	assertMovePresent(t, moves[key], BB_SQUARE_A1, BB_SQUARE_B1)
	assertMovePresent(t, moves[key], BB_SQUARE_A1, BB_SQUARE_C1)
	assertMovePresent(t, moves[key], BB_SQUARE_A1, BB_SQUARE_D1)
	assertMovePresent(t, moves[key], BB_SQUARE_A1, BB_SQUARE_E1)
	assertMovePresent(t, moves[key], BB_SQUARE_A1, BB_SQUARE_F1)
	assertMovePresent(t, moves[key], BB_SQUARE_A1, BB_SQUARE_G1)

	moves = make(map[uint16][]Move)
	d3Magic := magics[BB_SQUARE_D3]
	GenerateRookSlidingMoves(BB_SQUARE_D3, d3Magic, moves)

	bitboard = 0
	bitboard = SetBitboard(bitboard, BB_SQUARE_D2)
	key = uint16(((bitboard & d3Magic.Mask) * d3Magic.Magic) >> (64 - d3Magic.Bits))
	assert.Equal(t, 13, len(moves[key]))

	assertMovePresent(t, moves[key], BB_SQUARE_D3, BB_SQUARE_D2)
	assertMovePresent(t, moves[key], BB_SQUARE_D3, BB_SQUARE_D4)
	assertMovePresent(t, moves[key], BB_SQUARE_D3, BB_SQUARE_D5)
	assertMovePresent(t, moves[key], BB_SQUARE_D3, BB_SQUARE_D6)
	assertMovePresent(t, moves[key], BB_SQUARE_D3, BB_SQUARE_D7)
	assertMovePresent(t, moves[key], BB_SQUARE_D3, BB_SQUARE_D8)
	assertMoveNotPresent(t, moves[key], BB_SQUARE_D3, BB_SQUARE_D1)
}

func TestGenerateBishopSlidingMoves(t *testing.T) {
	magics, err := inputMagicFile("bishop-magics.json")
	if err != nil {
		panic(err)
	}

	moves := make(map[uint16][]Move)
	d3Magic := magics[BB_SQUARE_D3]
	GenerateBishopSlidingMoves(BB_SQUARE_D3, d3Magic, moves)

	var bitboard uint64
	bitboard = SetBitboard(bitboard, BB_SQUARE_C2)
	bitboard = SetBitboard(bitboard, BB_SQUARE_F1)
	bitboard = SetBitboard(bitboard, BB_SQUARE_G6)

	key := uint16(((bitboard & d3Magic.Mask) * d3Magic.Magic) >> (64 - d3Magic.Bits))
	assert.Equal(t, 9, len(moves[key]))

	assertMovePresent(t, moves[key], BB_SQUARE_D3, BB_SQUARE_E2)
	assertMovePresent(t, moves[key], BB_SQUARE_D3, BB_SQUARE_F1)
	assertMovePresent(t, moves[key], BB_SQUARE_D3, BB_SQUARE_E4)
	assertMovePresent(t, moves[key], BB_SQUARE_D3, BB_SQUARE_F5)
	assertMovePresent(t, moves[key], BB_SQUARE_D3, BB_SQUARE_G6)
	assertMovePresent(t, moves[key], BB_SQUARE_D3, BB_SQUARE_C2)
	assertMovePresent(t, moves[key], BB_SQUARE_D3, BB_SQUARE_C4)
	assertMovePresent(t, moves[key], BB_SQUARE_D3, BB_SQUARE_B5)
	assertMovePresent(t, moves[key], BB_SQUARE_D3, BB_SQUARE_A6)
}
