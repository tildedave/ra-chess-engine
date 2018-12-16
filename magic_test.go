package main

import (
	"fmt"
	"math"
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
	assert.Equal(t, uint64(0x8080876080800), RookMask(SQUARE_D4))
	assert.Equal(t, 10, bits.OnesCount64(RookMask(SQUARE_D4)))
	assert.Equal(t, uint64(0x101010101017e), RookMask(SQUARE_A1))
	assert.Equal(t, 12, bits.OnesCount64(RookMask(SQUARE_A1)))
	assert.Equal(t, uint64(0x202020202027c), RookMask(SQUARE_B1))
	assert.Equal(t, 11, bits.OnesCount64(RookMask(SQUARE_B1)))
	assert.Equal(t, uint64(0x7e80808080808000), RookMask(SQUARE_H8))
	assert.Equal(t, 12, bits.OnesCount64(RookMask(SQUARE_H8)))
}

func TestBishopMask(t *testing.T) {
	assert.Equal(t, uint64(0x40221400142200), BishopMask(SQUARE_D4))
	assert.Equal(t, 9, bits.OnesCount64(BishopMask(SQUARE_D4)))
	assert.Equal(t, uint64(0x40201008040200), BishopMask(SQUARE_A1))
	assert.Equal(t, 6, bits.OnesCount64(BishopMask(SQUARE_A1)))
	assert.Equal(t, uint64(0x402010080400), BishopMask(SQUARE_B1))
	assert.Equal(t, 5, bits.OnesCount64(BishopMask(SQUARE_B1)))
	assert.Equal(t, uint64(0x2040810204000), BishopMask(SQUARE_H1))
	assert.Equal(t, 6, bits.OnesCount64(BishopMask(SQUARE_H1)))
	assert.Equal(t, uint64(0x4020100a0000), BishopMask(SQUARE_C2))
	assert.Equal(t, 5, bits.OnesCount64(BishopMask(SQUARE_C2)))
}

func TestBishopMoveBoard(t *testing.T) {
	var occupancies uint64
	occupancies = SetBitboard(occupancies, SQUARE_B2)
	occupancies = SetBitboard(occupancies, SQUARE_B6)
	occupancies = SetBitboard(occupancies, SQUARE_E5)

	assert.Equal(t, uint64(0x21400142240), BishopMoveBoard(SQUARE_D4, occupancies))
}

func TestRookOccupancies(t *testing.T) {
	assert.Equal(t, 1024, len(GenerateRookOccupancies(SQUARE_D4, false)))
	assert.Equal(t, 4096, len(GenerateRookOccupancies(SQUARE_A1, false)))
	assert.Equal(t, 4096, len(GenerateRookOccupancies(SQUARE_H1, false)))

	assert.Equal(t, 16384, len(GenerateRookOccupancies(SQUARE_D4, true)))
	assert.Equal(t, 16384, len(GenerateRookOccupancies(SQUARE_A1, true)))
	assert.Equal(t, 16384, len(GenerateRookOccupancies(SQUARE_H1, true)))
}

func TestBishopOccupancies(t *testing.T) {
	assert.Equal(t, 64, len(GenerateBishopOccupancies(SQUARE_A1, false)))
	assert.Equal(t, 128, len(GenerateBishopOccupancies(SQUARE_A1, true)))
	assert.Equal(t, 32, len(GenerateBishopOccupancies(SQUARE_D1, false)))
	assert.Equal(t, 128, len(GenerateBishopOccupancies(SQUARE_D1, true)))
}

func TestRookMoveBoard(t *testing.T) {
	var occupancies uint64
	occupancies = SetBitboard(occupancies, SQUARE_D1)
	occupancies = SetBitboard(occupancies, SQUARE_G3)
	occupancies = SetBitboard(occupancies, SQUARE_D6)

	assert.Equal(t, uint64(0x80808770808), RookMoveBoard(SQUARE_D3, occupancies))
}

