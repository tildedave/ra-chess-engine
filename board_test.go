package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var _ = fmt.Println

func TestInitialBoard(t *testing.T) {
	var initialBoard BoardState = CreateInitialBoardState()

	assert.Equal(t, initialBoard.PieceAtSquare(RowAndColToSquare(0, 0)), byte(0x04))
	assert.Equal(t, initialBoard.PieceAtSquare(RowAndColToSquare(0, 1)), byte(0x02))

	for i := 0; i < 8; i++ {
		piece := initialBoard.PieceAtSquare(RowAndColToSquare(1, byte(i)))
		assert.Equal(t, piece, byte(0x01))
		assert.True(t, isPieceWhite(piece))
		assert.True(t, isPawn(piece))
	}
	for i := 2; i <= 5; i++ {
		for j := 0; j < 8; j++ {
			piece := initialBoard.PieceAtSquare(RowAndColToSquare(byte(i), byte(j)))
			assert.Equal(t, piece, byte(0x00))
			assert.True(t, isSquareEmpty(piece))
		}
	}
	for i := 0; i < 8; i++ {
		piece := initialBoard.PieceAtSquare(RowAndColToSquare(6, byte(i)))
		assert.Equal(t, piece, byte(0x81))
		assert.True(t, isPieceBlack(piece))
		assert.True(t, isPawn(piece))
	}
}

func TestToString(t *testing.T) {
	var initialBoard BoardState = CreateInitialBoardState()

	var str = initialBoard.ToString()
	assert.Equal(t, "rnbqkbnr\npppppppp\n........\n........\n........\n........\nPPPPPPPP\nRNBQKBNR\n", str)
}

func TestToFEN(t *testing.T) {
	var initialBoard BoardState = CreateInitialBoardState()
	assert.Equal(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", initialBoard.ToFENString())

	var emptyBoard BoardState = CreateEmptyBoardState()
	assert.Equal(t, "8/8/8/8/8/8/8/8 w - - 0 1", emptyBoard.ToFENString())
}

func TestSquareToAlgebraicString(t *testing.T) {
	assert.Equal(t, "??", SquareToAlgebraicString(0))
	assert.Equal(t, "??", SquareToAlgebraicString(10))
	assert.Equal(t, "??", SquareToAlgebraicString(20))
	assert.Equal(t, "a1", SquareToAlgebraicString(SQUARE_A1))
	assert.Equal(t, "b1", SquareToAlgebraicString(SQUARE_B1))
	assert.Equal(t, "c1", SquareToAlgebraicString(SQUARE_C1))
	assert.Equal(t, "d1", SquareToAlgebraicString(SQUARE_D1))
	assert.Equal(t, "e1", SquareToAlgebraicString(SQUARE_E1))
	assert.Equal(t, "f1", SquareToAlgebraicString(SQUARE_F1))
	assert.Equal(t, "g1", SquareToAlgebraicString(SQUARE_G1))
	assert.Equal(t, "h1", SquareToAlgebraicString(SQUARE_H1))
	assert.Equal(t, "??", SquareToAlgebraicString(29))
	assert.Equal(t, "??", SquareToAlgebraicString(30))
	assert.Equal(t, "a2", SquareToAlgebraicString(SQUARE_A2))
	assert.Equal(t, "a3", SquareToAlgebraicString(SQUARE_A3))
	assert.Equal(t, "a4", SquareToAlgebraicString(SQUARE_A4))
	assert.Equal(t, "a5", SquareToAlgebraicString(SQUARE_A5))
	assert.Equal(t, "a6", SquareToAlgebraicString(SQUARE_A6))
	assert.Equal(t, "a7", SquareToAlgebraicString(SQUARE_A7))
	assert.Equal(t, "a8", SquareToAlgebraicString(SQUARE_A8))
}

func TestApplyMove(t *testing.T) {
	var emptyBoard BoardState = CreateEmptyBoardState()
	emptyBoard.board[SQUARE_A2] = WHITE_MASK | PAWN_MASK

	emptyBoard.ApplyMove(CreateMove(SQUARE_A2, SQUARE_A4))

	assert.Equal(t, EMPTY_SQUARE, emptyBoard.board[SQUARE_A2])
	assert.Equal(t, WHITE_MASK|PAWN_MASK, emptyBoard.board[SQUARE_A4])
	assert.False(t, emptyBoard.whiteToMove)
	assert.Equal(t, emptyBoard.boardInfo.enPassantTargetSquare, SQUARE_A3)

	emptyBoard.UnapplyMove(CreateMove(SQUARE_A2, SQUARE_A4))

	assert.Equal(t, WHITE_MASK|PAWN_MASK, emptyBoard.board[SQUARE_A2])
	assert.Equal(t, EMPTY_SQUARE, emptyBoard.board[SQUARE_A4])
	assert.True(t, emptyBoard.whiteToMove)
	assert.Equal(t, emptyBoard.boardInfo.enPassantTargetSquare, EMPTY_SQUARE)
}

func TestApplyBlackPawnMove(t *testing.T) {
	var emptyBoard BoardState = CreateEmptyBoardState()
	emptyBoard.board[SQUARE_A7] = BLACK_MASK | PAWN_MASK
	emptyBoard.whiteToMove = false

	emptyBoard.ApplyMove(CreateMove(SQUARE_A7, SQUARE_A5))

	assert.Equal(t, EMPTY_SQUARE, emptyBoard.board[SQUARE_A7])
	assert.Equal(t, BLACK_MASK|PAWN_MASK, emptyBoard.board[SQUARE_A5])
	assert.True(t, emptyBoard.whiteToMove)
	assert.Equal(t, emptyBoard.boardInfo.enPassantTargetSquare, SQUARE_A6)

	emptyBoard.UnapplyMove(CreateMove(SQUARE_A7, SQUARE_A5))

	assert.Equal(t, BLACK_MASK|PAWN_MASK, emptyBoard.board[SQUARE_A7])
	assert.Equal(t, EMPTY_SQUARE, emptyBoard.board[SQUARE_A5])
	assert.False(t, emptyBoard.whiteToMove)
	assert.Equal(t, emptyBoard.boardInfo.enPassantTargetSquare, EMPTY_SQUARE)
}

func TestApplyCapture(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.board[SQUARE_A2] = WHITE_MASK | PAWN_MASK
	testBoard.board[42] = BLACK_MASK | ROOK_MASK

	var m Move = CreateCapture(SQUARE_A2, 42)
	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.board[SQUARE_A2])
	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.board[42])

	testBoard.UnapplyMove(m)

	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.board[SQUARE_A2])
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.board[42])
}

