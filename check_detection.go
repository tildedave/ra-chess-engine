package main

import (
	"fmt"
)

var _ = fmt.Println

func (boardState *BoardState) IsInCheck(whiteInCheck bool) bool {
	var kingSq byte
	var oppositeColorMask byte
	if whiteInCheck {
		kingSq = boardState.lookupInfo.whiteKingSquare
		oppositeColorMask = BLACK_MASK
	} else {
		kingSq = boardState.lookupInfo.blackKingSquare
		oppositeColorMask = WHITE_MASK
	}

	// test bishop rays for queen or bishop
	// test rook rays for queen or rook

	for _, pieceMask := range []byte{BISHOP_MASK, ROOK_MASK} {
		for _, offset := range offsetArr[pieceMask] {
			piece := followRay(boardState, kingSq, offset)
			// TODO(perf) should be possible to combine these checks both
			// here and in move generation w/an appropriate bit test
			if piece != SENTINEL_MASK && piece&oppositeColorMask == oppositeColorMask {
				// we care about this piece, is it the piece we're looking for
				// or a queen
				if isQueen(piece) || piece&pieceMask == pieceMask {
					// bam in check
					return true
				}
			}
		}
	}

	// test all knight squares
	var knightPiece byte
	if whiteInCheck {
		knightPiece = BLACK_MASK | KNIGHT_MASK
	} else {
		knightPiece = WHITE_MASK | KNIGHT_MASK
	}

	for _, offset := range offsetArr[KNIGHT_MASK] {
		sq := uint8(int8(kingSq) + offset)
		piece := boardState.PieceAtSquare(sq)

		if piece == knightPiece {
			return true
		}
	}
	// test pawn squares
	var pawnPiece byte
	var pawnCaptureOffsetArr [2]int8

	if whiteInCheck {
		pawnPiece = BLACK_MASK | PAWN_MASK
		pawnCaptureOffsetArr = whitePawnCaptureOffsetArr
	} else {
		pawnPiece = WHITE_MASK | PAWN_MASK
		pawnCaptureOffsetArr = blackPawnCaptureOffsetArr
	}

	for _, offset := range pawnCaptureOffsetArr {
		sq := uint8(int8(kingSq) + offset)
		piece := boardState.PieceAtSquare(sq)

		if piece == pawnPiece {
			return true
		}
	}

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
