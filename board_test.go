package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var _ = fmt.Println

func TestInitialBoard(t *testing.T) {
	var initialBoard BoardState = CreateInitialBoardState()

	assert.Equal(t, PieceAtSquare(initialBoard, RowAndColToSquare(0, 0)), byte(0x04))
	assert.Equal(t, PieceAtSquare(initialBoard, RowAndColToSquare(0, 1)), byte(0x02))

	for i := 0; i < 8; i++ {
		piece := PieceAtSquare(initialBoard, RowAndColToSquare(1, byte(i)))
		assert.Equal(t, piece, byte(0x01))
		assert.True(t, isPieceWhite(piece))
		assert.True(t, isPawn(piece))
	}
	for i := 2; i <= 5; i++ {
		for j := 0; j < 8; j++ {
			piece := PieceAtSquare(initialBoard, RowAndColToSquare(byte(i), byte(j)))
			assert.Equal(t, piece, byte(0x00))
			assert.True(t, isSquareEmpty(piece))
		}
	}
	for i := 0; i < 8; i++ {
		piece := PieceAtSquare(initialBoard, RowAndColToSquare(6, byte(i)))
		assert.Equal(t, piece, byte(0x81))
		assert.True(t, isPieceBlack(piece))
		assert.True(t, isPawn(piece))
	}
}

func TestToString(t *testing.T) {
	var initialBoard BoardState = CreateInitialBoardState()

	var str = BoardToString(initialBoard)
	assert.Equal(t, "rnbqkbnr\npppppppp\n........\n........\n........\n........\nPPPPPPPP\nRNBQKBNR\n", str)
}

func TestToFEN(t *testing.T) {
	var b BoardState = CreateInitialBoardState()

	var str = BoardStateToFENString(b)
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
