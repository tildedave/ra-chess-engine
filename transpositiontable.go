package main

type TranspositionEntry struct {
	depth       uint
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

func StoreTranspositionTable(boardState *BoardState, result *SearchResult, depth uint, searchPhase int) {
	entry := TranspositionEntry{result: result, depth: depth, searchPhase: searchPhase}

	e := boardState.transpositionTable[boardState.hashKey]

	// Don't overwrite regular hash tables with quiescent results
	if e == nil || (e.depth < depth && searchPhase <= e.searchPhase) {
		boardState.transpositionTable[boardState.hashKey] = &entry
	}
}
