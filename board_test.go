package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitialBoard(t *testing.T) {
	var initialBoard BoardState = CreateInitialBoardState()

	assert.Equal(t, initialBoard.PieceAtSquare(idx(0, 0)), WHITE_MASK|ROOK_MASK)
	assert.Equal(t, initialBoard.PieceAtSquare(idx(1, 0)), WHITE_MASK|KNIGHT_MASK)

	for i := 0; i < 8; i++ {
		piece := initialBoard.PieceAtSquare(idx(byte(i), 1))
		assert.Equal(t, piece, byte(WHITE_MASK|PAWN_MASK))
		assert.True(t, isPieceWhite(piece))
		assert.True(t, isPawn(piece))
	}
	for i := 2; i <= 5; i++ {
		for j := 0; j < 8; j++ {
			piece := initialBoard.PieceAtSquare(idx(byte(j), byte(i)))
			assert.Equal(t, piece, byte(0x00))
			assert.True(t, isSquareEmpty(piece))
		}
	}
	for i := 0; i < 8; i++ {
		piece := initialBoard.PieceAtSquare(idx(byte(i), 6))
		assert.Equal(t, piece, byte(BLACK_MASK|PAWN_MASK))
		assert.True(t, isPieceBlack(piece))
		assert.True(t, isPawn(piece))
	}
}

func TestToString(t *testing.T) {
	var initialBoard BoardState = CreateInitialBoardState()

	var str = initialBoard.ToString()
	assert.Equal(t, "rnbqkbnr\npppppppp\n........\n........\n........\n........\nPPPPPPPP\nRNBQKBNR\n\nrnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", str)
}

