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

	return boardState.IsSquareUnderAttack(kingSq, oppositeColorMask)
}

func (boardState *BoardState) IsSquareUnderAttack(sq byte, colorMask byte) bool {
	for _, pieceMask := range []byte{BISHOP_MASK, ROOK_MASK} {
		for _, offset := range offsetArr[pieceMask] {
			piece := followRay(boardState, sq, offset)
			// TODO(perf) should be possible to combine these checks both
			// here and in move generation w/an appropriate bit test
			if piece != SENTINEL_MASK && piece&colorMask == colorMask {
				// we care about this piece, is it the piece we're looking for
				// or a queen
				if isQueen(piece) || piece == colorMask|pieceMask {
					// bam in check
					return true
				}
			}
		}
	}

	// test all knight squares
	knightPiece := colorMask | KNIGHT_MASK

	for _, offset := range offsetArr[KNIGHT_MASK] {
		sq := uint8(int8(sq) + offset)
		piece := boardState.PieceAtSquare(sq)

		if piece == knightPiece {
			return true
		}
	}
	// test pawn squares
	pawnPiece := colorMask | PAWN_MASK
	var pawnCaptureOffsetArr [2]int8

	if colorMask == BLACK_MASK {
		pawnCaptureOffsetArr = whitePawnCaptureOffsetArr
	} else {
		pawnCaptureOffsetArr = blackPawnCaptureOffsetArr
	}

	for _, offset := range pawnCaptureOffsetArr {
		sq := uint8(int8(sq) + offset)
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

func (boardState *BoardState) TestCastleLegality(move Move) bool {
	if boardState.whiteToMove {
		if move.IsKingsideCastle() {
			// test	if F1 is being attacked
			return !boardState.IsSquareUnderAttack(SQUARE_F1, BLACK_MASK) &&
				!boardState.IsSquareUnderAttack(SQUARE_E1, BLACK_MASK)
		}

		// test	if B1 or C1 are being attacked
		return !boardState.IsSquareUnderAttack(SQUARE_D1, BLACK_MASK) &&
			!boardState.IsSquareUnderAttack(SQUARE_C1, BLACK_MASK) &&
			!boardState.IsSquareUnderAttack(SQUARE_E1, BLACK_MASK)
	}

	if move.IsKingsideCastle() {
		return !boardState.IsSquareUnderAttack(SQUARE_F8, WHITE_MASK) &&
			!boardState.IsSquareUnderAttack(SQUARE_E8, WHITE_MASK)
	}

	return !boardState.IsSquareUnderAttack(SQUARE_D8, WHITE_MASK) &&
		!boardState.IsSquareUnderAttack(SQUARE_C8, WHITE_MASK) &&
		!boardState.IsSquareUnderAttack(SQUARE_E8, WHITE_MASK)
}
