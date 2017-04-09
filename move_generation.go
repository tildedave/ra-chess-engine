package main

import (
	"fmt"
)

var _ = fmt.Println

func GenerateMoves(boardState *BoardState) []Move {
	// ray starts from going up, then clockwise around
	var offsetArr [7][8]int8
	offsetArr[6] = [8]int8{-10, -9, 1, 11, 10, 9, -1, -11}

	var pieces []Move

	for i := byte(0); i < 8; i++ {
		for j := byte(0); j < 8; j++ {
			offset := RowAndColToSquare(i, j)
			p := boardState.board[offset]
			if p == EMPTY_SQUARE {
				continue
			}

			isWhite := p&BLACK_MASK != BLACK_MASK
			if isWhite == boardState.whiteToMove {
				if !isPawn(p) {
					// 8 possible directions to go (for the queen + king + knight)
					// 4 directions for bishop + rook
					// queen, bishop, and rook will continue along a "ray" until they find their own piece
					// Knight = 2, Bishop = 3, Rook = 4, Queen = 5, King = 6
					offsets := offsetArr[p&0x0F]

					for _, value := range offsets {
						dest := uint8(int8(offset) + value)
						destPiece := boardState.board[dest]

						if destPiece == SENTINEL_MASK {
							// can't go there
							continue
						} else if destPiece == EMPTY_SQUARE {
							pieces = append(pieces, CreateMove(offset, dest))
						} else {
							isDestPieceWhite := destPiece&BLACK_MASK != BLACK_MASK
							if isWhite != isDestPieceWhite {
								pieces = append(pieces, CreateCapture(offset, dest))
							}
						}
					}
				} else {
					// pawn movement
				}
			}
		}
	}

	return pieces
}
