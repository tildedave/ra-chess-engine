package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _ = fmt.Println

func TestApplyMove(t *testing.T) {
	emptyBoard := CreateEmptyBoardState()
	emptyBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|PAWN_MASK)

	// bitboard asserts
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A2), emptyBoard.bitboards.color[0])
	assert.Equal(t, uint64(0), emptyBoard.bitboards.color[1])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A2), emptyBoard.bitboards.piece[1])

	move := CreateMove(SQUARE_A2, SQUARE_A4)
	emptyBoard.ApplyMove(move)

	assert.Equal(t, EMPTY_SQUARE, emptyBoard.PieceAtSquare(SQUARE_A2))
	assert.Equal(t, WHITE_MASK|PAWN_MASK, emptyBoard.PieceAtSquare(SQUARE_A4))
	assert.False(t, emptyBoard.whiteToMove)
	assert.Equal(t, emptyBoard.boardInfo.enPassantTargetSquare, SQUARE_A3)

	// bitboard asserts
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A4), emptyBoard.bitboards.color[0])
	assert.Equal(t, uint64(0), emptyBoard.bitboards.color[1])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A4), emptyBoard.bitboards.piece[1])

	emptyBoard.UnapplyMove(move)

	assert.Equal(t, WHITE_MASK|PAWN_MASK, emptyBoard.PieceAtSquare(SQUARE_A2))
	assert.Equal(t, EMPTY_SQUARE, emptyBoard.PieceAtSquare(SQUARE_A4))
	assert.True(t, emptyBoard.whiteToMove)
	assert.Equal(t, emptyBoard.boardInfo.enPassantTargetSquare, EMPTY_SQUARE)

	// bitboard asserts
	assert.Equal(t, uint64(0x100), emptyBoard.bitboards.color[0])
	assert.Equal(t, uint64(0), emptyBoard.bitboards.color[1])
	assert.Equal(t, uint64(0x100), emptyBoard.bitboards.piece[1])
}

func TestApplyBlackPawnSingleMove(t *testing.T) {
	emptyBoard := CreateEmptyBoardState()
	emptyBoard.SetPieceAtSquare(SQUARE_A7, BLACK_MASK|PAWN_MASK)
	emptyBoard.SetPieceAtSquare(SQUARE_B6, WHITE_MASK|ROOK_MASK)
	emptyBoard.whiteToMove = false

	emptyBoard.ApplyMove(CreateCapture(SQUARE_A7, SQUARE_B6))

	assert.Equal(t, emptyBoard.boardInfo.enPassantTargetSquare, uint8(0))
}

func TestApplyBlackPawnMove(t *testing.T) {
	emptyBoard := CreateEmptyBoardState()
	emptyBoard.SetPieceAtSquare(SQUARE_A7, BLACK_MASK|PAWN_MASK)
	emptyBoard.whiteToMove = false

	// bitboard asserts
	assert.Equal(t, uint64(0), emptyBoard.bitboards.color[0])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A7), emptyBoard.bitboards.color[1])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A7), emptyBoard.bitboards.piece[1])

	emptyBoard.ApplyMove(CreateMove(SQUARE_A7, SQUARE_A5))

	assert.Equal(t, EMPTY_SQUARE, emptyBoard.PieceAtSquare(SQUARE_A7))
	assert.Equal(t, BLACK_MASK|PAWN_MASK, emptyBoard.PieceAtSquare(SQUARE_A5))
	assert.True(t, emptyBoard.whiteToMove)
	assert.Equal(t, emptyBoard.boardInfo.enPassantTargetSquare, SQUARE_A6)

	// bitboard asserts
	assert.Equal(t, uint64(0), emptyBoard.bitboards.color[0])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A5), emptyBoard.bitboards.color[1])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A5), emptyBoard.bitboards.piece[1])

	emptyBoard.UnapplyMove(CreateMove(SQUARE_A7, SQUARE_A5))

	assert.Equal(t, BLACK_MASK|PAWN_MASK, emptyBoard.PieceAtSquare(SQUARE_A7))
	assert.Equal(t, EMPTY_SQUARE, emptyBoard.PieceAtSquare(SQUARE_A5))
	assert.False(t, emptyBoard.whiteToMove)
	assert.Equal(t, emptyBoard.boardInfo.enPassantTargetSquare, EMPTY_SQUARE)

	// bitboard asserts
	assert.Equal(t, uint64(0), emptyBoard.bitboards.color[0])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A7), emptyBoard.bitboards.color[1])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A7), emptyBoard.bitboards.piece[1])
}

