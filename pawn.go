package main

import (
	"math/bits"
)

type PawnTableEntry struct {
	pawns          [2]uint64
	passedPawns    [2]uint64
	doubledPawns   [2]uint64
	isolatedPawns  [2]uint64
	connectedPawns [2]uint64
	pawnsPerRank   [2][8]uint64
}

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
			columnBoard = SetBitboard(columnBoard, idx(col, j))
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

func GetPassedPawnBitboard(pawnBitboard uint64, otherSidePawnBitboard uint64, sideToMove int) uint64 {
	var passedPawnBoard uint64
	for pawnBitboard != 0 {
		sq := byte(bits.TrailingZeros64(pawnBitboard))
		pawnBitboard ^= 1 << sq

		col := sq % 8
		row := int(sq / 8)

		var columnBoard uint64
		var inc int
		if sideToMove == WHITE_OFFSET {
			inc = 1
		} else {
			inc = -1
		}
		for j := row + inc; j < 8 && j >= 0; j += inc {
			maskSq := idx(col, byte(j))
			columnBoard = SetBitboard(columnBoard, maskSq)
			if col > 0 {
				columnBoard = SetBitboard(columnBoard, maskSq-1)
			}
			if col < 7 {
				columnBoard = SetBitboard(columnBoard, maskSq+1)
			}
		}
		if otherSidePawnBitboard&columnBoard == 0 {
			passedPawnBoard = SetBitboard(passedPawnBoard, sq)
		}
	}

	return passedPawnBoard
}

func GetPawnRankBitboard(pawnBitboard uint64, rank byte) uint64 {
	var bitboard uint64
	for j := byte(0); j < 8; j++ {
		bitboard = SetBitboard(bitboard, idx(j, rank-1))
	}
	return pawnBitboard & bitboard
}

func GetIsolatedPawnBitboard(pawnBitboard uint64) uint64 {
	var bitboard uint64
	originalBoard := pawnBitboard
	for pawnBitboard != 0 {
		sq := byte(bits.TrailingZeros64(pawnBitboard))
		pawnBitboard ^= 1 << sq

		col := sq % 8
		var columnBoard uint64
		for j := byte(1); j < 7; j++ {
			maskSq := idx(col, j)
			if col > 0 {
				columnBoard = SetBitboard(columnBoard, maskSq-1)
			}
			if col < 7 {
				columnBoard = SetBitboard(columnBoard, maskSq+1)
			}
		}

		if columnBoard&originalBoard == 0 {
			bitboard = SetBitboard(bitboard, sq)
		}
	}

	return bitboard
}

// Return the pawn table entry for the given board state
func GetPawnTableEntry(boardState *BoardState) *PawnTableEntry {
	tableEntry := boardState.pawnTable[boardState.pawnHashKey]
	if tableEntry != nil {
		return tableEntry
	}

	// compute
	entry := PawnTableEntry{}
	allPawns := boardState.bitboards.piece[PAWN_MASK]
	whitePawns := allPawns & boardState.bitboards.color[WHITE_OFFSET]
	blackPawns := allPawns & boardState.bitboards.color[BLACK_OFFSET]
	entry.pawns[WHITE_OFFSET] = whitePawns
	entry.pawns[BLACK_OFFSET] = blackPawns
	entry.doubledPawns[WHITE_OFFSET] = GetDoubledPawnBitboard(whitePawns)
	entry.doubledPawns[BLACK_OFFSET] = GetDoubledPawnBitboard(blackPawns)
	entry.passedPawns[WHITE_OFFSET] = GetPassedPawnBitboard(whitePawns, blackPawns, WHITE_OFFSET)
	entry.passedPawns[BLACK_OFFSET] = GetPassedPawnBitboard(blackPawns, whitePawns, BLACK_OFFSET)
	for rank := RANK_2; rank <= RANK_7; rank++ {
		entry.pawnsPerRank[WHITE_OFFSET][rank] = GetPawnRankBitboard(whitePawns, rank)
		entry.pawnsPerRank[BLACK_OFFSET][rank] = GetPawnRankBitboard(blackPawns, rank)
	}
	entry.isolatedPawns[WHITE_OFFSET] = GetIsolatedPawnBitboard(whitePawns)
	entry.isolatedPawns[BLACK_OFFSET] = GetIsolatedPawnBitboard(blackPawns)
	entry.connectedPawns[WHITE_OFFSET] = whitePawns ^ entry.isolatedPawns[WHITE_OFFSET]
	entry.connectedPawns[BLACK_OFFSET] = blackPawns ^ entry.isolatedPawns[BLACK_OFFSET]
	boardState.pawnTable[boardState.pawnHashKey] = &entry

	return &entry
}
