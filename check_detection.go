package main

import (
	"fmt"
	"math/bits"
)

type CheckDetectionInfo struct {
	bishopMask uint64
	rookMask   uint64
	knightMask uint64
	pawnMask   uint64
}

func makeCheckDetectionInfo(boardState *BoardState) CheckDetectionInfo {
	otherSide := oppositeColorOffset(boardState.sideToMove)

	enemyKingSq := bits.TrailingZeros64(boardState.bitboards.piece[KING_MASK] & boardState.bitboards.color[otherSide])
	allOccupancies := boardState.bitboards.color[WHITE_OFFSET] | boardState.bitboards.color[BLACK_OFFSET]
	bishopKey := hashKey(allOccupancies, boardState.moveBitboards.bishopMagics[enemyKingSq])
	rookKey := hashKey(allOccupancies, boardState.moveBitboards.rookMagics[enemyKingSq])
	bishopMask := boardState.moveBitboards.bishopAttacks[enemyKingSq][bishopKey].board
	rookMask := boardState.moveBitboards.rookAttacks[enemyKingSq][rookKey].board

	return CheckDetectionInfo{
		bishopMask: bishopMask,
		rookMask:   rookMask,
		knightMask: boardState.moveBitboards.knightAttacks[enemyKingSq].board,
		pawnMask:   boardState.moveBitboards.pawnAttacks[otherSide][enemyKingSq],
	}
}

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

	// Quiescent search can end up capturing the king, we need to prevent this somehow.  This is one way /shrug
	if kingSq == 64 {
		return true
	}

	allOccupancies := boardState.bitboards.color[WHITE_OFFSET] | boardState.bitboards.color[BLACK_OFFSET]
	return boardState.IsSquareUnderAttack(allOccupancies, kingSq, oppositeColorOffset, offset)
}

func (boardState *BoardState) IsMoveCheck(move Move, checkInfo *CheckDetectionInfo) bool {
	switch boardState.board[move.From()] & 0x0F {
	case KNIGHT_MASK:
		return IsBitboardSet(checkInfo.knightMask, move.To())
	case PAWN_MASK:
		return IsBitboardSet(checkInfo.pawnMask, move.To())
	case QUEEN_MASK:
		return IsBitboardSet(checkInfo.bishopMask, move.To()) ||
			IsBitboardSet(checkInfo.rookMask, move.To())
	case BISHOP_MASK:
		return IsBitboardSet(checkInfo.bishopMask, move.To())
	case ROOK_MASK:
		return IsBitboardSet(checkInfo.rookMask, move.To())
	}

	return false
}

func (boardState *BoardState) TestCastleLegality(move Move) bool {
	allOccupancies := boardState.bitboards.color[WHITE_OFFSET] | boardState.bitboards.color[BLACK_OFFSET]
	if boardState.sideToMove == WHITE_OFFSET {
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
	if !boardState.IsInCheck(boardState.sideToMove) {
		return false
	}

	moves := make([]Move, 256)
	end := GenerateMoves(boardState, moves[:], 0)
	offset := boardState.sideToMove

	for _, move := range moves[0:end] {
		boardState.ApplyMove(move)
		inCheck := boardState.IsInCheck(offset)
		boardState.UnapplyMove(move)
		if !inCheck {
			return false
		}
	}

	return true
}
