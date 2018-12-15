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
	// var offset int
	var offset int
	var offsetForOurColor int
	if colorMask == WHITE_MASK {
		offset = WHITE_OFFSET
		offsetForOurColor = BLACK_OFFSET
	} else {
		offset = BLACK_OFFSET
		offsetForOurColor = WHITE_OFFSET
	}

	occupancy := boardState.bitboards.color[offset]
	bbSq := legacySquareToBitboardSquare(sq)

	if boardState.moveBitboards.knightAttacks[bbSq].board&boardState.bitboards.piece[KNIGHT_MASK]&occupancy != 0 {
		return true
	}

	if boardState.moveBitboards.kingAttacks[bbSq].board&boardState.bitboards.piece[KING_MASK]&occupancy != 0 {
		return true
	}

	allOccupancies := occupancy | boardState.bitboards.color[offsetForOurColor]
	bishopKey := hashKey(allOccupancies, boardState.moveBitboards.bishopMagics[bbSq])

	if boardState.moveBitboards.bishopAttacks[bbSq][bishopKey].board&
		(boardState.bitboards.piece[BISHOP_MASK]|boardState.bitboards.piece[QUEEN_MASK])&occupancy != 0 {
		return true
	}

	rookKey := hashKey(allOccupancies, boardState.moveBitboards.rookMagics[bbSq])

	if boardState.moveBitboards.rookAttacks[bbSq][rookKey].board&
		(boardState.bitboards.piece[ROOK_MASK]|boardState.bitboards.piece[QUEEN_MASK])&occupancy != 0 {
		return true
	}

	// pawn attacks - pretend to be a pawn of our color, see if we're attacking a pawn of the opposite color
	// TODO: does not handle en-passant (though this is not a check detection case)
	if boardState.moveBitboards.pawnAttacks[offsetForOurColor][bbSq]&boardState.bitboards.piece[PAWN_MASK]&occupancy != 0 {
		return true
	}

	return false
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
