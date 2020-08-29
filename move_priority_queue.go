package main

import (
	"container/heap"
	"fmt"
)

// from https://golang.org/pkg/container/heap/

type Item struct {
	move  Move
	score int
	index int
}

func (item *Item) String() string {
	return fmt.Sprintf("%s (%d)", MoveToXboardString(item.move), item.score)
}

type MoveSort struct {
	moves      *[]Move
	moveScores *[]int
	startIndex int
	endIndex   int
}

func (pq MoveSort) Len() int {
	return pq.endIndex - pq.startIndex
}

func (pq MoveSort) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	idx := pq.startIndex
	return (*pq.moveScores)[idx+i] > (*pq.moveScores)[idx+j]
}

func (pq MoveSort) Swap(i, j int) {
	idx := pq.startIndex
	(*pq.moves)[idx+i], (*pq.moves)[idx+j] = (*pq.moves)[idx+j], (*pq.moves)[idx+i]
	(*pq.moveScores)[idx+i], (*pq.moveScores)[idx+j] = (*pq.moveScores)[idx+j], (*pq.moveScores)[idx+i]
}

type MovePriorityQueue []*Item

func (pq MovePriorityQueue) Len() int { return len(pq) }

func (pq MovePriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].score > pq[j].score
}

func (pq MovePriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *MovePriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *MovePriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and move of an Item in the queue.
func (pq *MovePriorityQueue) update(item *Item, move Move, priority int) {
	item.move = move
	item.score = priority
	heap.Fix(pq, item.index)
}
