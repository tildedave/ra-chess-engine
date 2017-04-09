package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var _ = fmt.Println

func filterMovesFrom(moves []Move, from uint8) []Move {
	var filteredMoves []Move
	for _, move := range moves {
		if move.from == SQUARE_A2 {
			filteredMoves = append(filteredMoves, move)
		}
	}

	return filteredMoves
}

func filterCaptures(moves []Move) []Move {
	var captures []Move

	for _, move := range moves {
		if move.IsCapture() {
			captures = append(captures, move)
		}
	}

	return captures
}

func TestMoveGenerationWorks(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.board[SQUARE_A2] = WHITE_MASK | KING_MASK
	testBoard.board[SQUARE_A3] = WHITE_MASK | PAWN_MASK
	testBoard.board[SQUARE_A1] = BLACK_MASK | PAWN_MASK

	moves := GenerateMoves(&testBoard)

	movesFromKing := filterMovesFrom(moves, SQUARE_A2)
	numCaptures := len(filterCaptures(moves))

	assert.Equal(t, 4, len(movesFromKing))
	assert.Equal(t, 1, numCaptures)
}

func TestMoveGenerationFromRook(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.board[SQUARE_A2] = WHITE_MASK | ROOK_MASK
	testBoard.board[SQUARE_A4] = WHITE_MASK | PAWN_MASK
	testBoard.board[SQUARE_G2] = BLACK_MASK | ROOK_MASK
	testBoard.board[SQUARE_A1] = BLACK_MASK | QUEEN_MASK

	moves := GenerateMoves(&testBoard)
	movesFromRook := filterMovesFrom(moves, SQUARE_A2)
	numCaptures := len(filterCaptures(moves))

	// total 8 moves: 2 captures, A3 (1 step move), B2, C2, D2, E2, F2
	assert.Equal(t, 8, len(movesFromRook))
	assert.Equal(t, 2, numCaptures)
}

func TestMoveGenerationFromQueen(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.board[SQUARE_A2] = WHITE_MASK | QUEEN_MASK

	moves := GenerateMoves(&testBoard)

	assert.Equal(t, 21, len(moves))
}

func TestMoveGenerationFromInitialBoard(t *testing.T) {
	var testBoard BoardState = CreateInitialBoardState()

	moves := GenerateMoves(&testBoard)
	assert.Equal(t, 20, len(moves))
}

func TestMoveGenerationFromPawn(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.board[SQUARE_A2] = WHITE_MASK | PAWN_MASK
	testBoard.board[SQUARE_B3] = BLACK_MASK | ROOK_MASK

	moves := GenerateMoves(&testBoard)
	assert.Equal(t, 3, len(moves))
}

func TestEnPassantCaptureFromPawn(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.board[SQUARE_A5] = BLACK_MASK | PAWN_MASK
	testBoard.board[SQUARE_B5] = WHITE_MASK | PAWN_MASK
	testBoard.boardInfo.enPassantTargetSquare = SQUARE_A6

	moves := GenerateMoves(&testBoard)

	assert.Equal(t, 2, len(moves))
	assert.Equal(t, 1, len(filterCaptures(moves)))
}
