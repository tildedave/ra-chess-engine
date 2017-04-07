package main

import (
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestInitialBoard(t *testing.T) {
    assert.Equal(t, pieceAtSquare(initialBoard, 0, 0), byte(0x04))
}