func TestGenerateRookSlidingMoves(t *testing.T) {
	magics, err := inputMagicFile("rook-magics.json")
	if err != nil {
		panic(err)
	}

	attacks := make([]SquareAttacks, math.MaxInt16)

	a1Magic := magics[SQUARE_A1]
	GenerateRookSlidingMoves(SQUARE_A1, a1Magic, attacks)

	var bitboard uint64
	bitboard = SetBitboard(bitboard, SQUARE_A5)
	bitboard = SetBitboard(bitboard, SQUARE_G1)

	key := hashKey(bitboard, a1Magic)
	assert.Equal(t, 10, len(attacks[key].moves))
	assertMovePresent(t, attacks[key].moves, SQUARE_A1, SQUARE_A2)
	assertMovePresent(t, attacks[key].moves, SQUARE_A1, SQUARE_A3)
	assertMovePresent(t, attacks[key].moves, SQUARE_A1, SQUARE_A4)
	assertMovePresent(t, attacks[key].moves, SQUARE_A1, SQUARE_A5)
	assertMovePresent(t, attacks[key].moves, SQUARE_A1, SQUARE_B1)
	assertMovePresent(t, attacks[key].moves, SQUARE_A1, SQUARE_C1)
	assertMovePresent(t, attacks[key].moves, SQUARE_A1, SQUARE_D1)
	assertMovePresent(t, attacks[key].moves, SQUARE_A1, SQUARE_E1)
	assertMovePresent(t, attacks[key].moves, SQUARE_A1, SQUARE_F1)
	assertMovePresent(t, attacks[key].moves, SQUARE_A1, SQUARE_G1)

	attacks = make([]SquareAttacks, math.MaxInt16)

	d3Magic := magics[SQUARE_D3]
	GenerateRookSlidingMoves(SQUARE_D3, d3Magic, attacks)

	bitboard = 0
	bitboard = SetBitboard(bitboard, SQUARE_D2)
	key = hashKey(bitboard, d3Magic)
	assert.Equal(t, 13, len(attacks[key].moves))

	assertMovePresent(t, attacks[key].moves, SQUARE_D3, SQUARE_D2)
	assertMovePresent(t, attacks[key].moves, SQUARE_D3, SQUARE_D4)
	assertMovePresent(t, attacks[key].moves, SQUARE_D3, SQUARE_D5)
	assertMovePresent(t, attacks[key].moves, SQUARE_D3, SQUARE_D6)
	assertMovePresent(t, attacks[key].moves, SQUARE_D3, SQUARE_D7)
	assertMovePresent(t, attacks[key].moves, SQUARE_D3, SQUARE_D8)
	assertMoveNotPresent(t, attacks[key].moves, SQUARE_D3, SQUARE_D1)
}

func TestGenerateBishopSlidingMoves(t *testing.T) {
	magics, err := inputMagicFile("bishop-magics.json")
	if err != nil {
		panic(err)
	}

	attacks := make([]SquareAttacks, math.MaxInt16)
	d3Magic := magics[SQUARE_D3]
	GenerateBishopSlidingMoves(SQUARE_D3, d3Magic, attacks)

	var bitboard uint64
	bitboard = SetBitboard(bitboard, SQUARE_C2)
	bitboard = SetBitboard(bitboard, SQUARE_F1)
	bitboard = SetBitboard(bitboard, SQUARE_G6)

	key := hashKey(bitboard, d3Magic)
	assert.Equal(t, 9, len(attacks[key].moves))

	assertMovePresent(t, attacks[key].moves, SQUARE_D3, SQUARE_E2)
	assertMovePresent(t, attacks[key].moves, SQUARE_D3, SQUARE_F1)
	assertMovePresent(t, attacks[key].moves, SQUARE_D3, SQUARE_E4)
	assertMovePresent(t, attacks[key].moves, SQUARE_D3, SQUARE_F5)
	assertMovePresent(t, attacks[key].moves, SQUARE_D3, SQUARE_G6)
	assertMovePresent(t, attacks[key].moves, SQUARE_D3, SQUARE_C2)
	assertMovePresent(t, attacks[key].moves, SQUARE_D3, SQUARE_C4)
	assertMovePresent(t, attacks[key].moves, SQUARE_D3, SQUARE_B5)
	assertMovePresent(t, attacks[key].moves, SQUARE_D3, SQUARE_A6)
}