func TestApplyCaptureTwice(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.board[SQUARE_A2] = WHITE_MASK | PAWN_MASK
	testBoard.board[42] = BLACK_MASK | ROOK_MASK
	testBoard.board[63] = BLACK_MASK | KNIGHT_MASK

	var m1 Move = CreateCapture(SQUARE_A2, 42)
	var m2 Move = CreateCapture(63, 42)

	testBoard.ApplyMove(m1)

	assert.Equal(t, EMPTY_SQUARE, testBoard.board[SQUARE_A2])
	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.board[42])
	assert.Equal(t, BLACK_MASK|KNIGHT_MASK, testBoard.board[63])

	testBoard.ApplyMove(m2)

	assert.Equal(t, EMPTY_SQUARE, testBoard.board[SQUARE_A2])
	assert.Equal(t, BLACK_MASK|KNIGHT_MASK, testBoard.board[42])
	assert.Equal(t, EMPTY_SQUARE, testBoard.board[63])

	testBoard.UnapplyMove(m2)

	assert.Equal(t, EMPTY_SQUARE, testBoard.board[SQUARE_A2])
	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.board[42])
	assert.Equal(t, BLACK_MASK|KNIGHT_MASK, testBoard.board[63])

	testBoard.UnapplyMove(m1)

	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.board[SQUARE_A2])
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.board[42])
	assert.Equal(t, BLACK_MASK|KNIGHT_MASK, testBoard.board[63])
}

func TestApplyWhiteKingsideCastle(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.board[25] = WHITE_MASK | KING_MASK
	testBoard.board[28] = WHITE_MASK | ROOK_MASK

	var m Move = CreateKingsideCastle(25, 27)

	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.board[25])
	assert.Equal(t, WHITE_MASK|KING_MASK, testBoard.board[27])
	assert.Equal(t, WHITE_MASK|ROOK_MASK, testBoard.board[26])
	assert.False(t, testBoard.boardInfo.whiteCanCastleKingside)

	testBoard.UnapplyMove(m)

	assert.Equal(t, WHITE_MASK|KING_MASK, testBoard.board[25])
	assert.Equal(t, WHITE_MASK|ROOK_MASK, testBoard.board[28])
	assert.True(t, testBoard.boardInfo.whiteCanCastleKingside)
}

