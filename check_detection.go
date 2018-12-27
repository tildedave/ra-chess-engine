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

	allOccupancies := boardState.bitboards.color[WHITE_OFFSET] | boardState.bitboards.color[BLACK_OFFSET]
	return boardState.IsSquareUnderAttack(allOccupancies, kingSq, oppositeColorOffset, offset)
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
	allOccupancies := boardState.bitboards.color[WHITE_OFFSET] | boardState.bitboards.color[BLACK_OFFSET]
	if boardState.offsetToMove == WHITE_OFFSET {
		if move.IsKingsideCastle() {
			// test	if F1 is being attacked
			return !boardState.IsSquareUnderAttack(allOccupancies, SQUARE_F1, BLACK_OFFSET, WHITE_OFFSET) &&
				!boardState.IsSquareUnderAttack(allOccupancies, SQUARE_E1, BLACK_OFFSET, WHITE_OFFSET)
		}

		// test	if B1 or C1 are being attacked
		return !boardState.IsSquareUnderAttack(allOccupancies, SQUARE_D1, BLACK_OFFSET, WHITE_OFFSET) &&
			!boardState.IsSquareUnderAttack(allOccupancies, SQUARE_C1, BLACK_OFFSET, WHITE_OFFSET) &&
			!boardState.IsSquareUnderAttack(allOccupancies, SQUARE_E1, BLACK_OFFSET, WHITE_OFFSET)
	}

	if move.IsKingsideCastle() {
		return !boardState.IsSquareUnderAttack(allOccupancies, SQUARE_F8, WHITE_OFFSET, BLACK_OFFSET) &&
			!boardState.IsSquareUnderAttack(allOccupancies, SQUARE_E8, WHITE_OFFSET, BLACK_OFFSET)
	}

	return !boardState.IsSquareUnderAttack(allOccupancies, SQUARE_D8, WHITE_OFFSET, BLACK_OFFSET) &&
		!boardState.IsSquareUnderAttack(allOccupancies, SQUARE_C8, WHITE_OFFSET, BLACK_OFFSET) &&
		!boardState.IsSquareUnderAttack(allOccupancies, SQUARE_E8, WHITE_OFFSET, BLACK_OFFSET)
}

func (boardState *BoardState) IsCheckmate() bool {
	if !boardState.IsInCheck(boardState.offsetToMove) {
		return false
	}

	moves := GenerateMoves(boardState)
	offset := boardState.offsetToMove

	for _, move := range moves {
		boardState.ApplyMove(move)
		inCheck := boardState.IsInCheck(offset)
		boardState.UnapplyMove(move)
		if !inCheck {
			return false
		}
	}

	return true
}

func (boardState *BoardState) IsCheckmate() bool {
	if !boardState.IsInCheck(boardState.offsetToMove) {
		return false
	}

	moves := GenerateMoves(boardState)
	offset := boardState.offsetToMove

	for _, move := range moves {
		boardState.ApplyMove(move)
		inCheck := boardState.IsInCheck(offset)
		boardState.UnapplyMove(move)
		if !inCheck {
			return false
		}
	}

	return true
}
