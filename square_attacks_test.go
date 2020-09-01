package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSquareAttackersBoard(t *testing.T) {
	boardState := CreateInitialBoardState()
	occupancies := boardState.GetAllOccupanciesBitboard()
	attackBoard := boardState.GetSquareAttackersBoard(occupancies, SQUARE_F3)

	assert.Equal(t, OnesCount64(attackBoard), 3, "Should have only had 3 attackers")
	assert.True(t, IsBitboardSet(attackBoard, SQUARE_G1))
	assert.True(t, IsBitboardSet(attackBoard, SQUARE_G2))
	assert.True(t, IsBitboardSet(attackBoard, SQUARE_E2))

	assert.Equal(t, OnesCount64(boardState.GetSquareAttackersBoard(occupancies, SQUARE_E5)), 0,
		"No attackers in initial board state")
}
