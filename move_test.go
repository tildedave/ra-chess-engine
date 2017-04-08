package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"unsafe"
)

func TestCreateMove(t *testing.T) {
	var m = CreateMove(31, 51)

	assert.Equal(t, uintptr(3), unsafe.Sizeof(m))
	assert.Equal(t, Move{from: 31, to: 51}, m)
}

func TestCreateCapture(t *testing.T) {
	var m = CreateCapture(31, 51)

	assert.Equal(t, Move{from: 31, to: 51, flags: 0x80}, m)
}

func TestMoveToString(t *testing.T) {
	assert.Equal(t, "a2xa4", MoveToString(CreateCapture(31, 51)))
	assert.Equal(t, "a2-a4", MoveToString(CreateMove(31, 51)))
	assert.Equal(t, "O-O", MoveToString(CreateKingsideCastle(25, 27)))
	assert.Equal(t, "O-O-O", MoveToString(CreateQueensideCastle(25, 23)))
}
