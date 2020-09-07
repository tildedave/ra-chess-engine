package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoubledPawnBitboard(t *testing.T) {
	entry := PawnTableEntry{}
	whitePawns := SetBitboardMultiple(0, SQUARE_D4, SQUARE_E5)
	computePawnStructure(&entry, whitePawns, 0, WHITE_OFFSET)
	assert.Equal(t, uint64(0), entry.doubledPawnBoard[WHITE_OFFSET])

	whitePawns = SetBitboardMultiple(0, SQUARE_D4, SQUARE_E4)
	computePawnStructure(&entry, whitePawns, 0, WHITE_OFFSET)
	assert.Equal(t, uint64(0), entry.doubledPawnBoard[WHITE_OFFSET])

	whitePawns = SetBitboardMultiple(0, SQUARE_D4, SQUARE_D3)
	computePawnStructure(&entry, whitePawns, 0, WHITE_OFFSET)
	assert.Equal(t, whitePawns, entry.doubledPawnBoard[WHITE_OFFSET])

	whitePawns = SetBitboardMultiple(0, SQUARE_A4, SQUARE_A3, SQUARE_A2)
	computePawnStructure(&entry, whitePawns, 0, WHITE_OFFSET)
	assert.Equal(t, whitePawns, entry.doubledPawnBoard[WHITE_OFFSET])
}

func TestOpenFiles(t *testing.T) {
	boardState := CreateEmptyBoardState()
	boardState.SetPieceAtSquare(SQUARE_D4, PAWN_MASK|WHITE_MASK)
	boardState.SetPieceAtSquare(SQUARE_E5, PAWN_MASK|WHITE_MASK)
	boardState.SetPieceAtSquare(SQUARE_D7, PAWN_MASK|BLACK_MASK)
	boardState.SetPieceAtSquare(SQUARE_A7, PAWN_MASK|BLACK_MASK)
	entry := GetPawnTableEntry(&boardState)

	assert.True(t, IsBitboardSet(entry.halfOpenFileBoard[WHITE_OFFSET], SQUARE_E5))
	assert.True(t, IsBitboardSet(entry.halfOpenFileBoard[BLACK_OFFSET], SQUARE_A7))

	assert.False(t, IsBitboardSet(entry.halfOpenFileBoard[WHITE_OFFSET], SQUARE_D4))
	assert.False(t, IsBitboardSet(entry.halfOpenFileBoard[BLACK_OFFSET], SQUARE_D7))

	assert.False(t, IsBitboardSet(entry.openFileBoard, SQUARE_A1))
	assert.True(t, IsBitboardSet(entry.openFileBoard, SQUARE_B1))
	assert.True(t, IsBitboardSet(entry.openFileBoard, SQUARE_C1))
	assert.False(t, IsBitboardSet(entry.openFileBoard, SQUARE_D1))
	assert.False(t, IsBitboardSet(entry.openFileBoard, SQUARE_E1))
	assert.True(t, IsBitboardSet(entry.openFileBoard, SQUARE_F1))
	assert.True(t, IsBitboardSet(entry.openFileBoard, SQUARE_G1))
	assert.True(t, IsBitboardSet(entry.openFileBoard, SQUARE_H1))
}

func TestPassedPawnBitboard(t *testing.T) {
	entry := PawnTableEntry{}
	bitboard := SetBitboard(0, SQUARE_D4)
	otherBitboard := SetBitboard(0, SQUARE_D7)

	computePawnStructure(&entry, bitboard, otherBitboard, WHITE_OFFSET)
	assert.Equal(t, uint64(0), entry.passedPawns[WHITE_OFFSET])

	computePawnStructure(&entry, otherBitboard, bitboard, BLACK_OFFSET)
	assert.Equal(t, uint64(0), entry.passedPawns[BLACK_OFFSET])

	computePawnStructure(&entry, bitboard, 0, WHITE_OFFSET)
	assert.Equal(t, bitboard, entry.passedPawns[WHITE_OFFSET])

	bitboard = SetBitboard(0, SQUARE_A2)
	otherBitboard = SetBitboard(0, SQUARE_H4)

	computePawnStructure(&entry, bitboard, otherBitboard, WHITE_OFFSET)
	computePawnStructure(&entry, otherBitboard, bitboard, BLACK_OFFSET)

	assert.Equal(t, bitboard, entry.passedPawns[WHITE_OFFSET])
	assert.Equal(t, otherBitboard, entry.passedPawns[BLACK_OFFSET])
}

