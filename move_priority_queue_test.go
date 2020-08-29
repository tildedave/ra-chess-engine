package main

import (
	"container/heap"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMovePriorityQueue(t *testing.T) {
	captureQueue := make(MovePriorityQueue, 0, 3)

	item1 := Item{move: CreateMove(SQUARE_A8, SQUARE_A2), score: 21}
	item2 := Item{move: CreateMove(SQUARE_A8, SQUARE_A3), score: 3}
	item3 := Item{move: CreateMove(SQUARE_A8, SQUARE_A1), score: 27}
	captureQueue.Push(&item1)
	captureQueue.Push(&item2)
	captureQueue.Push(&item3)
	heap.Init(&captureQueue)

	assert.Len(t, captureQueue, 3)
	assert.Equal(t, item3.move, heap.Pop(&captureQueue).(*Item).move)
	assert.Equal(t, item1.move, heap.Pop(&captureQueue).(*Item).move)
	assert.Equal(t, item2.move, heap.Pop(&captureQueue).(*Item).move)
}

func TestMoveSort(t *testing.T) {
	move1 := CreateMove(SQUARE_A8, SQUARE_A2)
	move2 := CreateMove(SQUARE_A8, SQUARE_A3)
	move3 := CreateMove(SQUARE_A8, SQUARE_A1)
	moves := []Move{move1, move2, move3}
	moveScores := []int{21, 3, 27}

	captureQueue := MoveSort{
		startIndex: 0,
		endIndex:   len(moves),
		moves:      &moves,
		moveScores: &moveScores,
	}
	sort.Sort(&captureQueue)

	assert.Equal(t, move3, moves[0])
	assert.Equal(t, move1, moves[1])
	assert.Equal(t, move2, moves[2])
}
