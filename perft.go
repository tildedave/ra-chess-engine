package main

func Perft(boardState *BoardState, depth int) uint64 {
	var moveCount uint64 = 0
	if depth == 0 {
		return 1
	}

	moves := GenerateMoves(boardState)

	for _, move := range moves {
		boardState.ApplyMove(move)
		if !boardState.IsInCheck(!boardState.whiteToMove) {
			moveCount += Perft(boardState, depth-1)
		}
		boardState.UnapplyMove(move)
	}

	return moveCount
}
