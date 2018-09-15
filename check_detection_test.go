package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var _ = fmt.Println

func TestNotInCheck(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|KING_MASK)
	testBoard.lookupInfo.whiteKingSquare = SQUARE_A2

	assert.False(t, testBoard.IsInCheck(true))
}

func TestInCheckBishop(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_G8, BLACK_MASK|BISHOP_MASK)
	testBoard.lookupInfo.whiteKingSquare = SQUARE_A2

	assert.True(t, testBoard.IsInCheck(true))
}

func TestInCheckBishopAsBlack(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	// TODO - replace when we parse FEN into board state
	testBoard.SetPieceAtSquare(SQUARE_A2, BLACK_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_G8, WHITE_MASK|BISHOP_MASK)
	testBoard.lookupInfo.blackKingSquare = SQUARE_A2

	assert.True(t, testBoard.IsInCheck(false))
}

func TestInCheckRook(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A8, BLACK_MASK|ROOK_MASK)
	testBoard.lookupInfo.whiteKingSquare = SQUARE_A2

	assert.True(t, testBoard.IsInCheck(true))
}

func TestInCheckKnight(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_B4, BLACK_MASK|KNIGHT_MASK)
	testBoard.lookupInfo.whiteKingSquare = SQUARE_A2

	assert.True(t, testBoard.IsInCheck(true))
}

func TestInCheckPawn(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_B3, BLACK_MASK|PAWN_MASK)
	testBoard.lookupInfo.whiteKingSquare = SQUARE_A2

	assert.True(t, testBoard.IsInCheck(true))
}

func TestBlackKingDoesNotCheckWhiteKing(t *testing.T) {
	var fen string = "r6r/1b2k1bq/8/8/7B/8/8/R3K2R b QK - 3 2"
	testBoard, _ := CreateBoardStateFromFENString(fen)

	testBoard.ApplyMove(CreateCapture(SQUARE_H7, SQUARE_H4))

	assert.False(t, testBoard.IsInCheck(false))
}

func TestKingsCheckEachOther(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A3, BLACK_MASK|KING_MASK)
	testBoard.lookupInfo.whiteKingSquare = SQUARE_A2
	testBoard.lookupInfo.blackKingSquare = SQUARE_A3

	assert.True(t, testBoard.IsInCheck(true))
	assert.True(t, testBoard.IsInCheck(false))
}

func TestEnPassantRemovesCheck(t *testing.T) {
	var fen string = "8/8/8/2k5/2pP4/8/B7/4K3 b - d3 5 3"
	testBoard, _ := CreateBoardStateFromFENString(fen)

	testBoard.ApplyMove(CreateCapture(SQUARE_C4, SQUARE_D3))

	assert.False(t, testBoard.IsInCheck(false))
}
