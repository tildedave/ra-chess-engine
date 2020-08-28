package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _ = fmt.Println

func TestNotInCheck(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|KING_MASK)

	assert.False(t, testBoard.IsInCheck(WHITE_OFFSET))
}

func TestInCheckBishop(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_G8, BLACK_MASK|BISHOP_MASK)

	assert.True(t, testBoard.IsInCheck(WHITE_OFFSET))
}

func TestInCheckBishopAsBlack(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	// TODO - replace when we parse FEN into board state
	testBoard.SetPieceAtSquare(SQUARE_A2, BLACK_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_G8, WHITE_MASK|BISHOP_MASK)

	assert.True(t, testBoard.IsInCheck(BLACK_OFFSET))
}

func TestInCheckRook(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A8, BLACK_MASK|ROOK_MASK)

	assert.True(t, testBoard.IsInCheck(WHITE_OFFSET))
}

func TestInCheckKnight(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_B4, BLACK_MASK|KNIGHT_MASK)

	assert.True(t, testBoard.IsInCheck(WHITE_OFFSET))
}

func TestInCheckPawn(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_B3, BLACK_MASK|PAWN_MASK)

	assert.True(t, testBoard.IsInCheck(WHITE_OFFSET))
}

func TestBlackKingDoesNotCheckWhiteKing(t *testing.T) {
	var fen string = "r6r/1b2k1bq/8/8/7B/8/8/R3K2R b QK - 3 2"
	testBoard, _ := CreateBoardStateFromFENString(fen)

	testBoard.ApplyMove(CreateMove(SQUARE_H7, SQUARE_H4))

	assert.False(t, testBoard.IsInCheck(BLACK_OFFSET))
}

func TestKingsCheckEachOther(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A3, BLACK_MASK|KING_MASK)

	assert.True(t, testBoard.IsInCheck(WHITE_OFFSET))
	assert.True(t, testBoard.IsInCheck(BLACK_OFFSET))
}

func TestEnPassantRemovesCheck(t *testing.T) {
	var fen string = "8/8/8/2k5/2pP4/8/B7/4K3 b - d3 5 3"
	testBoard, _ := CreateBoardStateFromFENString(fen)

	testBoard.ApplyMove(CreateEnPassantCapture(SQUARE_C4, SQUARE_D3))

	assert.False(t, testBoard.IsInCheck(BLACK_OFFSET))
}

func TestFilterChecks(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A1, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A3, BLACK_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_H2, WHITE_MASK|QUEEN_MASK)

	moves := make([]Move, 64)
	end := GenerateMoves(&testBoard, moves, 0)
	checks := testBoard.FilterChecks(moves[0:end])

	assertMovePresent(t, checks, SQUARE_H2, SQUARE_H3)
	assertMovePresent(t, checks, SQUARE_H2, SQUARE_G3)
	assertMovePresent(t, checks, SQUARE_H2, SQUARE_D6)
	assertMovePresent(t, checks, SQUARE_H2, SQUARE_B2)
}

func TestKingAttack(t *testing.T) {
	boardState, _ := CreateBoardStateFromFENString("8/8/5k2/4K3/8/8/8/8 w - - 0 2")

	assert.True(t, boardState.IsInCheck(WHITE_OFFSET))
	assert.True(t, boardState.IsInCheck(BLACK_OFFSET))
}