func TestApplyCapture(t *testing.T) {
	testBoard := CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_B3, BLACK_MASK|ROOK_MASK)
	originalKey := testBoard.hashKey

	// bitboard asserts
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A2), testBoard.bitboards.color[0])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_B3), testBoard.bitboards.color[1])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A2), testBoard.bitboards.piece[BITBOARD_PAWN_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_B3), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])

	m := CreateCapture(SQUARE_A2, SQUARE_B3)
	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_A2))
	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_B3))

	// bitboard asserts
	assert.Equal(t, SetBitboard(0, BB_SQUARE_B3), testBoard.bitboards.color[0])
	assert.Equal(t, uint64(0), testBoard.bitboards.color[1])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_B3), testBoard.bitboards.piece[BITBOARD_PAWN_OFFSET])
	assert.Equal(t, uint64(0), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])

	testBoard.UnapplyMove(m)

	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_A2))
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_B3))
	assert.Equal(t, originalKey, testBoard.hashKey)

	// bitboard asserts
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A2), testBoard.bitboards.color[0])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_B3), testBoard.bitboards.color[1])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A2), testBoard.bitboards.piece[BITBOARD_PAWN_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_B3), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])
}

func TestApplyCaptureTwice(t *testing.T) {
	testBoard := CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_B3, BLACK_MASK|ROOK_MASK)
	testBoard.SetPieceAtSquare(SQUARE_C5, BLACK_MASK|KNIGHT_MASK)

	// bitboard asserts
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A2), testBoard.bitboards.color[0])
	assert.Equal(t, SetBitboard(SetBitboard(0, BB_SQUARE_B3), BB_SQUARE_C5), testBoard.bitboards.color[1])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A2), testBoard.bitboards.piece[BITBOARD_PAWN_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_B3), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_C5), testBoard.bitboards.piece[BITBOARD_KNIGHT_OFFSET])

	m1 := CreateCapture(SQUARE_A2, SQUARE_B3)
	m2 := CreateCapture(SQUARE_C5, SQUARE_B3)

	testBoard.ApplyMove(m1)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_A2))
	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_B3))
	assert.Equal(t, BLACK_MASK|KNIGHT_MASK, testBoard.PieceAtSquare(SQUARE_C5))

	// bitboard asserts
	assert.Equal(t, SetBitboard(0, BB_SQUARE_B3), testBoard.bitboards.color[0])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_C5), testBoard.bitboards.color[1])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_B3), testBoard.bitboards.piece[BITBOARD_PAWN_OFFSET])
	assert.Equal(t, uint64(0), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_C5), testBoard.bitboards.piece[BITBOARD_KNIGHT_OFFSET])

	testBoard.ApplyMove(m2)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_A2))
	assert.Equal(t, BLACK_MASK|KNIGHT_MASK, testBoard.PieceAtSquare(SQUARE_B3))
	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_C5))

	// bitboard asserts
	assert.Equal(t, uint64(0), testBoard.bitboards.color[0])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_B3), testBoard.bitboards.color[1])
	assert.Equal(t, uint64(0), testBoard.bitboards.piece[BITBOARD_PAWN_OFFSET])
	assert.Equal(t, uint64(0), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_B3), testBoard.bitboards.piece[BITBOARD_KNIGHT_OFFSET])

	testBoard.UnapplyMove(m2)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_A2))
	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_B3))
	assert.Equal(t, BLACK_MASK|KNIGHT_MASK, testBoard.PieceAtSquare(SQUARE_C5))

	// bitboard asserts
	assert.Equal(t, SetBitboard(0, BB_SQUARE_B3), testBoard.bitboards.color[0])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_C5), testBoard.bitboards.color[1])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_B3), testBoard.bitboards.piece[BITBOARD_PAWN_OFFSET])
	assert.Equal(t, uint64(0), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_C5), testBoard.bitboards.piece[BITBOARD_KNIGHT_OFFSET])

	testBoard.UnapplyMove(m1)

	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_A2))
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_B3))
	assert.Equal(t, BLACK_MASK|KNIGHT_MASK, testBoard.PieceAtSquare(SQUARE_C5))

	// bitboard asserts
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A2), testBoard.bitboards.color[0])
	assert.Equal(t, SetBitboard(SetBitboard(0, BB_SQUARE_B3), BB_SQUARE_C5), testBoard.bitboards.color[1])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A2), testBoard.bitboards.piece[BITBOARD_PAWN_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_B3), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_C5), testBoard.bitboards.piece[BITBOARD_KNIGHT_OFFSET])
}

