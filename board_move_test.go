package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var _ = fmt.Println

func TestApplyMove(t *testing.T) {
	emptyBoard := CreateEmptyBoardState()
	emptyBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|PAWN_MASK)

	move := CreateMove(SQUARE_A2, SQUARE_A4)
	emptyBoard.ApplyMove(move)

	assert.Equal(t, EMPTY_SQUARE, emptyBoard.PieceAtSquare(SQUARE_A2))
	assert.Equal(t, WHITE_MASK|PAWN_MASK, emptyBoard.PieceAtSquare(SQUARE_A4))
	assert.False(t, emptyBoard.whiteToMove)
	assert.Equal(t, emptyBoard.boardInfo.enPassantTargetSquare, SQUARE_A3)

	emptyBoard.UnapplyMove(move)

	assert.Equal(t, WHITE_MASK|PAWN_MASK, emptyBoard.PieceAtSquare(SQUARE_A2))
	assert.Equal(t, EMPTY_SQUARE, emptyBoard.PieceAtSquare(SQUARE_A4))
	assert.True(t, emptyBoard.whiteToMove)
	assert.Equal(t, emptyBoard.boardInfo.enPassantTargetSquare, EMPTY_SQUARE)
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

	emptyBoard.ApplyMove(CreateMove(SQUARE_A7, SQUARE_A5))

	assert.Equal(t, EMPTY_SQUARE, emptyBoard.PieceAtSquare(SQUARE_A7))
	assert.Equal(t, BLACK_MASK|PAWN_MASK, emptyBoard.PieceAtSquare(SQUARE_A5))
	assert.True(t, emptyBoard.whiteToMove)
	assert.Equal(t, emptyBoard.boardInfo.enPassantTargetSquare, SQUARE_A6)

	emptyBoard.UnapplyMove(CreateMove(SQUARE_A7, SQUARE_A5))

	assert.Equal(t, BLACK_MASK|PAWN_MASK, emptyBoard.PieceAtSquare(SQUARE_A7))
	assert.Equal(t, EMPTY_SQUARE, emptyBoard.PieceAtSquare(SQUARE_A5))
	assert.False(t, emptyBoard.whiteToMove)
	assert.Equal(t, emptyBoard.boardInfo.enPassantTargetSquare, EMPTY_SQUARE)
}

func TestApplyCapture(t *testing.T) {
	testBoard := CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_B3, BLACK_MASK|ROOK_MASK)

	m := CreateCapture(SQUARE_A2, SQUARE_B3)
	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_A2))
	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_B3))

	testBoard.UnapplyMove(m)

	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_A2))
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_B3))
}

func TestApplyCaptureTwice(t *testing.T) {
	testBoard := CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_B3, BLACK_MASK|ROOK_MASK)
	testBoard.SetPieceAtSquare(SQUARE_C5, BLACK_MASK|KNIGHT_MASK)

	m1 := CreateCapture(SQUARE_A2, SQUARE_B3)
	m2 := CreateCapture(SQUARE_C5, SQUARE_B3)

	testBoard.ApplyMove(m1)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_A2))
	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_B3))
	assert.Equal(t, BLACK_MASK|KNIGHT_MASK, testBoard.PieceAtSquare(SQUARE_C5))

	testBoard.ApplyMove(m2)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_A2))
	assert.Equal(t, BLACK_MASK|KNIGHT_MASK, testBoard.PieceAtSquare(SQUARE_B3))
	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_C5))

	testBoard.UnapplyMove(m2)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_A2))
	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_B3))
	assert.Equal(t, BLACK_MASK|KNIGHT_MASK, testBoard.PieceAtSquare(SQUARE_C5))

	testBoard.UnapplyMove(m1)

	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_A2))
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_B3))
	assert.Equal(t, BLACK_MASK|KNIGHT_MASK, testBoard.PieceAtSquare(SQUARE_C5))
}

func TestApplyWhiteKingsideCastle(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_E1, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_H1, WHITE_MASK|ROOK_MASK)

	var m Move = CreateKingsideCastle(SQUARE_E1, SQUARE_G1)

	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_E1))
	assert.Equal(t, WHITE_MASK|KING_MASK, testBoard.PieceAtSquare(SQUARE_G1))
	assert.Equal(t, WHITE_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_F1))
	assert.False(t, testBoard.boardInfo.whiteCanCastleKingside)

	testBoard.UnapplyMove(m)

	assert.Equal(t, WHITE_MASK|KING_MASK, testBoard.PieceAtSquare(SQUARE_E1))
	assert.Equal(t, WHITE_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_H1))
	assert.True(t, testBoard.boardInfo.whiteCanCastleKingside)
}

func TestApplyBlackKingsideCastle(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_E8, BLACK_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_H8, BLACK_MASK|ROOK_MASK)
	testBoard.whiteToMove = false
	testBoard.boardInfo.blackCanCastleKingside = true

	var m Move = CreateKingsideCastle(SQUARE_E8, SQUARE_G8)

	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_E8))
	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_H8))
	assert.Equal(t, BLACK_MASK|KING_MASK, testBoard.PieceAtSquare(SQUARE_G8))
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_F8))
	assert.False(t, testBoard.boardInfo.blackCanCastleKingside)

	testBoard.UnapplyMove(m)

	assert.Equal(t, BLACK_MASK|KING_MASK, testBoard.PieceAtSquare(SQUARE_E8))
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_H8))
	assert.True(t, testBoard.boardInfo.blackCanCastleKingside)
}

