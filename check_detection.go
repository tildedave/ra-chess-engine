package main

import (
	"fmt"
	"math/bits"
)

var _ = fmt.Println

func (boardState *BoardState) IsInCheck(offset int) bool {
	var kingSq byte
	var oppositeColorOffset int
	if offset == WHITE_OFFSET {
		kingSq = byte(bits.TrailingZeros64(boardState.bitboards.color[WHITE_OFFSET] & boardState.bitboards.piece[KING_MASK]))
		oppositeColorOffset = BLACK_OFFSET
	} else {
		kingSq = byte(bits.TrailingZeros64(boardState.bitboards.color[BLACK_OFFSET] & boardState.bitboards.piece[KING_MASK]))
		oppositeColorOffset = WHITE_OFFSET
	}

	return boardState.IsSquareUnderAttack(kingSq, oppositeColorOffset, offset)
}

func (boardState *BoardState) IsSquareUnderAttack(sq byte, offset int, offsetForOurColor int) bool {
	occupancy := boardState.bitboards.color[offset]

	if boardState.moveBitboards.knightAttacks[sq].board&boardState.bitboards.piece[KNIGHT_MASK]&occupancy != 0 {
		return true
	}

	if boardState.moveBitboards.kingAttacks[sq].board&boardState.bitboards.piece[KING_MASK]&occupancy != 0 {
		return true
	}

	allOccupancies := occupancy | boardState.bitboards.color[offsetForOurColor]
	bishopKey := hashKey(allOccupancies, boardState.moveBitboards.bishopMagics[sq])

	if boardState.moveBitboards.bishopAttacks[sq][bishopKey].board&
		(boardState.bitboards.piece[BISHOP_MASK]|boardState.bitboards.piece[QUEEN_MASK])&occupancy != 0 {
		return true
	}

	rookKey := hashKey(allOccupancies, boardState.moveBitboards.rookMagics[sq])

	if boardState.moveBitboards.rookAttacks[sq][rookKey].board&
		(boardState.bitboards.piece[ROOK_MASK]|boardState.bitboards.piece[QUEEN_MASK])&occupancy != 0 {
		return true
	}

	// pawn attacks - pretend to be a pawn of our color, see if we're attacking a pawn of the opposite color
	// TODO: does not handle en-passant (though this is not a check detection case)
	if boardState.moveBitboards.pawnAttacks[offsetForOurColor][sq]&boardState.bitboards.piece[PAWN_MASK]&occupancy != 0 {
		return true
	}

	return false
}

func (boardState *BoardState) GetSquareAttackersBoard(sq byte) uint64 {
	whiteOccupancies := boardState.bitboards.color[WHITE_OFFSET]
	blackOccupancies := boardState.bitboards.color[BLACK_OFFSET]
	allOccupancies := whiteOccupancies | blackOccupancies
	bishopKey := hashKey(allOccupancies, boardState.moveBitboards.bishopMagics[sq])
	rookKey := hashKey(allOccupancies, boardState.moveBitboards.rookMagics[sq])
	bishopQueens := boardState.bitboards.piece[BISHOP_MASK] | boardState.bitboards.piece[QUEEN_MASK]
	rookQueens := boardState.bitboards.piece[BISHOP_MASK] | boardState.bitboards.piece[QUEEN_MASK]

	return ((boardState.moveBitboards.knightAttacks[sq].board & boardState.bitboards.piece[KNIGHT_MASK]) |
		(boardState.moveBitboards.kingAttacks[sq].board & boardState.bitboards.piece[KING_MASK]) |
		(boardState.moveBitboards.bishopAttacks[sq][bishopKey].board & bishopQueens) |
		(boardState.moveBitboards.bishopAttacks[sq][rookKey].board & rookQueens) |
		(boardState.moveBitboards.pawnAttacks[WHITE_OFFSET][sq] & whiteOccupancies & boardState.bitboards.piece[PAWN_MASK]) |
		(boardState.moveBitboards.pawnAttacks[BLACK_OFFSET][sq] & blackOccupancies & boardState.bitboards.piece[PAWN_MASK]))
}

func (boardState *BoardState) FilterChecks(moves []Move) []Move {
	offset := oppositeColorOffset(boardState.offsetToMove)
	enemyKingSq := bits.TrailingZeros64(boardState.bitboards.piece[KING_MASK] & boardState.bitboards.color[offset])

	allOccupancies := boardState.bitboards.color[WHITE_OFFSET] | boardState.bitboards.color[BLACK_OFFSET]
	bishopKey := hashKey(allOccupancies, boardState.moveBitboards.bishopMagics[enemyKingSq])
	rookKey := hashKey(allOccupancies, boardState.moveBitboards.rookMagics[enemyKingSq])
	bishopMask := boardState.moveBitboards.bishopAttacks[enemyKingSq][bishopKey].board
	rookMask := boardState.moveBitboards.rookAttacks[enemyKingSq][rookKey].board
	checks := make([]Move, 0, len(moves))

	for _, move := range moves {
		switch boardState.board[move.from] & 0x0F {
		case KNIGHT_MASK:
			if IsBitboardSet(boardState.moveBitboards.knightAttacks[enemyKingSq].board, move.to) {
				checks = append(checks, move)
			}
		case PAWN_MASK:
			if IsBitboardSet(boardState.moveBitboards.pawnAttacks[offset][enemyKingSq], move.to) {
				checks = append(checks, move)
			}
		case QUEEN_MASK:
			if IsBitboardSet(bishopMask, move.to) || IsBitboardSet(rookMask, move.to) {
				checks = append(checks, move)
			}
		case BISHOP_MASK:
			if IsBitboardSet(bishopMask, move.to) {
				checks = append(checks, move)
			}
		case ROOK_MASK:
			if IsBitboardSet(rookMask, move.to) {
				checks = append(checks, move)
			}
		}
	}

	return checks
}

func (boardState *BoardState) TestCastleLegality(move Move) bool {
	if boardState.offsetToMove == WHITE_OFFSET {
		if move.IsKingsideCastle() {
			// test	if F1 is being attacked
			return !boardState.IsSquareUnderAttack(SQUARE_F1, BLACK_OFFSET, WHITE_OFFSET) &&
				!boardState.IsSquareUnderAttack(SQUARE_E1, BLACK_OFFSET, WHITE_OFFSET)
		}

		// test	if B1 or C1 are being attacked
		return !boardState.IsSquareUnderAttack(SQUARE_D1, BLACK_OFFSET, WHITE_OFFSET) &&
			!boardState.IsSquareUnderAttack(SQUARE_C1, BLACK_OFFSET, WHITE_OFFSET) &&
			!boardState.IsSquareUnderAttack(SQUARE_E1, BLACK_OFFSET, WHITE_OFFSET)
	}

	if move.IsKingsideCastle() {
		return !boardState.IsSquareUnderAttack(SQUARE_F8, WHITE_OFFSET, BLACK_OFFSET) &&
			!boardState.IsSquareUnderAttack(SQUARE_E8, WHITE_OFFSET, BLACK_OFFSET)
	}

	return !boardState.IsSquareUnderAttack(SQUARE_D8, WHITE_OFFSET, BLACK_OFFSET) &&
		!boardState.IsSquareUnderAttack(SQUARE_C8, WHITE_OFFSET, BLACK_OFFSET) &&
		!boardState.IsSquareUnderAttack(SQUARE_E8, WHITE_OFFSET, BLACK_OFFSET)
}
