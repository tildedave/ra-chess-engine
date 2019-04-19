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
