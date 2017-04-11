package main

import (
	"fmt"
)

var _ = fmt.Println

func (boardState *BoardState) IsInCheck() bool {
	isWhite := boardState.whiteToMove

	var kingSq byte
	if isWhite {
		kingSq = boardState.lookupInfo.whiteKingSquare
	} else {
		kingSq = boardState.lookupInfo.blackKingSquare
	}

	// test bishop rays for queen or bishop
	// test rook rays for queen or rook

	for _, pieceMask := range []byte{BISHOP_MASK, ROOK_MASK} {
		for _, offset := range offsetArr[pieceMask] {
			piece := followRay(boardState, kingSq, offset)
			if piece != SENTINEL_MASK {
				isDestPieceWhite := piece&BLACK_MASK != BLACK_MASK
				if isWhite != isDestPieceWhite {
					// we care about this piece, is it the piece we're looking for
					// or a queen
					if isQueen(piece) || piece&pieceMask == pieceMask {
						// bam in check
						return true
					}
				}
			}
		}
	}

	// test all knight squares (dumb but mindless)
	// test pawn squares

	return false
}

func followRay(boardState *BoardState, sq byte, offset int8) byte {
	var dest byte = sq
	if offset == 0 {
		return SENTINEL_MASK
	}

	// we'll continue until we have to stop
	for true {
		dest = uint8(int8(dest) + offset)
		destPiece := boardState.board[dest]

		if destPiece == SENTINEL_MASK {
			// stop - end of the board
			return SENTINEL_MASK
		} else if destPiece == EMPTY_SQUARE {
			// keep moving
		} else {
			// got a piece, return it
			return destPiece
		}
	}

	// impossible to get here
	return SENTINEL_MASK
}