func TestGetPawnRankBitboard(t *testing.T) {
	bitboard := SetBitboardMultiple(0, SQUARE_D4, SQUARE_H4)
	assert.Equal(t, bitboard, GetPawnRankBitboard(bitboard, RANK_4))
}

func TestGetIsolatedPawnBitboard(t *testing.T) {
	entry := PawnTableEntry{}
	bitboard := SetBitboardMultiple(0, SQUARE_H5, SQUARE_G4)

	computePawnStructure(&entry, bitboard, 0, WHITE_OFFSET)
	assert.Equal(t, uint64(0), entry.isolatedPawnBoard[WHITE_OFFSET])

	bitboard = SetBitboard(0, SQUARE_D4)
	computePawnStructure(&entry, bitboard, 0, WHITE_OFFSET)
	assert.Equal(t, bitboard, entry.isolatedPawnBoard[WHITE_OFFSET])

	// From wikipedia
	boardState, _ := CreateBoardStateFromFENString("8/8/8/PP2P2P/2P3P1/4P3/8/8 w - - 0 1")
	pawnBoard := boardState.bitboards.piece[PAWN_MASK]

	bitboard = SetBitboardMultiple(0, SQUARE_E3, SQUARE_E5)
	computePawnStructure(&entry, bitboard, 0, WHITE_OFFSET)

	assert.Equal(t, bitboard, entry.isolatedPawnBoard[WHITE_OFFSET])
	assert.Equal(t,
		SetBitboardMultiple(0, SQUARE_A5, SQUARE_B5, SQUARE_C4, SQUARE_G4, SQUARE_H5),
		pawnBoard^entry.isolatedPawnBoard[WHITE_OFFSET])
}

func TestGetPawnTableEntry(t *testing.T) {
	boardState := CreateEmptyBoardState()
	// White: a pawn and f pawns are passed
	// Black: D pawn is passed
	boardState.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|PAWN_MASK)
	boardState.SetPieceAtSquare(SQUARE_B2, WHITE_MASK|PAWN_MASK)
	boardState.SetPieceAtSquare(SQUARE_F5, WHITE_MASK|PAWN_MASK)
	boardState.SetPieceAtSquare(SQUARE_D6, BLACK_MASK|PAWN_MASK)
	boardState.SetPieceAtSquare(SQUARE_C7, BLACK_MASK|PAWN_MASK)

	entry := GetPawnTableEntry(&boardState)
	assert.Equal(t,
		SetBitboardMultiple(0, SQUARE_A2, SQUARE_F5),
		entry.passedPawns[WHITE_OFFSET])
	assert.Equal(t, SetBitboardMultiple(0, SQUARE_D6), entry.passedPawns[BLACK_OFFSET])

	assert.Equal(t,
		SetBitboardMultiple(0, SQUARE_A8, SQUARE_F8),
		entry.passedPawnQueeningSquares[WHITE_OFFSET])
	assert.Equal(t,
		SetBitboardMultiple(0, SQUARE_D1),
		entry.passedPawnQueeningSquares[BLACK_OFFSET])

	assert.Equal(t,
		SetBitboardMultiple(0, SQUARE_A3, SQUARE_A4, SQUARE_A5, SQUARE_A6, SQUARE_A7,
			SQUARE_A8, SQUARE_F6, SQUARE_F7, SQUARE_F8),
		entry.passedPawnAdvanceSquares[WHITE_OFFSET])
	assert.Equal(t,
		SetBitboardMultiple(0, SQUARE_D5, SQUARE_D4, SQUARE_D3, SQUARE_D2, SQUARE_D1),
		entry.passedPawnAdvanceSquares[BLACK_OFFSET])
}
