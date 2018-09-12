package main

// http://stackoverflow.com/questions/28541609/looking-for-reasonable-stack-implementation-in-golang
// rather than pulling in a dependency

type byteStack struct {
	arr []byte
}

func (s *byteStack) Push(v byte) {
	s.arr = append(s.arr, v)
}

func (s *byteStack) Pop() byte {
	l := len(s.arr)
	i := s.arr[l-1]
	s.arr = s.arr[:l-1]

	return i
}

func (s *byteStack) Peek() byte {
	return s.arr[len(s.arr)-1]
}

// Used for tests

func FlipBoardColors(boardState *BoardState) {
	for i := byte(0); i < 8; i++ {
		for j := byte(0); j < 8; j++ {
			sq := RowAndColToSquare(i, j)
			p := boardState.PieceAtSquare(sq)
			if p != 0x00 && p != SENTINEL_MASK {
				boardState.board[sq] = p ^ (WHITE_MASK | BLACK_MASK)
			}
		}
	}

	boardState.whiteToMove = !boardState.whiteToMove
}
