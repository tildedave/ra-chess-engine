package main

import (
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestInitialBoard(t *testing.T) {
    assert.Equal(t, pieceAtSquare(initialBoard, 0, 0), byte(0x04))
    assert.Equal(t, pieceAtSquare(initialBoard, 0, 1), byte(0x02))

    for i := 0; i < 8; i++ {
        piece := pieceAtSquare(initialBoard, 1, byte(i))
        assert.Equal(t, piece, byte(0x01))
        assert.True(t, isPieceWhite(piece))
        assert.True(t, isPawn(piece))
    }
    for i := 2; i <= 5; i++ {
        for j := 0; j < 8; j++ {
            piece := pieceAtSquare(initialBoard, byte(i), byte(j))
            assert.Equal(t, piece, byte(0x00))
            assert.True(t, isSquareEmpty(piece))
        }
    }
    for i := 0; i < 8; i++ {
        piece := pieceAtSquare(initialBoard, 6, byte(i))
        assert.Equal(t, piece, byte(0x81))
        assert.True(t, isPieceBlack(piece))
        assert.True(t, isPawn(piece))
    }
}

func TestToString(t *testing.T) {
    var str = boardToString(initialBoard)
    assert.Equal(t, str, "RNBQKBNR\nPPPPPPPP\n........\n........\n........\n........\npppppppp\nrnbqkbnr\n")
}
