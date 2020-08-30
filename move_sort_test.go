package main

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMoveSort(t *testing.T) {
	move1 := CreateMove(SQUARE_A8, SQUARE_A2)
	move2 := CreateMove(SQUARE_A8, SQUARE_A3)
	move3 := CreateMove(SQUARE_A8, SQUARE_A1)
	moves := []Move{move1, move2, move3}
	moveScores := []int{21, 3, 27}

	captureQueue := MoveSort{
		startIndex: 0,
		endIndex:   len(moves),
		moves:      moves[:],
		moveScores: moveScores[:],
	}
	sort.Sort(&captureQueue)

	assert.Equal(t, move3, moves[0])
	assert.Equal(t, move1, moves[1])
	assert.Equal(t, move2, moves[2])
}
