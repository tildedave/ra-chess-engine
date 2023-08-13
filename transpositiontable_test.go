package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStoreAndProbe(t *testing.T) {
	boardState := CreateInitialBoardState()

	assert.Nil(t, ProbeTranspositionTable(&boardState))
	move := Move{from: SQUARE_A1, to: SQUARE_H4, flags: CAPTURE_MASK}
	StoreTranspositionTable(&boardState, move, 14_000, TT_EXACT, 32)

	assert.Equal(t,
		TranspositionEntry{move: move, depth: 32, entryType: TT_EXACT, score: 14_000},
		*ProbeTranspositionTable(&boardState))

	move = Move{from: SQUARE_B3, to: SQUARE_A8, flags: PROMOTION_MASK | CAPTURE_MASK}
	StoreTranspositionTable(&boardState, move, -32, TT_FAIL_HIGH, -10)

	assert.Equal(t,
		TranspositionEntry{move: move, depth: -10, entryType: TT_FAIL_HIGH, score: -32},
		*ProbeTranspositionTable(&boardState))
}
