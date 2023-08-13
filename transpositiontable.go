package main

import (
	"fmt"
)

type TranspositionEntry struct {
	move      Move
	depth     int8
	score     int16
	entryType uint8
	// technically this could just be 1-2 bits, we might want to give the
	// other bytes to the score.
}

const TT_INITIAL_SIZE = 1048576
const PAWN_ENTRY_TABLE_INITIAL_SIZE = 262144

const (
	TT_FAIL_HIGH = iota
	TT_FAIL_LOW  = iota
	TT_EXACT     = iota
)

var hashArray map[uint64]*TranspositionEntry

func generateTranspositionTable(boardState *BoardState) {
	boardState.transpositionTable = make(map[uint64]*TranspositionEntry, TT_INITIAL_SIZE)
	boardState.pawnTable = make(map[uint64]*PawnTableEntry, PAWN_ENTRY_TABLE_INITIAL_SIZE)
}

func ProbeTranspositionTable(boardState *BoardState) *TranspositionEntry {
	return boardState.transpositionTable[boardState.hashKey]
}

func EntryTypeToString(entryType uint8) string {
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

func StoreTranspositionTable(boardState *BoardState, move Move, score int16, entryType uint8, depth int8) {
	// Try to avoid a new heap allocation if we already have something at this hash key.
	var entry *TranspositionEntry
	if boardState.transpositionTable[boardState.hashKey] != nil {
		entry = boardState.transpositionTable[boardState.hashKey]
	} else {
		entry = &(TranspositionEntry{})
	}
	entry.score = score
	entry.move = move
	entry.entryType = entryType
	entry.depth = depth
	boardState.transpositionTable[boardState.hashKey] = entry
}