func TestApplyWhiteQueensideCastle(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_E1, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A1, WHITE_MASK|ROOK_MASK)

	var m Move = CreateQueensideCastle(SQUARE_E1, SQUARE_C1)

	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_E1))
	assert.Equal(t, WHITE_MASK|KING_MASK, testBoard.PieceAtSquare(SQUARE_C1))
	assert.Equal(t, WHITE_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_D1))
	assert.False(t, testBoard.boardInfo.whiteCanCastleQueenside)

	testBoard.UnapplyMove(m)

	assert.Equal(t, WHITE_MASK|KING_MASK, testBoard.PieceAtSquare(SQUARE_E1))
	assert.Equal(t, WHITE_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_A1))
	assert.True(t, testBoard.boardInfo.whiteCanCastleQueenside)
}

func TestApplyBlackQueensideCastle(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_E8, BLACK_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A8, BLACK_MASK|ROOK_MASK)
	testBoard.whiteToMove = false
	testBoard.boardInfo.blackCanCastleQueenside = true

	var m Move = CreateQueensideCastle(SQUARE_E8, SQUARE_C8)

	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_E8))
	assert.Equal(t, BLACK_MASK|KING_MASK, testBoard.PieceAtSquare(SQUARE_C8))
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_D8))
	assert.False(t, testBoard.boardInfo.blackCanCastleQueenside)

	testBoard.UnapplyMove(m)

	assert.Equal(t, BLACK_MASK|KING_MASK, testBoard.PieceAtSquare(SQUARE_E8))
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_A8))
	assert.True(t, testBoard.boardInfo.blackCanCastleQueenside)
}

func TestWhitePawnPromotes(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_H7, WHITE_MASK|PAWN_MASK)

	var m Move = CreatePromotion(SQUARE_H7, SQUARE_H8, QUEEN_MASK)

	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_H7))
	assert.Equal(t, WHITE_MASK|QUEEN_MASK, testBoard.PieceAtSquare(SQUARE_H8))

	testBoard.UnapplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_H8))
	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_H7))
}

func TestBlackPawnPromotes(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_D2, BLACK_MASK|PAWN_MASK)
	testBoard.whiteToMove = false

	var m Move = CreatePromotion(SQUARE_D2, SQUARE_D1, ROOK_MASK)

	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_D2))
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.PieceAtSquare(SQUARE_D1))

	testBoard.UnapplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_D1))
	assert.Equal(t, BLACK_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_D2))
}

func TestEnPassantCapture(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A5, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_B5, BLACK_MASK|PAWN_MASK)
	testBoard.boardInfo.enPassantTargetSquare = SQUARE_B6

	var m Move = CreateEnPassantCapture(SQUARE_A5, SQUARE_B6)
	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_B5))
	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_B6))

	testBoard.UnapplyMove(m)

	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_A5))
	assert.Equal(t, BLACK_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_B5))
	assert.Equal(t, SQUARE_B6, testBoard.boardInfo.enPassantTargetSquare)
}

func TestEnPassantCaptureBlack(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_F4, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_E4, BLACK_MASK|PAWN_MASK)
	testBoard.boardInfo.enPassantTargetSquare = SQUARE_F3
	testBoard.whiteToMove = false

	var m Move = CreateEnPassantCapture(SQUARE_E4, SQUARE_F3)
	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.PieceAtSquare(SQUARE_F4))
	assert.Equal(t, BLACK_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_F3))

	testBoard.UnapplyMove(m)

	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_F4))
	assert.Equal(t, BLACK_MASK|PAWN_MASK, testBoard.PieceAtSquare(SQUARE_E4))
	assert.Equal(t, SQUARE_F3, testBoard.boardInfo.enPassantTargetSquare)
}

func TestFiddlingWithQueensideRooks(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()

	testBoard.SetPieceAtSquare(SQUARE_E8, BLACK_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A8, BLACK_MASK|ROOK_MASK)
	testBoard.SetPieceAtSquare(SQUARE_E1, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A1, WHITE_MASK|ROOK_MASK)
	testBoard.boardInfo.blackCanCastleKingside = true
	testBoard.boardInfo.blackCanCastleQueenside = true
	testBoard.boardInfo.whiteCanCastleKingside = true
	testBoard.boardInfo.whiteCanCastleQueenside = true

	var m1 Move = CreateMove(SQUARE_A1, SQUARE_A3)
	var m2 Move = CreateMove(SQUARE_A8, SQUARE_A6)
	var m3 Move = CreateMove(SQUARE_A3, SQUARE_A1)
	var m4 Move = CreateMove(SQUARE_A6, SQUARE_A8)

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
	var testBoard BoardState = CreateEmptyBoardState()

	testBoard.SetPieceAtSquare(SQUARE_E8, BLACK_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_H8, BLACK_MASK|ROOK_MASK)
	testBoard.SetPieceAtSquare(SQUARE_E1, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_H1, WHITE_MASK|ROOK_MASK)
	testBoard.boardInfo.blackCanCastleKingside = true
	testBoard.boardInfo.blackCanCastleQueenside = true
	testBoard.boardInfo.whiteCanCastleKingside = true
	testBoard.boardInfo.whiteCanCastleQueenside = true

	var m1 Move = CreateMove(SQUARE_H1, SQUARE_H3)
	var m2 Move = CreateMove(SQUARE_H8, SQUARE_H6)
	var m3 Move = CreateMove(SQUARE_H3, SQUARE_H1)
	var m4 Move = CreateMove(SQUARE_H6, SQUARE_H8)

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
