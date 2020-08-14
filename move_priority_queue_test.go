package main

import (
	"container/heap"
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
