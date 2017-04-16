package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var _ = fmt.Println

func TestNotInCheck(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	// TODO - replace when we parse FEN into board state
	testBoard.board[SQUARE_A2] = WHITE_MASK | KING_MASK
	testBoard.lookupInfo.whiteKingSquare = SQUARE_A2

	assert.False(t, testBoard.IsInCheck(true))
}

func TestInCheckBishop(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	// TODO - replace when we parse FEN into board state
	testBoard.board[SQUARE_A2] = WHITE_MASK | KING_MASK
	testBoard.board[SQUARE_G8] = BLACK_MASK | BISHOP_MASK
	testBoard.lookupInfo.whiteKingSquare = SQUARE_A2

	assert.True(t, testBoard.IsInCheck(true))
}

func TestInCheckBishopAsBlack(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	// TODO - replace when we parse FEN into board state
	testBoard.board[SQUARE_A2] = BLACK_MASK | KING_MASK
	testBoard.board[SQUARE_G8] = WHITE_MASK | BISHOP_MASK
	testBoard.lookupInfo.blackKingSquare = SQUARE_A2

	assert.True(t, testBoard.IsInCheck(false))
}

func TestInCheckRook(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	// TODO - replace when we parse FEN into board state
	testBoard.board[SQUARE_A2] = WHITE_MASK | KING_MASK
	testBoard.board[SQUARE_A8] = BLACK_MASK | ROOK_MASK
	testBoard.lookupInfo.whiteKingSquare = SQUARE_A2

	assert.True(t, testBoard.IsInCheck(true))
}

func TestInCheckKnight(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.board[SQUARE_A2] = WHITE_MASK | KING_MASK
	testBoard.board[SQUARE_B4] = BLACK_MASK | KNIGHT_MASK
	testBoard.lookupInfo.whiteKingSquare = SQUARE_A2

	assert.True(t, testBoard.IsInCheck(true))
}

func TestInCheckPawn(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.board[SQUARE_A2] = WHITE_MASK | KING_MASK
	testBoard.board[SQUARE_B3] = BLACK_MASK | PAWN_MASK
	testBoard.lookupInfo.whiteKingSquare = SQUARE_A2

	assert.True(t, testBoard.IsInCheck(true))
}
