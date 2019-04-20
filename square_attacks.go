package main

func (boardState *BoardState) GetSquareAttackersBoard(sq byte) uint64 {
	whiteOccupancies := boardState.bitboards.color[WHITE_OFFSET]
	blackOccupancies := boardState.bitboards.color[BLACK_OFFSET]
	allOccupancies := whiteOccupancies | blackOccupancies
	bishopKey := hashKey(allOccupancies, boardState.moveBitboards.bishopMagics[sq])
	rookKey := hashKey(allOccupancies, boardState.moveBitboards.rookMagics[sq])
	bishopQueens := boardState.bitboards.piece[BISHOP_MASK] | boardState.bitboards.piece[QUEEN_MASK]
	rookQueens := boardState.bitboards.piece[ROOK_MASK] | boardState.bitboards.piece[QUEEN_MASK]

	return ((boardState.moveBitboards.knightAttacks[sq].board & boardState.bitboards.piece[KNIGHT_MASK]) |
		(boardState.moveBitboards.kingAttacks[sq].board & boardState.bitboards.piece[KING_MASK]) |
		(boardState.moveBitboards.bishopAttacks[sq][bishopKey].board & bishopQueens) |
		(boardState.moveBitboards.rookAttacks[sq][rookKey].board & rookQueens) |
		(boardState.moveBitboards.pawnAttacks[WHITE_OFFSET][sq] & boardState.bitboards.piece[PAWN_MASK]) |
		(boardState.moveBitboards.pawnAttacks[BLACK_OFFSET][sq] & boardState.bitboards.piece[PAWN_MASK]))
}

func (boardState *BoardState) IsSquareUnderAttack(allOccupancies uint64, sq byte, offset int, offsetForOurColor int) bool {
	occupancy := boardState.bitboards.color[offset]

	if boardState.moveBitboards.knightAttacks[sq].board&boardState.bitboards.piece[KNIGHT_MASK]&occupancy != 0 {
		return true
	}

	if boardState.moveBitboards.kingAttacks[sq].board&boardState.bitboards.piece[KING_MASK]&occupancy != 0 {
		return true
	}

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
