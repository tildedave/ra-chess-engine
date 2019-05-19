package main

import (
	"fmt"
)

type TranspositionEntry struct {
	move        Move
	depth       int
	score       int
	entryType   int
	searchPhase int
}

const (
	TT_FAIL_HIGH = iota
	TT_FAIL_LOW  = iota
	TT_EXACT     = iota
)

var hashArray map[uint64]*TranspositionEntry

func generateTranspositionTable(boardState *BoardState) {
	boardState.transpositionTable = make(map[uint64]*TranspositionEntry)
}

func ProbeTranspositionTable(boardState *BoardState) *TranspositionEntry {
	return boardState.transpositionTable[boardState.hashKey]
}

func EntryTypeToString(entryType int) string {
	switch entryType {
	case TT_FAIL_HIGH:
		return "TT_FAIL_HIGH"
	case TT_FAIL_LOW:
		return "TT_FAIL_LOW"
	case TT_EXACT:
		return "TT_EXACT"
	default:
		return ""
	}
}

func (entry *TranspositionEntry) String() string {
	var entryTypeAsString string
	switch entry.entryType {
	case TT_FAIL_HIGH:
		entryTypeAsString = "beta"
	case TT_FAIL_LOW:
		entryTypeAsString = "alpha"
	case TT_EXACT:
		entryTypeAsString = "exact"
	}
	return fmt.Sprintf("{score=%d, depth=%d, type=%s}", entry.score, entry.depth, entryTypeAsString)
}

func StoreTranspositionTable(boardState *BoardState, move Move, score int, entryType int, depth int) {
	entry := TranspositionEntry{score: score, move: move, entryType: entryType, depth: depth}
	boardState.transpositionTable[boardState.hashKey] = &entry
}
