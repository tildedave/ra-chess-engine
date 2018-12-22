package main

type TranspositionEntry struct {
	depth       int
	result      *SearchResult
	searchPhase int
}

var hashArray map[uint64]*TranspositionEntry

func generateTranspositionTable(boardState *BoardState) {
	boardState.transpositionTable = make(map[uint64]*TranspositionEntry)
}

func ProbeTranspositionTable(boardState *BoardState) *TranspositionEntry {
	return boardState.transpositionTable[boardState.hashKey]
}

func StoreTranspositionTable(boardState *BoardState, result *SearchResult, depth int, searchPhase int) {
	entry := TranspositionEntry{result: result, depth: depth, searchPhase: searchPhase}
	e := boardState.transpositionTable[boardState.hashKey]
	if e == nil || searchPhase < e.searchPhase {
		boardState.transpositionTable[boardState.hashKey] = &entry
	}
}
