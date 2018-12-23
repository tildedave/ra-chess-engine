package main

type TranspositionEntry struct {
	move        Move
	depth       uint
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

func StoreTranspositionTable(boardState *BoardState, move Move, score int, entryType int, depth uint, searchPhase int) {
	entry := TranspositionEntry{score: score, move: move, entryType: entryType, depth: depth, searchPhase: searchPhase}

	e := boardState.transpositionTable[boardState.hashKey]

	// Don't overwrite regular hash tables with quiescent results
	if e == nil || (e.depth < depth && searchPhase <= e.searchPhase) {
		boardState.transpositionTable[boardState.hashKey] = &entry
	}
}
