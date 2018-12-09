package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _ = fmt.Println

func filterMovesFrom(moves []Move, from uint8) []Move {
	var filteredMoves []Move
	for _, move := range moves {
		if move.from == from {
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
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A3, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A1, BLACK_MASK|PAWN_MASK)

	moves := GenerateMoves(&testBoard)

	movesFromKing := filterMovesFrom(moves, SQUARE_A2)
	numCaptures := len(filterCaptures(moves))

	assert.Equal(t, 4, len(movesFromKing))
	assert.Equal(t, 1, numCaptures)
}

func TestMoveGenerationFromRook(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|ROOK_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A4, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_G2, BLACK_MASK|ROOK_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A1, BLACK_MASK|QUEEN_MASK)

	moves := GenerateMoves(&testBoard)
	movesFromRook := filterMovesFrom(moves, SQUARE_A2)
	numCaptures := len(filterCaptures(moves))

	// total 8 moves: 2 captures, A3 (1 step move), B2, C2, D2, E2, F2
	assert.Equal(t, 8, len(movesFromRook))
	assert.Equal(t, 2, numCaptures)
}

func TestMoveGenerationFromQueen(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|QUEEN_MASK)

	moves := GenerateMoves(&testBoard)

	assert.Equal(t, 21, len(moves))
}

func TestMoveGenerationFromBishop(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_D2, WHITE_MASK|BISHOP_MASK)

	moves := GenerateMoves(&testBoard)

	assert.Equal(t, 9, len(moves))
}

func TestMoveGenerationFromInitialBoard(t *testing.T) {
	var testBoard BoardState = CreateInitialBoardState()

	moves := GenerateMoves(&testBoard)
	assert.Equal(t, 20, len(moves))
}

func TestMoveGenerationFromPawn(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_B3, BLACK_MASK|ROOK_MASK)

	moves := GenerateMoves(&testBoard)
	assert.Equal(t, 3, len(moves))
}

func TestMoveGenerationFromPawnPromotion(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A7, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_B7, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_C7, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_D7, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_E7, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_F7, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_G7, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_H7, WHITE_MASK|PAWN_MASK)

	moves := GenerateMoves(&testBoard)
	assert.Equal(t, 32, len(moves))
}
func TestMoveGenerationFromPawnPromotionBlack(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, BLACK_MASK|PAWN_MASK)
	testBoard.whiteToMove = false

	moves := GenerateMoves(&testBoard)
	assert.Equal(t, 4, len(moves))
}

func TestMoveGenerationFromPawnDoesNotMoveIntoOwnPiece(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A4, WHITE_MASK|ROOK_MASK)

	moves := GenerateMoves(&testBoard)
	pawnMoves := filterMovesFrom(moves, SQUARE_A2)

	assert.Equal(t, 1, len(pawnMoves))
}

func TestEnPassantCaptureFromPawn(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A5, BLACK_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_B5, WHITE_MASK|PAWN_MASK)
	testBoard.boardInfo.enPassantTargetSquare = SQUARE_A6

	moves := GenerateMoves(&testBoard)

	assert.Equal(t, 2, len(moves))
	assert.Equal(t, 1, len(filterCaptures(moves)))
}

func TestCreateMoveBitboards(t *testing.T) {
	moveBitboards := CreateMoveBitboards()
	assert.Equal(t, uint64(0x2020000), moveBitboards.pawnMoves[WHITE_OFFSET][BB_SQUARE_B2])
	assert.Equal(t, uint64(0x50000), moveBitboards.pawnAttacks[WHITE_OFFSET][BB_SQUARE_B2])

	assert.Equal(t, uint64(0x1000000), moveBitboards.pawnMoves[WHITE_OFFSET][BB_SQUARE_A3])
	assert.Equal(t, uint64(0x2000000), moveBitboards.pawnAttacks[WHITE_OFFSET][BB_SQUARE_A3])

	assert.Equal(t, uint64(0x80800000), moveBitboards.pawnMoves[WHITE_OFFSET][BB_SQUARE_H2])
	assert.Equal(t, uint64(0x400000), moveBitboards.pawnAttacks[WHITE_OFFSET][BB_SQUARE_H2])
}
