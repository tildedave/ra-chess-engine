package main

import (
	"math/bits"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSquareAttackersBoard(t *testing.T) {
	boardState := CreateInitialBoardState()
	attackBoard := boardState.GetSquareAttackersBoard(SQUARE_F3)

	assert.Equal(t, bits.OnesCount64(attackBoard), 3, "Should have only had 3 attackers")
	assert.True(t, IsBitboardSet(attackBoard, SQUARE_G1))
	assert.True(t, IsBitboardSet(attackBoard, SQUARE_G2))
	assert.True(t, IsBitboardSet(attackBoard, SQUARE_E2))

	assert.Equal(t, bits.OnesCount64(boardState.GetSquareAttackersBoard(SQUARE_E5)), 0,
		"No attackers in initial board state")
}
