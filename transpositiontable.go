package main

import (
	"fmt"
)

type TranspositionEntry struct {
	move  Move
	depth int8
	// technically this could just be 1-2 bits, we might want to give the
	// other bytes to the score.
	entryType uint8
	// We have 8 extra bytes at the end, we just pad them on the front of the
	// score.
	score int16
}

const TT_INITIAL_SIZE = 1048576
const PAWN_ENTRY_TABLE_INITIAL_SIZE = 262144

const (
	TT_FAIL_HIGH = iota
	TT_FAIL_LOW  = iota
	TT_EXACT     = iota
)

var hashArray map[uint64]int

func generateTranspositionTable(boardState *BoardState) {
	boardState.transpositionTable = make(map[uint64]uint64, TT_INITIAL_SIZE)
	boardState.pawnTable = make(map[uint64]*PawnTableEntry, PAWN_ENTRY_TABLE_INITIAL_SIZE)
}

func ProbeTranspositionTable(boardState *BoardState) (bool, TranspositionEntry) {
	t := TranspositionEntry{}
	entry := boardState.transpositionTable[boardState.hashKey]
	if entry == 0 {
		return false, TranspositionEntry{}
	}
	var move Move
	move = SetFrom(move, uint8(entry>>(64-8)))
	move = SetTo(move, uint8((entry>>(64-16))&0xFF))
	move = SetFlags(move, uint8((entry>>(64-24))&0xFF))
	t.move = move
	t.depth = int8((entry >> (64 - 32)) & 0xFF)
	t.entryType = uint8((entry >> (64 - 40)) & 0xFF)
	t.score = int16(entry & 0xFFFFFF)
	return true, t
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
	return fmt.Sprintf("{move=%s score=%d, depth=%d, type=%s}",
		MoveToDebugString(entry.move), entry.score, entry.depth, entryTypeAsString)
}

func StoreTranspositionTable(boardState *BoardState, move Move, score int16, entryType uint8, depth int8) {
	var entry uint64
	entry |= uint64(move.From()) << (64 - 8)
	entry |= uint64(move.To()) << (64 - 16)
	entry |= uint64(move.Flags()) << (64 - 24)
	entry |= (uint64(depth) & 0xFF) << (64 - 32)
	entry |= uint64(entryType) << (64 - 40)
	entry |= uint64(score) & 0xFFFFFF
	boardState.transpositionTable[boardState.hashKey] = entry
}
