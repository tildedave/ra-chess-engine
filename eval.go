package main

import (
	"fmt"
)

var _ = fmt.Println

type BoardEval struct {
	material int
}

var materialScore = [7]int{
	0, PAWN_EVAL_SCORE, KNIGHT_EVAL_SCORE, BISHOP_EVAL_SCORE, ROOK_EVAL_SCORE, QUEEN_EVAL_SCORE, 0,
}

func Eval(boardState *BoardState) BoardEval {
	material := 0
	for i := byte(0); i < 8; i++ {
		for j := byte(0); j < 8; j++ {
			p := boardState.PieceAtSquare(RowAndColToSquare(i, j))
			if p != 0x00 && p != SENTINEL_MASK {
				score := materialScore[p&0x0F]

				// Search algorithm will invert this

				if p&0xF0 == BLACK_MASK {
					material -= score
				} else {
					material += score
				}
			}
		}
	}

	return BoardEval{material: material}
}

func (eval BoardEval) value() int {
	return eval.material
}

// TODO: incrementally update evaluation as a result of a move
