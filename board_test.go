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

	var str = initialBoard.ToFENString()
	assert.Equal(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", str)
}

func TestSquareToAlgebraicString(t *testing.T) {
	assert.Equal(t, "??", SquareToAlgebraicString(0))
	assert.Equal(t, "??", SquareToAlgebraicString(10))
	assert.Equal(t, "??", SquareToAlgebraicString(20))
	assert.Equal(t, "a1", SquareToAlgebraicString(21))
	assert.Equal(t, "b1", SquareToAlgebraicString(22))
	assert.Equal(t, "c1", SquareToAlgebraicString(23))
	assert.Equal(t, "d1", SquareToAlgebraicString(24))
	assert.Equal(t, "e1", SquareToAlgebraicString(25))
	assert.Equal(t, "f1", SquareToAlgebraicString(26))
	assert.Equal(t, "g1", SquareToAlgebraicString(27))
	assert.Equal(t, "h1", SquareToAlgebraicString(28))
	assert.Equal(t, "??", SquareToAlgebraicString(29))
	assert.Equal(t, "??", SquareToAlgebraicString(30))
	assert.Equal(t, "a2", SquareToAlgebraicString(31))
	assert.Equal(t, "a3", SquareToAlgebraicString(41))
	assert.Equal(t, "a4", SquareToAlgebraicString(51))
	assert.Equal(t, "a5", SquareToAlgebraicString(61))
	assert.Equal(t, "a6", SquareToAlgebraicString(71))
	assert.Equal(t, "a7", SquareToAlgebraicString(81))
	assert.Equal(t, "a8", SquareToAlgebraicString(91))
}

func TestApplyMove(t *testing.T) {
	var emptyBoard BoardState = CreateEmptyBoardState()
	emptyBoard.board[31] = WHITE_MASK | PAWN_MASK

	emptyBoard.ApplyMove(CreateMove(31, 51))

	assert.Equal(t, EMPTY_SQUARE, emptyBoard.board[31])
	assert.Equal(t, WHITE_MASK|PAWN_MASK, emptyBoard.board[51])
	assert.False(t, emptyBoard.whiteToMove)

	emptyBoard.UnapplyMove(CreateMove(31, 51))

	assert.Equal(t, WHITE_MASK|PAWN_MASK, emptyBoard.board[31])
	assert.Equal(t, EMPTY_SQUARE, emptyBoard.board[51])
	assert.True(t, emptyBoard.whiteToMove)
}

func TestApplyCapture(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.board[31] = WHITE_MASK | PAWN_MASK
	testBoard.board[42] = BLACK_MASK | ROOK_MASK

	var m Move = CreateCapture(31, 42)
	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.board[31])
	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.board[42])

	testBoard.UnapplyMove(m)

	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.board[31])
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.board[42])
}

func TestApplyCaptureTwice(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.board[31] = WHITE_MASK | PAWN_MASK
	testBoard.board[42] = BLACK_MASK | ROOK_MASK
	testBoard.board[63] = BLACK_MASK | KNIGHT_MASK

	var m1 Move = CreateCapture(31, 42)
	var m2 Move = CreateCapture(63, 42)

	testBoard.ApplyMove(m1)

	assert.Equal(t, EMPTY_SQUARE, testBoard.board[31])
	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.board[42])
	assert.Equal(t, BLACK_MASK|KNIGHT_MASK, testBoard.board[63])

	testBoard.ApplyMove(m2)

	assert.Equal(t, EMPTY_SQUARE, testBoard.board[31])
	assert.Equal(t, BLACK_MASK|KNIGHT_MASK, testBoard.board[42])
	assert.Equal(t, EMPTY_SQUARE, testBoard.board[63])

	testBoard.UnapplyMove(m2)

	assert.Equal(t, EMPTY_SQUARE, testBoard.board[31])
	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.board[42])
	assert.Equal(t, BLACK_MASK|KNIGHT_MASK, testBoard.board[63])

	testBoard.UnapplyMove(m1)

	assert.Equal(t, WHITE_MASK|PAWN_MASK, testBoard.board[31])
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
	assert.False(t, testBoard.whiteCanCastleKingside)

	testBoard.UnapplyMove(m)

	assert.Equal(t, WHITE_MASK|KING_MASK, testBoard.board[25])
	assert.Equal(t, WHITE_MASK|ROOK_MASK, testBoard.board[28])
	assert.True(t, testBoard.whiteCanCastleKingside)
}

func TestApplyBlackKingsideCastle(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.board[95] = BLACK_MASK | KING_MASK
	testBoard.board[98] = BLACK_MASK | ROOK_MASK
	testBoard.whiteToMove = false

	var m Move = CreateKingsideCastle(95, 97)

	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.board[95])
	assert.Equal(t, BLACK_MASK|KING_MASK, testBoard.board[97])
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.board[96])
	assert.False(t, testBoard.blackCanCastleKingside)

	testBoard.UnapplyMove(m)

	assert.Equal(t, BLACK_MASK|KING_MASK, testBoard.board[95])
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.board[98])
	assert.True(t, testBoard.blackCanCastleKingside)
}

func TestApplyWhiteQueensideCastle(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.board[25] = WHITE_MASK | KING_MASK
	testBoard.board[21] = WHITE_MASK | ROOK_MASK

	var m Move = CreateQueensideCastle(25, 23)

	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.board[25])
	assert.Equal(t, WHITE_MASK|KING_MASK, testBoard.board[23])
	assert.Equal(t, WHITE_MASK|ROOK_MASK, testBoard.board[24])
	assert.False(t, testBoard.whiteCanCastleQueenside)

	testBoard.UnapplyMove(m)

	assert.Equal(t, WHITE_MASK|KING_MASK, testBoard.board[25])
	assert.Equal(t, WHITE_MASK|ROOK_MASK, testBoard.board[21])
	assert.True(t, testBoard.whiteCanCastleQueenside)
}

func TestApplyBlackQueensideCastle(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.board[95] = BLACK_MASK | KING_MASK
	testBoard.board[91] = BLACK_MASK | ROOK_MASK
	testBoard.whiteToMove = false

	var m Move = CreateQueensideCastle(95, 93)

	testBoard.ApplyMove(m)

	assert.Equal(t, EMPTY_SQUARE, testBoard.board[95])
	assert.Equal(t, BLACK_MASK|KING_MASK, testBoard.board[93])
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.board[94])
	assert.False(t, testBoard.blackCanCastleQueenside)

	testBoard.UnapplyMove(m)

	assert.Equal(t, BLACK_MASK|KING_MASK, testBoard.board[95])
	assert.Equal(t, BLACK_MASK|ROOK_MASK, testBoard.board[91])
	assert.True(t, testBoard.blackCanCastleQueenside)
}
