package main

// GetLeastValuableAttacker returns the bitboard associated with the least valuable
// member of a bitboard and the piece mask of the least valuable attacke.  Its
// intended use is to be passed a bitboard generated by GetSquareAttackersBoard.
// https://www.chessprogramming.org/SEE_-_The_Swap_Algorithm
func GetLeastValuableAttacker(boardState *BoardState, attackers uint64) (uint64, byte) {
	for _, piece := range []byte{PAWN_MASK, KNIGHT_MASK, BISHOP_MASK, ROOK_MASK, QUEEN_MASK, KING_MASK} {
		match := boardState.bitboards.piece[piece] & attackers
		if match != 0 {
			return match & -match, piece
		}
	}

	return 0, 0
}

// StaticExchangeEvaluation returns the expected gain from a capture on a given square.
// It will be used to prune the quiesecent search tree so we don't examine garbage captures.
// https://www.chessprogramming.org/SEE_-_The_Swap_Algorithm
func StaticExchangeEvaluation(boardState *BoardState, destSq byte, fromPiece byte, fromSq byte) int {
	occupancies := boardState.GetAllOccupanciesBitboard()
	attackers := boardState.GetSquareAttackersBoard(occupancies, destSq)
	offset := boardState.sideToMove
	var gain [32]int
	initialPiece := boardState.board[destSq]
	if initialPiece == 0 {
		panic("Called StaticExchangeEvaluation on square without piece")
	}
	gain[0] = MATERIAL_SCORE[initialPiece&0x0F]
	i := 0
	attackerBoard := SetBitboard(0, fromSq)

	for attackers != 0 {
		i++
		gain[i] = MATERIAL_SCORE[fromPiece] - gain[i-1]

		// clear attacker from allOccupancies, so we compute sliding piece attacks correctly
		// TODO: recompute the attacker board in a smarter way, see wiki page.
		occupancies ^= attackerBoard
		attackers = occupancies & boardState.GetSquareAttackersBoard(occupancies, destSq)

		// Swap sides
		offset = oppositeColorOffset(offset)
		offsetOccupancies := occupancies & boardState.bitboards.color[offset]
		attackerBoard, fromPiece = GetLeastValuableAttacker(boardState, attackers&offsetOccupancies)
	}

	for i > 1 {
		i--
		gain[i-1] = -Max(-gain[i-1], gain[i])
	}

	return gain[0]
}

func (boardState *BoardState) FilterSEECaptures(moves []Move, start int, end int) int {
	for start < end {
		capture := moves[start]
		fromPiece := boardState.board[capture.From()] & 0x0F
		if fromPiece == PAWN_MASK {
			start++
		} else if boardState.board[capture.To()] == EMPTY_SQUARE {
			moves[start] = moves[start+1]
			end--
		} else if StaticExchangeEvaluation(boardState, capture.To(), fromPiece, capture.From()) > 0 {
			start++
		} else {
			// this move is garbage
			moves[start] = moves[start+1]
			end--
		}
	}

	return end
}
