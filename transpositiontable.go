package main

type TranspositionEntry struct {
	depth  uint
	result *SearchResult
}

var hashArray map[uint64]*TranspositionEntry

func generateTranspositionTable(boardState *BoardState) {
	boardState.transpositionTable = make(map[uint64]*TranspositionEntry)
}

func ProbeTranspositionTable(boardState *BoardState) *TranspositionEntry {
	return boardState.transpositionTable[boardState.hashKey]
}

func StoreTranspositionTable(boardState *BoardState, result *SearchResult, depth uint) {
	entry := TranspositionEntry{result: result, depth: depth}
	// TODO: don't store if depth is lower than depth in table
	e := boardState.transpositionTable[boardState.hashKey]
	if e == nil || e.depth < depth {
		boardState.transpositionTable[boardState.hashKey] = &entry
	}
}
