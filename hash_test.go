package main

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _ = fmt.Println

func TestGenerateZobristHash(t *testing.T) {
	r := rand.New(rand.NewSource(0))
	h := CreateHashInfo(r)
	b := CreateInitialBoardState()

	key := b.CreateHashKey(&h)

	var expectedKey uint64 = 10674149984763701137
	assert.Equal(t, key, expectedKey)
}
