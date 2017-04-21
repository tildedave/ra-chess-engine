package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

var _ = fmt.Println

func TestGenerateZobristHash(t *testing.T) {
	r := rand.New(rand.NewSource(0))
	h := CreateHashInfo(r)
	b := CreateInitialBoardState()

	key := b.CreateHashKey(&h)
	fmt.Println(key)
	assert.True(t, true)
}
