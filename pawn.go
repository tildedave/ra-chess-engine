package main

import (
	"math/bits"
)

type PawnTableEntry struct {
	pawns                     [2]uint64
	passedPawns               [2]uint64
	passedPawnAdvanceSquares  [2]uint64
	passedPawnQueeningSquares [2]uint64
	doubledPawnBoard          [2]uint64
	doubledPawnCount          [2]int
	isolatedPawnBoard         [2]uint64
	isolatedPawnCount         [2]int
	connectedPawnBoard        [2]uint64
	connectedPawnCount        [2]int
	pawnsPerRank              [2][8]uint64
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

// GetPawnTableEntry will return the pawn table entry for the given board state,
// creating it if it does not yet exist.
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
	entry.doubledPawnBoard[WHITE_OFFSET] = GetDoubledPawnBitboard(whitePawns)
	entry.doubledPawnBoard[BLACK_OFFSET] = GetDoubledPawnBitboard(blackPawns)
	whitePassers := GetPassedPawnBitboard(whitePawns, blackPawns, WHITE_OFFSET)
	blackPassers := GetPassedPawnBitboard(blackPawns, whitePawns, BLACK_OFFSET)

	entry.passedPawns[WHITE_OFFSET] = whitePassers
	entry.passedPawns[BLACK_OFFSET] = blackPassers

	var whiteQueeningSquares uint64
	var blackQueeningSquares uint64
	var whiteAdvanceSquares uint64
	var blackAdvanceSquares uint64
	for whitePassers != 0 {
		sq := byte(bits.TrailingZeros64(whitePassers))
		whitePassers ^= 1 << sq

		whiteQueeningSquares = SetBitboard(whiteQueeningSquares, 56+sq%8)
		row := sq / 8

		// We're using the magic bitboard FollowRay here which requires a "distance" which is
		// a bit vector of which square should have a bit set.  We just want them all set.
		distance := (1 << (8 - row + 1)) - 1
		whiteAdvanceSquares = FollowRay(whiteAdvanceSquares, sq%8, row, NORTH, distance)
	}
	for blackPassers != 0 {
		sq := byte(bits.TrailingZeros64(blackPassers))
		blackPassers ^= 1 << sq

		row := sq / 8
		distance := (1 << (row + 1)) - 1
		blackQueeningSquares = SetBitboard(blackQueeningSquares, sq%8)
		blackAdvanceSquares = FollowRay(blackAdvanceSquares, sq%8, row, SOUTH, distance)
	}
	entry.passedPawnAdvanceSquares[WHITE_OFFSET] = whiteAdvanceSquares
	entry.passedPawnAdvanceSquares[BLACK_OFFSET] = blackAdvanceSquares
	entry.passedPawnQueeningSquares[WHITE_OFFSET] = whiteQueeningSquares
	entry.passedPawnQueeningSquares[BLACK_OFFSET] = blackQueeningSquares

	for rank := RANK_2; rank <= RANK_7; rank++ {
		entry.pawnsPerRank[WHITE_OFFSET][rank] = GetPawnRankBitboard(whitePawns, rank)
		entry.pawnsPerRank[BLACK_OFFSET][rank] = GetPawnRankBitboard(blackPawns, rank)
	}
	entry.isolatedPawnBoard[WHITE_OFFSET] = GetIsolatedPawnBitboard(whitePawns)
	entry.isolatedPawnBoard[BLACK_OFFSET] = GetIsolatedPawnBitboard(blackPawns)
	entry.connectedPawnBoard[WHITE_OFFSET] = whitePawns ^ entry.isolatedPawnBoard[WHITE_OFFSET]
	entry.connectedPawnBoard[BLACK_OFFSET] = blackPawns ^ entry.isolatedPawnBoard[BLACK_OFFSET]
	boardState.pawnTable[boardState.pawnHashKey] = &entry
	for side := 0; side <= 1; side++ {
		entry.isolatedPawnCount[side] = bits.OnesCount64(entry.isolatedPawnBoard[side])
		entry.doubledPawnCount[side] = bits.OnesCount64(entry.doubledPawnBoard[side])
		entry.connectedPawnCount[side] = bits.OnesCount64(entry.connectedPawnBoard[side])
	}
	return &entry
}
