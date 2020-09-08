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
	attackBoard               [2]uint64
	doubledPawnCount          [2]int
	isolatedPawnBoard         [2]uint64
	isolatedPawnCount         [2]int
	connectedPawnBoard        [2]uint64
	connectedPawnCount        [2]int
	backwardsPawnBoard        [2]uint64
	pawnsPerRank              [2][8]uint64
	openFileBoard             uint64
	halfOpenFileBoard         [2]uint64
	pawnColumnBoard           [2]uint64
}

func GetPawnRankBitboard(pawnBitboard uint64, rank byte) uint64 {
	var bitboard uint64
	for j := byte(0); j < 8; j++ {
		bitboard = SetBitboard(bitboard, idx(j, rank-1))
	}
	return pawnBitboard & bitboard
}

func computePawnStructure(
	entry *PawnTableEntry,
	boardState *BoardState,
	pawnBitboard uint64,
	otherSidePawnBitboard uint64,
	side int,
) {
	originalBoard := pawnBitboard
	var passedPawnBoard uint64
	var isolatedPawnBoard uint64
	var doubledPawnBoard uint64
	var backwardsPawnBoard uint64
	var pawnColumnBoard uint64
	var attackBoard uint64

	for pawnBitboard != 0 {
		sq := byte(bits.TrailingZeros64(pawnBitboard))
		pawnBitboard ^= 1 << sq

		col := sq % 8
		pawnRank := Rank(sq)

		var adjacentBoard uint64
		var columnBoard uint64
		var inc int
		var start int
		if side == WHITE_OFFSET {
			inc = 1
			start = 0
		} else {
			inc = -1
			start = 7
		}
		for j := start; j < 8 && j >= 0; j += inc {
			maskSq := idx(col, byte(j))
			columnBoard = SetBitboard(columnBoard, maskSq)
			if col > 0 {
				adjacentBoard = SetBitboard(adjacentBoard, maskSq-1)
			}
			if col < 7 {
				adjacentBoard = SetBitboard(adjacentBoard, maskSq+1)
			}
		}
		if otherSidePawnBitboard&(columnBoard|adjacentBoard) == 0 {
			passedPawnBoard = SetBitboard(passedPawnBoard, sq)
		}
		supportingBoard := adjacentBoard & originalBoard
		if supportingBoard == 0 {
			isolatedPawnBoard = SetBitboard(isolatedPawnBoard, sq)
		} else {
			// determine if backwards
			isBackwards := true
			for supportingBoard != 0 {
				otherPawnSq := byte(bits.TrailingZeros64(supportingBoard))
				rankOtherPawn := Rank(otherPawnSq)
				if side == WHITE_OFFSET {
					if rankOtherPawn < pawnRank {
						isBackwards = false
						break
					}
				} else {
					if side == BLACK_OFFSET && pawnRank < rankOtherPawn {
						isBackwards = false
						break
					}
				}
				supportingBoard ^= 1 << otherPawnSq
			}
			if isBackwards {
				backwardsPawnBoard = SetBitboard(backwardsPawnBoard, sq)
			}
		}

		attackBoard |= boardState.moveBitboards.pawnAttacks[side][sq]
		doubledPawns := columnBoard & (originalBoard ^ (1 << sq))
		doubledPawnBoard |= doubledPawns
		pawnColumnBoard |= columnBoard
	}

	entry.attackBoard[side] = attackBoard
	entry.passedPawns[side] = passedPawnBoard
	entry.isolatedPawnBoard[side] = isolatedPawnBoard
	entry.doubledPawnBoard[side] = doubledPawnBoard
	entry.pawnColumnBoard[side] = pawnColumnBoard
	entry.backwardsPawnBoard[side] = backwardsPawnBoard
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

	computePawnStructure(&entry, boardState, whitePawns, blackPawns, WHITE_OFFSET)
	computePawnStructure(&entry, boardState, blackPawns, whitePawns, BLACK_OFFSET)

	whitePassers := entry.passedPawns[WHITE_OFFSET]
	blackPassers := entry.passedPawns[BLACK_OFFSET]

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
	entry.connectedPawnBoard[WHITE_OFFSET] = whitePawns ^ entry.isolatedPawnBoard[WHITE_OFFSET]
	entry.connectedPawnBoard[BLACK_OFFSET] = blackPawns ^ entry.isolatedPawnBoard[BLACK_OFFSET]
	boardState.pawnTable[boardState.pawnHashKey] = &entry
	for side := 0; side <= 1; side++ {
		entry.isolatedPawnCount[side] = bits.OnesCount64(entry.isolatedPawnBoard[side])
		entry.doubledPawnCount[side] = bits.OnesCount64(entry.doubledPawnBoard[side])
		entry.connectedPawnCount[side] = bits.OnesCount64(entry.connectedPawnBoard[side])
	}

	entry.halfOpenFileBoard[WHITE_OFFSET] = (entry.pawnColumnBoard[WHITE_OFFSET] &
		(entry.pawnColumnBoard[WHITE_OFFSET] ^ entry.pawnColumnBoard[BLACK_OFFSET]))
	entry.halfOpenFileBoard[BLACK_OFFSET] = (entry.pawnColumnBoard[BLACK_OFFSET] &
		(entry.pawnColumnBoard[BLACK_OFFSET] ^ entry.pawnColumnBoard[WHITE_OFFSET]))
	entry.openFileBoard = 0xFFFFFFFFFFFFFFFF ^ (entry.pawnColumnBoard[WHITE_OFFSET] | entry.pawnColumnBoard[BLACK_OFFSET])

	return &entry
}