func TestApplyBlackKingsideCastle(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.board[95] = BLACK_MASK | KING_MASK
	testBoard.board[98] = BLACK_MASK | ROOK_MASK
	testBoard.whiteToMove = false
	testBoard.boardInfo.blackCanCastleKingside = true

	var m Move = CreateKingsideCastle(95, 97)

	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.board[95])
	assert.Equal(t, BLACK_MASK|KING_MASK, testBoard.board[97])
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.board[96])
	assert.False(t, testBoard.boardInfo.blackCanCastleKingside)

	testBoard.UnapplyMove(m)

	assert.Equal(t, BLACK_MASK|KING_MASK, testBoard.board[95])
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.board[98])
	assert.True(t, testBoard.boardInfo.blackCanCastleKingside)
}

func TestApplyWhiteQueensideCastle(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.board[25] = WHITE_MASK | KING_MASK
	testBoard.board[SQUARE_A1] = WHITE_MASK | ROOK_MASK

	var m Move = CreateQueensideCastle(25, 23)

	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.board[25])
	assert.Equal(t, WHITE_MASK|KING_MASK, testBoard.board[23])
	assert.Equal(t, WHITE_MASK|ROOK_MASK, testBoard.board[24])
	assert.False(t, testBoard.boardInfo.whiteCanCastleQueenside)

	testBoard.UnapplyMove(m)

	assert.Equal(t, WHITE_MASK|KING_MASK, testBoard.board[25])
	assert.Equal(t, WHITE_MASK|ROOK_MASK, testBoard.board[SQUARE_A1])
	assert.True(t, testBoard.boardInfo.whiteCanCastleQueenside)
}

func TestApplyBlackQueensideCastle(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.board[95] = BLACK_MASK | KING_MASK
	testBoard.board[91] = BLACK_MASK | ROOK_MASK
	testBoard.whiteToMove = false
	testBoard.boardInfo.blackCanCastleQueenside = true

	var m Move = CreateQueensideCastle(95, 93)

	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.board[95])
	assert.Equal(t, BLACK_MASK|KING_MASK, testBoard.board[93])
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.board[94])
	assert.False(t, testBoard.boardInfo.blackCanCastleQueenside)

	testBoard.UnapplyMove(m)

	assert.Equal(t, BLACK_MASK|KING_MASK, testBoard.board[95])
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.board[91])
	assert.True(t, testBoard.boardInfo.blackCanCastleQueenside)
}

func TestFiddlingWithQueensideRooks(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()

	testBoard.board[95] = BLACK_MASK | KING_MASK
	testBoard.board[91] = BLACK_MASK | ROOK_MASK
	testBoard.board[25] = WHITE_MASK | KING_MASK
	testBoard.board[SQUARE_A1] = WHITE_MASK | ROOK_MASK
	testBoard.boardInfo.blackCanCastleKingside = true
	testBoard.boardInfo.blackCanCastleQueenside = true
	testBoard.boardInfo.whiteCanCastleKingside = true
	testBoard.boardInfo.whiteCanCastleQueenside = true

	var m1 Move = CreateMove(SQUARE_A1, 41)
	var m2 Move = CreateMove(91, 71)
	var m3 Move = CreateMove(41, SQUARE_A1)
	var m4 Move = CreateMove(71, 91)

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

	testBoard.board[95] = BLACK_MASK | KING_MASK
	testBoard.board[98] = BLACK_MASK | ROOK_MASK
	testBoard.board[25] = WHITE_MASK | KING_MASK
	testBoard.board[28] = WHITE_MASK | ROOK_MASK
	testBoard.boardInfo.blackCanCastleKingside = true
	testBoard.boardInfo.blackCanCastleQueenside = true
	testBoard.boardInfo.whiteCanCastleKingside = true
	testBoard.boardInfo.whiteCanCastleQueenside = true

	var m1 Move = CreateMove(28, 48)
	var m2 Move = CreateMove(98, 78)
	var m3 Move = CreateMove(48, 28)
	var m4 Move = CreateMove(78, 98)

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