func TestApplyWhiteKingsideCastle(t *testing.T) {
	var testBoard = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_E1, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_H1, WHITE_MASK|ROOK_MASK)
	originalKey := testBoard.hashKey

	// bitboard asserts
	assert.Equal(t, SetBitboard(SetBitboard(0, BB_SQUARE_E1), BB_SQUARE_H1), testBoard.bitboards.color[0])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_E1), testBoard.bitboards.piece[BITBOARD_KING_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_H1), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])

	testBoard.boardInfo.whiteCanCastleKingside = true
	var m = CreateKingsideCastle(SQUARE_E1, SQUARE_G1)

	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_E1))
	assert.Equal(t, WHITE_MASK|KING_MASK, testBoard.PieceAtSquare(SQUARE_G1))
	assert.Equal(t, WHITE_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_F1))
	assert.False(t, testBoard.boardInfo.whiteCanCastleKingside)

	// bitboard asserts
	assert.Equal(t, SetBitboard(SetBitboard(0, BB_SQUARE_G1), BB_SQUARE_F1), testBoard.bitboards.color[WHITE_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_G1), testBoard.bitboards.piece[BITBOARD_KING_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_F1), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])

	testBoard.UnapplyMove(m)

	assert.Equal(t, WHITE_MASK|KING_MASK, testBoard.PieceAtSquare(SQUARE_E1))
	assert.Equal(t, WHITE_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_H1))
	assert.True(t, testBoard.boardInfo.whiteCanCastleKingside)
	assert.Equal(t, testBoard.hashKey, originalKey)

	// bitboard asserts
	assert.Equal(t, SetBitboard(SetBitboard(0, BB_SQUARE_E1), BB_SQUARE_H1), testBoard.bitboards.color[0])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_E1), testBoard.bitboards.piece[BITBOARD_KING_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_H1), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])
}

func TestApplyBlackKingsideCastle(t *testing.T) {
	var testBoard = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_E8, BLACK_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_H8, BLACK_MASK|ROOK_MASK)
	testBoard.whiteToMove = false
	testBoard.boardInfo.blackCanCastleKingside = true

	// bitboard asserts
	assert.Equal(t, SetBitboard(SetBitboard(0, BB_SQUARE_E8), BB_SQUARE_H8), testBoard.bitboards.color[BLACK_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_E8), testBoard.bitboards.piece[BITBOARD_KING_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_H8), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])

	var m = CreateKingsideCastle(SQUARE_E8, SQUARE_G8)

	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_E8))
	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_H8))
	assert.Equal(t, BLACK_MASK|KING_MASK, testBoard.PieceAtSquare(SQUARE_G8))
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_F8))
	assert.False(t, testBoard.boardInfo.blackCanCastleKingside)

	// bitboard asserts
	assert.Equal(t, SetBitboard(SetBitboard(0, BB_SQUARE_G8), BB_SQUARE_F8), testBoard.bitboards.color[BLACK_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_G8), testBoard.bitboards.piece[BITBOARD_KING_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_F8), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])

	testBoard.UnapplyMove(m)

	assert.Equal(t, BLACK_MASK|KING_MASK, testBoard.PieceAtSquare(SQUARE_E8))
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_H8))
	assert.True(t, testBoard.boardInfo.blackCanCastleKingside)

	// bitboard asserts
	assert.Equal(t, SetBitboard(SetBitboard(0, BB_SQUARE_E8), BB_SQUARE_H8), testBoard.bitboards.color[BLACK_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_E8), testBoard.bitboards.piece[BITBOARD_KING_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_H8), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])
}

