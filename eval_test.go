package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _ = fmt.Println

func TestEvalEmptyBoard(t *testing.T) {
	testBoard := CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_H8, BLACK_MASK|KING_MASK)

	boardEval := Eval(&testBoard)

	assert.Equal(t, 0, boardEval.material)
}

func TestEvalPawn(t *testing.T) {
	testBoard := CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A1, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_H8, BLACK_MASK|KING_MASK)

	boardEval := Eval(&testBoard)

	assert.Equal(t, 100, boardEval.material)
}

func TestEvalPawnAgainstBishop(t *testing.T) {
	testBoard := CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A1, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A3, BLACK_MASK|BISHOP_MASK)
	testBoard.SetPieceAtSquare(SQUARE_H8, BLACK_MASK|KING_MASK)

	boardEval := Eval(&testBoard)

	assert.Equal(t, -230, boardEval.material)
}

func TestEvalPassedPawns(t *testing.T) {
	testBoard := CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A1, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_H8, BLACK_MASK|KING_MASK)

	boardEval := Eval(&testBoard)
	fmt.Println(boardEval)
}

func TestEvalStartingPosition(t *testing.T) {
	testBoard := CreateInitialBoardState()
	boardEval := Eval(&testBoard)

	assert.Equal(t, boardEval.material, 0)
}

func TestEvalStartingPositionCenterControl(t *testing.T) {
	testBoard := CreateInitialBoardState()
	testBoard.SetPieceAtSquare(SQUARE_E2, 0x00)
	testBoard.SetPieceAtSquare(SQUARE_E4, WHITE_MASK|PAWN_MASK)
	boardEval := Eval(&testBoard)

	assert.Equal(t, boardEval.material, 0)
}

func TestEvalBlockedPawns(t *testing.T) {
	testBoard := CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A1, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_B2, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A3, WHITE_MASK|KNIGHT_MASK)
	testBoard.SetPieceAtSquare(SQUARE_H8, BLACK_MASK|KING_MASK)
	// blocked
	testBoard.SetPieceAtSquare(SQUARE_H7, BLACK_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_H6, BLACK_MASK|BISHOP_MASK)
	// not blocked
	testBoard.SetPieceAtSquare(SQUARE_G7, BLACK_MASK|PAWN_MASK)
	entry := GetPawnTableEntry(&testBoard)

	eval := createEvalBitboards(&testBoard, entry)

	assert.True(t, IsBitboardSet(eval.blockedPawns[WHITE_OFFSET], SQUARE_A2))
	assert.True(t, IsBitboardSet(eval.blockedPawns[BLACK_OFFSET], SQUARE_H7))

	assert.False(t, IsBitboardSet(eval.blockedPawns[WHITE_OFFSET], SQUARE_B2))
	assert.False(t, IsBitboardSet(eval.blockedPawns[BLACK_OFFSET], SQUARE_G7))
}

func TestEvalKingSafety(t *testing.T) {
	testBoard := CreateInitialBoardState()

	testBoard.SetPieceAtSquare(SQUARE_E1, 0x00)
	testBoard.SetPieceAtSquare(SQUARE_F1, WHITE_MASK|ROOK_MASK)
	testBoard.SetPieceAtSquare(SQUARE_G1, WHITE_MASK|KING_MASK)

	// 3 pawns should mean okay
	boardEval := Eval(&testBoard)
	assert.Equal(t, KING_PAWN_COVER_SCORE[3], boardEval.pieceScore[WHITE_OFFSET][KING_MASK])

	testBoard.SetPieceAtSquare(SQUARE_F2, 0x00)
	boardEval = Eval(&testBoard)
	assert.Equal(t, KING_PAWN_COVER_SCORE[2], boardEval.pieceScore[WHITE_OFFSET][KING_MASK])

	testBoard.SetPieceAtSquare(SQUARE_G2, 0x00)
	boardEval = Eval(&testBoard)
	assert.Equal(t, KING_PAWN_COVER_SCORE[1], boardEval.pieceScore[WHITE_OFFSET][KING_MASK])
}
