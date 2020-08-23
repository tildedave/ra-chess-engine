package main

import (
	"math/bits"
)

// Given a bitboard which is the pawns for a single side, return a bitboard which
// contains only the doubled pawns.
func GetDoubledPawnBitboard(pawnBitboard uint64) uint64 {
	var doubledBitboard uint64
	originalBoard := pawnBitboard
	for pawnBitboard != 0 {
		sq := byte(bits.TrailingZeros64(pawnBitboard))
		pawnBitboard ^= 1 << sq

		col := sq % 8
		row := sq / 8

		var columnBoard uint64
		for j := byte(0); j < 8; j++ {
			if j == row {
				continue
			}
			columnBoard = SetBitboard(columnBoard, 8*j+col)
		}

		overlapBoard := (columnBoard & originalBoard)
		doubledBitboard |= overlapBoard

		if overlapBoard != 0 {
			// Since we detected a doubled-pawn from this square, it is itself doubled
			doubledBitboard ^= 1 << sq

			// For every pawn in the overlap, we don't need to check its column again
			for overlapBoard != 0 {
				sq := byte(bits.TrailingZeros64(overlapBoard))
				overlapBoard ^= 1 << sq
				pawnBitboard ^= 1 << sq
			}
		}
	}
	return doubledBitboard
}