func TestApplyWhiteQueensideCastle(t *testing.T) {
	var testBoard = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_E1, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A1, WHITE_MASK|ROOK_MASK)

	// bitboard asserts
	assert.Equal(t, SetBitboard(SetBitboard(0, BB_SQUARE_E1), BB_SQUARE_A1), testBoard.bitboards.color[WHITE_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_E1), testBoard.bitboards.piece[BITBOARD_KING_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A1), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])

	var m = CreateQueensideCastle(SQUARE_E1, SQUARE_C1)

	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_E1))
	assert.Equal(t, WHITE_MASK|KING_MASK, testBoard.PieceAtSquare(SQUARE_C1))
	assert.Equal(t, WHITE_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_D1))
	assert.False(t, testBoard.boardInfo.whiteCanCastleQueenside)

	// bitboard asserts
	assert.Equal(t, SetBitboard(SetBitboard(0, BB_SQUARE_C1), BB_SQUARE_D1), testBoard.bitboards.color[WHITE_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_C1), testBoard.bitboards.piece[BITBOARD_KING_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_D1), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])

	testBoard.UnapplyMove(m)

	assert.Equal(t, WHITE_MASK|KING_MASK, testBoard.PieceAtSquare(SQUARE_E1))
	assert.Equal(t, WHITE_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_A1))
	assert.True(t, testBoard.boardInfo.whiteCanCastleQueenside)

	// bitboard asserts
	assert.Equal(t, SetBitboard(SetBitboard(0, BB_SQUARE_E1), BB_SQUARE_A1), testBoard.bitboards.color[WHITE_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_E1), testBoard.bitboards.piece[BITBOARD_KING_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A1), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])
}

func TestApplyBlackQueensideCastle(t *testing.T) {
	var testBoard = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_E8, BLACK_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A8, BLACK_MASK|ROOK_MASK)
	testBoard.whiteToMove = false
	testBoard.boardInfo.blackCanCastleQueenside = true

	// bitboard asserts
	assert.Equal(t, SetBitboard(SetBitboard(0, BB_SQUARE_E8), BB_SQUARE_A8), testBoard.bitboards.color[BLACK_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_E8), testBoard.bitboards.piece[BITBOARD_KING_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A8), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])

	var m = CreateQueensideCastle(SQUARE_E8, SQUARE_C8)

	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_E8))
	assert.Equal(t, BLACK_MASK|KING_MASK, testBoard.PieceAtSquare(SQUARE_C8))
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_D8))
	assert.False(t, testBoard.boardInfo.blackCanCastleQueenside)

	// bitboard asserts
	assert.Equal(t, SetBitboard(SetBitboard(0, BB_SQUARE_C8), BB_SQUARE_D8), testBoard.bitboards.color[BLACK_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_C8), testBoard.bitboards.piece[BITBOARD_KING_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_D8), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])

	testBoard.UnapplyMove(m)

	assert.Equal(t, BLACK_MASK|KING_MASK, testBoard.PieceAtSquare(SQUARE_E8))
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_A8))
	assert.True(t, testBoard.boardInfo.blackCanCastleQueenside)

	// bitboard asserts
	assert.Equal(t, SetBitboard(SetBitboard(0, BB_SQUARE_E8), BB_SQUARE_A8), testBoard.bitboards.color[BLACK_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_E8), testBoard.bitboards.piece[BITBOARD_KING_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_A8), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])
}

