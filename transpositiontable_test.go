package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStoreAndProbe(t *testing.T) {
	boardState := CreateInitialBoardState()

	hasEntry, entry := ProbeTranspositionTable(&boardState)
	assert.False(t, hasEntry)
	move := CreateMoveWithFlags(SQUARE_A1, SQUARE_H4, CAPTURE_MASK)
	StoreTranspositionTable(&boardState, move, 14_000, TT_EXACT, 32)

	hasEntry, entry = ProbeTranspositionTable(&boardState)
	assert.True(t, hasEntry)
	assert.Equal(t,
		TranspositionEntry{move: move, depth: 32, entryType: TT_EXACT, score: 14_000},
		entry)

	move = CreateMoveWithFlags(SQUARE_B3, SQUARE_A8, PROMOTION_MASK|CAPTURE_MASK)
	StoreTranspositionTable(&boardState, move, -32, TT_FAIL_HIGH, -10)

	hasEntry, entry = ProbeTranspositionTable(&boardState)
	assert.Equal(t,
		TranspositionEntry{move: move, depth: -10, entryType: TT_FAIL_HIGH, score: -32},
		entry)
}
