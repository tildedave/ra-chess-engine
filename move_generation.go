package main

import (
	"fmt"
)

var _ = fmt.Println

func GenerateMoves(boardState *BoardState) []Move {
	var moves []Move

	for i := byte(0); i < 8; i++ {
		for j := byte(0); j < 8; j++ {
			sq := RowAndColToSquare(i, j)
			p := boardState.board[sq]
			if p == EMPTY_SQUARE {
				continue
			}

			isWhite := p&BLACK_MASK != BLACK_MASK
			if isWhite == boardState.whiteToMove {
				if !isPawn(p) {
					moves = generatePieceMoves(boardState, p, sq, isWhite, moves)
				} else {
					moves = generatePawnMoves(boardState, p, sq, isWhite, moves)
					// TODO: pawn promotion
				}
			}
		}
	}

	// also check for castling here.
	// forbid if king is in check (calculating this TBD - possibly
	// we will allow this and forbid it later)
	// there's enemy movement to one of the castle squares

	return moves
}

// These arrays are used for both move generation + king in check detection,
// so we'll pull them out of function scope

// ray starts from going up, then clockwise around
var offsetArr = [7][8]int8{
	[8]int8{0, 0, 0, 0, 0, 0, 0, 0},
	[8]int8{0, 0, 0, 0, 0, 0, 0, 0},
	[8]int8{-19, -8, 12, 21, 19, 8, -12, -21},
	[8]int8{-11, 9, 9, 11, 0, 0, 0, 0},
	[8]int8{-10, 1, 10, -1, 0, 0, 0, 0},
	[8]int8{-10, -9, 1, 11, 10, 9, -1, -11},
	[8]int8{-10, -9, 1, 11, 10, 9, -1, -11},
}

var slidingPieces = [7]bool{false, false, false, true, true, true, false}

func generatePieceMoves(boardState *BoardState, p byte, sq byte, isWhite bool, moves []Move) []Move {

	// 8 possible directions to go (for the queen + king + knight)
	// 4 directions for bishop + rook
	// queen, bishop, and rook will continue along a "ray" until they find
	// a stopping point
	// Knight = 2, Bishop = 3, Rook = 4, Queen = 5, King = 6
	offsets := offsetArr[p&0x0F]

	var captureMask byte
	if isWhite {
		captureMask = BLACK_MASK
	} else {
		captureMask = WHITE_MASK
	}

	for _, value := range offsets {
		var dest byte = sq
		if value == 0 {
			continue
		}

		// we'll continue until we have to stop
		for true {
			dest = uint8(int8(dest) + value)
			destPiece := boardState.board[dest]

			if destPiece == SENTINEL_MASK {
				// stop - end of the board
				break
			} else if destPiece == EMPTY_SQUARE {
				// keep moving
				moves = append(moves, CreateMove(sq, dest))
			} else {
				if destPiece&captureMask == captureMask {
					moves = append(moves, CreateCapture(sq, dest))
				}

				// stop - hit another piece or made a capture
				break
			}

			// stop - piece type only gets one move
			if !slidingPieces[p&0x0F] {
				break
			}
		}
	}

	return moves
}

func generatePawnMoves(boardState *BoardState, p byte, sq byte, isWhite bool, moves []Move) []Move {
	// black is negative
	var pawnCaptureOffsetArr = [8]int8{9, 11}

	var offset int8
	if isWhite {
		offset = 10
	} else {
		offset = -10
	}

	var dest byte = uint8(int8(sq) + offset)
	if boardState.board[dest] == EMPTY_SQUARE {
		// empty square
		moves = append(moves, CreateMove(sq, dest))

		if (isWhite && sq >= SQUARE_A2 && sq <= SQUARE_H2) ||
			(!isWhite && sq >= SQUARE_A7 && sq <= SQUARE_H7) {
			// home row for white so we can move one more
			dest = uint8(int8(dest) + offset)
			if boardState.board[dest] == EMPTY_SQUARE {
				moves = append(moves, CreateMove(sq, dest))
			}
		}
	}

	for _, offset := range pawnCaptureOffsetArr {
		if !isWhite {
			offset = -offset
		}

		var dest byte = uint8(int8(sq) + offset)

		if boardState.boardInfo.enPassantTargetSquare == dest {
			moves = append(moves, CreateCapture(sq, dest))
			continue
		}

		destPiece := boardState.board[dest]
		if destPiece == SENTINEL_MASK || destPiece == EMPTY_SQUARE {
			continue
		}

		isDestPieceWhite := destPiece&BLACK_MASK != BLACK_MASK
		if isWhite != isDestPieceWhite {
			moves = append(moves, CreateCapture(sq, dest))
		}
	}

	return moves
}