func TestWhitePawnPromotes(t *testing.T) {
	var testBoard = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_H7, WHITE_MASK|PAWN_MASK)
	originalKey := testBoard.hashKey

	// bitboard asserts
	assert.Equal(t, SetBitboard(0, BB_SQUARE_H7), testBoard.bitboards.color[WHITE_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_H7), testBoard.bitboards.piece[BITBOARD_PAWN_OFFSET])
	assert.Equal(t, uint64(0), testBoard.bitboards.piece[BITBOARD_QUEEN_OFFSET])

	var m = CreatePromotion(SQUARE_H7, SQUARE_H8, QUEEN_MASK)

	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_H7))
	assert.Equal(t, WHITE_MASK|QUEEN_MASK, testBoard.PieceAtSquare(SQUARE_H8))
	// bitboard asserts
	assert.Equal(t, SetBitboard(0, BB_SQUARE_H8), testBoard.bitboards.color[WHITE_OFFSET])
	assert.Equal(t, uint64(0), testBoard.bitboards.piece[BITBOARD_PAWN_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_H8), testBoard.bitboards.piece[BITBOARD_QUEEN_OFFSET])

	testBoard.UnapplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_H8))
	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_H7))
	assert.Equal(t, originalKey, testBoard.hashKey)

	// bitboard asserts
	assert.Equal(t, SetBitboard(0, BB_SQUARE_H7), testBoard.bitboards.color[WHITE_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_H7), testBoard.bitboards.piece[BITBOARD_PAWN_OFFSET])
	assert.Equal(t, uint64(0), testBoard.bitboards.piece[BITBOARD_QUEEN_OFFSET])
}

func TestBlackPawnPromotes(t *testing.T) {
	var testBoard = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_D2, BLACK_MASK|PAWN_MASK)
	testBoard.whiteToMove = false
	originalKey := testBoard.hashKey

	// bitboard asserts
	assert.Equal(t, SetBitboard(0, BB_SQUARE_D2), testBoard.bitboards.color[BLACK_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_D2), testBoard.bitboards.piece[BITBOARD_PAWN_OFFSET])
	assert.Equal(t, uint64(0), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])

	var m = CreatePromotion(SQUARE_D2, SQUARE_D1, ROOK_MASK)

	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_D2))
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_D1))

	// bitboard asserts
	assert.Equal(t, SetBitboard(0, BB_SQUARE_D1), testBoard.bitboards.color[BLACK_OFFSET])
	assert.Equal(t, uint64(0), testBoard.bitboards.piece[BITBOARD_PAWN_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_D1), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])

	testBoard.UnapplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_D1))
	assert.Equal(t, BLACK_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_D2))
	assert.Equal(t, originalKey, testBoard.hashKey)

	// bitboard asserts
	assert.Equal(t, SetBitboard(0, BB_SQUARE_D2), testBoard.bitboards.color[BLACK_OFFSET])
	assert.Equal(t, SetBitboard(0, BB_SQUARE_D2), testBoard.bitboards.piece[BITBOARD_PAWN_OFFSET])
	assert.Equal(t, uint64(0), testBoard.bitboards.piece[BITBOARD_ROOK_OFFSET])
}

func TestEnPassantCapture(t *testing.T) {
	var testBoard = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A5, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_B5, BLACK_MASK|PAWN_MASK)
	testBoard.boardInfo.enPassantTargetSquare = SQUARE_B6
	generateZobrishHashInfo(&testBoard)
	originalKey := testBoard.hashKey

	var m = CreateEnPassantCapture(SQUARE_A5, SQUARE_B6)
	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_B5))
	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_B6))

	testBoard.UnapplyMove(m)

	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_A5))
	assert.Equal(t, BLACK_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_B5))
	assert.Equal(t, SQUARE_B6, testBoard.boardInfo.enPassantTargetSquare)
	assert.Equal(t, originalKey, testBoard.hashKey)
}

func TestEnPassantCaptureBlack(t *testing.T) {
	var testBoard = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_F4, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_E4, BLACK_MASK|PAWN_MASK)
	testBoard.boardInfo.enPassantTargetSquare = SQUARE_F3
	testBoard.whiteToMove = false
	generateZobrishHashInfo(&testBoard)
	originalKey := testBoard.hashKey

	var m = CreateEnPassantCapture(SQUARE_E4, SQUARE_F3)
	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_F4))
	assert.Equal(t, BLACK_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_F3))

	testBoard.UnapplyMove(m)

	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_F4))
	assert.Equal(t, BLACK_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_E4))
	assert.Equal(t, SQUARE_F3, testBoard.boardInfo.enPassantTargetSquare)
	assert.Equal(t, originalKey, testBoard.hashKey)
}

