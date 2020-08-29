package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLeastValuableAttacker(t *testing.T) {
	boardState := CreateEmptyBoardState()

	boardState.SetPieceAtSquare(SQUARE_D2, WHITE_MASK|PAWN_MASK)
	boardState.SetPieceAtSquare(SQUARE_D1, WHITE_MASK|KNIGHT_MASK)
	boardState.SetPieceAtSquare(SQUARE_G5, WHITE_MASK|BISHOP_MASK)
	boardState.SetPieceAtSquare(SQUARE_E8, WHITE_MASK|ROOK_MASK)
	boardState.SetPieceAtSquare(SQUARE_A7, WHITE_MASK|QUEEN_MASK)

	occupancies := boardState.GetAllOccupanciesBitboard()
	attackers := boardState.GetSquareAttackersBoard(occupancies, SQUARE_E3)

	expectedPieces := []byte{PAWN_MASK, KNIGHT_MASK, BISHOP_MASK, ROOK_MASK, QUEEN_MASK}
	for i, sq := range []byte{SQUARE_D2, SQUARE_D1, SQUARE_G5, SQUARE_E8, SQUARE_A7} {
		attackerBoard, piece := GetLeastValuableAttacker(&boardState, attackers)
		assert.Equal(t, attackerBoard, SetBitboard(0, sq))
		assert.Equal(t, piece, expectedPieces[i])
		attackers = UnsetBitboard(attackers, sq)
	}
	attackerBoard, piece := GetLeastValuableAttacker(&boardState, attackers)
	assert.Equal(t, attackerBoard, uint64(0))
	assert.Equal(t, piece, byte(0))
}

func TestStaticExchangeEvaluationSingleCapture(t *testing.T) {
	boardState := CreateEmptyBoardState()
	boardState.sideToMove = BLACK_OFFSET

	boardState.SetPieceAtSquare(SQUARE_E3, WHITE_MASK|PAWN_MASK)
	boardState.SetPieceAtSquare(SQUARE_E8, BLACK_MASK|ROOK_MASK)

	assert.Equal(t, MATERIAL_SCORE[PAWN_MASK], StaticExchangeEvaluation(&boardState, SQUARE_E3, ROOK_MASK, SQUARE_E8))
}
func TestStaticExchangeEvaluationGuardedCapture(t *testing.T) {
	boardState := CreateEmptyBoardState()
	boardState.sideToMove = BLACK_OFFSET

	boardState.SetPieceAtSquare(SQUARE_E3, WHITE_MASK|PAWN_MASK)
	boardState.SetPieceAtSquare(SQUARE_E8, BLACK_MASK|ROOK_MASK)
	boardState.SetPieceAtSquare(SQUARE_D2, WHITE_MASK|PAWN_MASK)
	boardState.SetPieceAtSquare(SQUARE_D8, WHITE_MASK|KNIGHT_MASK)

	assert.Equal(t, -MATERIAL_SCORE[ROOK_MASK]+MATERIAL_SCORE[PAWN_MASK],
		StaticExchangeEvaluation(&boardState, SQUARE_E3, ROOK_MASK, SQUARE_E8))

	assert.Equal(t, MATERIAL_SCORE[KNIGHT_MASK],
		StaticExchangeEvaluation(&boardState, SQUARE_D8, ROOK_MASK, SQUARE_E8))

	moves := make([]Move, 64)
	end := GenerateMoveListing(&boardState, &moves, 0, false)
	end = boardState.FilterSEECaptures(&moves, 0, end)
	assert.Equal(t, 0, end)
}

func TestStaticExchangeEvaluationFromFENString(t *testing.T) {
	boardState, _ := CreateBoardStateFromFENString("1k1r4/1pp4p/p7/4p3/8/P5P1/1PP4P/2K1R3 w - -")

	assert.Equal(t, MATERIAL_SCORE[PAWN_MASK], StaticExchangeEvaluation(&boardState, SQUARE_E5, ROOK_MASK, SQUARE_E1))

}

func TestStaticExchangeEvaluationFromFENString2(t *testing.T) {
	boardState, _ := CreateBoardStateFromFENString("1k1r3q/1ppn3p/p4b2/4p3/8/P2N2P1/1PP1R1BP/2K1Q3 w - -")

	assert.Equal(t, -200, StaticExchangeEvaluation(&boardState, SQUARE_E5, KNIGHT_MASK, SQUARE_D3))
}

func TestStaticExchangeEvaluationStackOverflowQuestion(t *testing.T) {
	// https://stackoverflow.com/questions/44036416/chess-quiescence-search-dominating-runtime
	boardState, _ := CreateBoardStateFromFENString("rnbqkbnr/pppppppp/8/8/8/8/1PP1PPP1/RNBQKBNR w KQkq - 0 1")

	assert.Equal(t, -700, StaticExchangeEvaluation(&boardState, SQUARE_D7, QUEEN_MASK, SQUARE_E1))
	assert.Equal(t, -400, StaticExchangeEvaluation(&boardState, SQUARE_A7, ROOK_MASK, SQUARE_A1))
	assert.Equal(t, -400, StaticExchangeEvaluation(&boardState, SQUARE_H7, ROOK_MASK, SQUARE_H1))

	moves := make([]Move, 64)
	end := GenerateMoveListing(&boardState, &moves, 0, false)
	end = boardState.FilterSEECaptures(&moves, 0, end)
	assert.Equal(t, 0, end)
}
