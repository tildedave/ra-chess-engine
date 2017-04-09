package main

func Perft(boardState *BoardState, depth int) uint {
	var moveCount uint = 0
	if depth == 0 {
		return 1
	}

	moves := GenerateMoves(boardState)

	for _, move := range moves {
		boardState.ApplyMove(move)
		moveCount += Perft(boardState, depth-1)
		boardState.UnapplyMove(move)
	}

	return moveCount
}