func TestFiddlingWithQueensideRooks(t *testing.T) {
	var testBoard = CreateEmptyBoardState()

	testBoard.SetPieceAtSquare(SQUARE_E8, BLACK_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A8, BLACK_MASK|ROOK_MASK)
	testBoard.SetPieceAtSquare(SQUARE_E1, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A1, WHITE_MASK|ROOK_MASK)
	testBoard.boardInfo.blackCanCastleKingside = true
	testBoard.boardInfo.blackCanCastleQueenside = true
	testBoard.boardInfo.whiteCanCastleKingside = true
	testBoard.boardInfo.whiteCanCastleQueenside = true

	var m1 = CreateMove(SQUARE_A1, SQUARE_A3)
	var m2 = CreateMove(SQUARE_A8, SQUARE_A6)
	var m3 = CreateMove(SQUARE_A3, SQUARE_A1)
	var m4 = CreateMove(SQUARE_A6, SQUARE_A8)

	testBoard.ApplyMove(m1)
	assert.False(t, testBoard.boardInfo.whiteCanCastleQueenside)
	testBoard.ApplyMove(m2)
	assert.False(t, testBoard.boardInfo.blackCanCastleQueenside)
	testBoard.ApplyMove(m3)
	assert.False(t, testBoard.boardInfo.whiteCanCastleQueenside)
	testBoard.ApplyMove(m4)
	assert.False(t, testBoard.boardInfo.blackCanCastleQueenside)

	testBoard.UnapplyMove(m4)
	assert.False(t, testBoard.boardInfo.blackCanCastleQueenside)
	testBoard.UnapplyMove(m3)
	assert.False(t, testBoard.boardInfo.whiteCanCastleQueenside)
	testBoard.UnapplyMove(m2)
	assert.True(t, testBoard.boardInfo.blackCanCastleQueenside)
	testBoard.UnapplyMove(m1)
	assert.True(t, testBoard.boardInfo.whiteCanCastleQueenside)
}

func TestFiddlingWithKingsideRooks(t *testing.T) {
	var testBoard = CreateEmptyBoardState()

	testBoard.SetPieceAtSquare(SQUARE_E8, BLACK_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_H8, BLACK_MASK|ROOK_MASK)
	testBoard.SetPieceAtSquare(SQUARE_E1, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_H1, WHITE_MASK|ROOK_MASK)
	testBoard.boardInfo.blackCanCastleKingside = true
	testBoard.boardInfo.blackCanCastleQueenside = true
	testBoard.boardInfo.whiteCanCastleKingside = true
	testBoard.boardInfo.whiteCanCastleQueenside = true

	var m1 = CreateMove(SQUARE_H1, SQUARE_H3)
	var m2 = CreateMove(SQUARE_H8, SQUARE_H6)
	var m3 = CreateMove(SQUARE_H3, SQUARE_H1)
	var m4 = CreateMove(SQUARE_H6, SQUARE_H8)

	testBoard.ApplyMove(m1)
	assert.False(t, testBoard.boardInfo.whiteCanCastleKingside)
	testBoard.ApplyMove(m2)
	assert.False(t, testBoard.boardInfo.blackCanCastleKingside)
	testBoard.ApplyMove(m3)
	assert.False(t, testBoard.boardInfo.whiteCanCastleKingside)
	testBoard.ApplyMove(m4)
	assert.False(t, testBoard.boardInfo.blackCanCastleKingside)

	testBoard.UnapplyMove(m4)
	assert.False(t, testBoard.boardInfo.blackCanCastleKingside)
	testBoard.UnapplyMove(m3)
	assert.False(t, testBoard.boardInfo.whiteCanCastleKingside)
	testBoard.UnapplyMove(m2)
	assert.True(t, testBoard.boardInfo.blackCanCastleKingside)
	testBoard.UnapplyMove(m1)
	assert.True(t, testBoard.boardInfo.whiteCanCastleKingside)
}