func TestToFEN(t *testing.T) {
	var initialBoard BoardState = CreateInitialBoardState()
	assert.Equal(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", initialBoard.ToFENString())

	var emptyBoard BoardState = CreateEmptyBoardState()
	assert.Equal(t, "8/8/8/8/8/8/8/8 w - - 0 1", emptyBoard.ToFENString())
}

func TestRank(t *testing.T) {
	assert.Equal(t, RANK_1, Rank(SQUARE_A1))
	assert.Equal(t, RANK_2, Rank(SQUARE_A2))
	assert.Equal(t, RANK_3, Rank(SQUARE_A3))
	assert.Equal(t, RANK_4, Rank(SQUARE_A4))
	assert.Equal(t, RANK_5, Rank(SQUARE_A5))
	assert.Equal(t, RANK_6, Rank(SQUARE_A6))
	assert.Equal(t, RANK_7, Rank(SQUARE_A7))
	assert.Equal(t, RANK_8, Rank(SQUARE_A8))
}

func TestSquareToAlgebraicString(t *testing.T) {
	assert.Equal(t, "a1", SquareToAlgebraicString(SQUARE_A1))
	assert.Equal(t, "b1", SquareToAlgebraicString(SQUARE_B1))
	assert.Equal(t, "c1", SquareToAlgebraicString(SQUARE_C1))
	assert.Equal(t, "d1", SquareToAlgebraicString(SQUARE_D1))
	assert.Equal(t, "e1", SquareToAlgebraicString(SQUARE_E1))
	assert.Equal(t, "f1", SquareToAlgebraicString(SQUARE_F1))
	assert.Equal(t, "g1", SquareToAlgebraicString(SQUARE_G1))
	assert.Equal(t, "h1", SquareToAlgebraicString(SQUARE_H1))
	assert.Equal(t, "a2", SquareToAlgebraicString(SQUARE_A2))
	assert.Equal(t, "a3", SquareToAlgebraicString(SQUARE_A3))
	assert.Equal(t, "a4", SquareToAlgebraicString(SQUARE_A4))
	assert.Equal(t, "a5", SquareToAlgebraicString(SQUARE_A5))
	assert.Equal(t, "a6", SquareToAlgebraicString(SQUARE_A6))
	assert.Equal(t, "a7", SquareToAlgebraicString(SQUARE_A7))
	assert.Equal(t, "a8", SquareToAlgebraicString(SQUARE_A8))
}

func TestBoardFromFENString(t *testing.T) {
	s := "r4rk1/pppb1pp1/3p3p/bN1Pn3/4P3/4PNqP/PPQ1B1P1/2KR2R1 w - - 5 17"
	boardState, err := CreateBoardStateFromFENString(s)

	assert.Nil(t, err)
	assert.Equal(t, s, boardState.ToFENString())
}

func TestBoardWithEnPassantSquareAndCastling(t *testing.T) {
	s := "rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 0 2"
	boardState, err := CreateBoardStateFromFENString(s)

	assert.Nil(t, err)
	assert.Equal(t, s, boardState.ToFENString())
}

func TestParseAlgebraicSquare(t *testing.T) {
	var sq uint8
	var err error

	sq, err = ParseAlgebraicSquare("a1")

	assert.Equal(t, SQUARE_A1, sq)
	assert.Nil(t, err)

	sq, err = ParseAlgebraicSquare("c6")

	assert.Equal(t, SQUARE_C6, sq)
	assert.Nil(t, err)

	sq, err = ParseAlgebraicSquare("q1")

	assert.Equal(t, uint8(0), sq)
	assert.NotNil(t, err)
	assert.Equal(t, "Column out of range: q1", err.Error())

	sq, err = ParseAlgebraicSquare("a12")

	assert.Equal(t, uint8(0), sq)
	assert.NotNil(t, err)
	assert.Equal(t, "Algebraic square was not two characters: a12", err.Error())
}

func TestThreefoldRepetition(t *testing.T) {
	boardState := CreateEmptyBoardState()
	boardState.SetPieceAtSquare(SQUARE_A1, KING_MASK|WHITE_MASK)
	boardState.SetPieceAtSquare(SQUARE_H8, KING_MASK|BLACK_MASK)
	assert.False(t, boardState.HasStateOccurred())

	for i := 1; i < 3; i++ {
		assert.True(t, boardState.RepetitionCount(i))
		assert.False(t, boardState.RepetitionCount(3))

		boardState.ApplyMove(CreateMove(SQUARE_A1, SQUARE_B1))
		boardState.ApplyMove(CreateMove(SQUARE_H8, SQUARE_H7))
		assert.False(t, boardState.RepetitionCount(3))

		boardState.ApplyMove(CreateMove(SQUARE_B1, SQUARE_A1))
		boardState.ApplyMove(CreateMove(SQUARE_H7, SQUARE_H8))
		assert.True(t, boardState.RepetitionCount(i+1))
	}

	assert.True(t, boardState.RepetitionCount(3))
}

func TestPawnMovePreventsThreefoldRepetition(t *testing.T) {
	boardState := CreateEmptyBoardState()
	boardState.SetPieceAtSquare(SQUARE_A1, KING_MASK|WHITE_MASK)
	boardState.SetPieceAtSquare(SQUARE_H8, KING_MASK|BLACK_MASK)
	boardState.sideToMove = BLACK_OFFSET
	boardState.SetPieceAtSquare(SQUARE_H6, PAWN_MASK|BLACK_MASK)
	boardState.ApplyMove(CreateMove(SQUARE_H6, SQUARE_H5))

	for i := 0; i < 2; i++ {
		assert.False(t, boardState.RepetitionCount(3))

		boardState.ApplyMove(CreateMove(SQUARE_A1, SQUARE_B1))
		boardState.ApplyMove(CreateMove(SQUARE_H8, SQUARE_H7))
		assert.False(t, boardState.RepetitionCount(3))

		boardState.ApplyMove(CreateMove(SQUARE_B1, SQUARE_A1))
		boardState.ApplyMove(CreateMove(SQUARE_H7, SQUARE_H8))
	}

	assert.False(t, boardState.RepetitionCount(3))
}
