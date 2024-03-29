package main

import (
	"fmt"
	"math/bits"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _ = fmt.Println

func filterMovesFrom(moves []Move, from uint8) []Move {
	var filteredMoves []Move
	for _, move := range moves {
		if move.From() == from {
			filteredMoves = append(filteredMoves, move)
		}
	}

	return filteredMoves
}

func filterCaptures(moves []Move, boardState *BoardState) []Move {
	var captures []Move

	for _, move := range moves {
		if boardState.board[move.To()] != EMPTY_SQUARE || move.IsEnPassantCapture() {
			captures = append(captures, move)
		}
	}

	return captures
}

func generateMovesFromBoard(boardState *BoardState) []Move {
	moves := make([]Move, 64)
	end := GenerateMoves(boardState, moves[:], 0)
	return moves[:end]
}

func TestMoveGenerationWorks(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|KING_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A3, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A1, BLACK_MASK|PAWN_MASK)

	moves := generateMovesFromBoard(&testBoard)

	movesFromKing := filterMovesFrom(moves, SQUARE_A2)
	numCaptures := len(filterCaptures(moves, &testBoard))

	assert.Equal(t, 4, len(movesFromKing))
	assert.Equal(t, 1, numCaptures)
}

func TestMoveGenerationFromRook(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|ROOK_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A4, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_G2, BLACK_MASK|ROOK_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A1, BLACK_MASK|QUEEN_MASK)

	moves := generateMovesFromBoard(&testBoard)
	movesFromRook := filterMovesFrom(moves, SQUARE_A2)
	numCaptures := len(filterCaptures(moves, &testBoard))

	// total 8 moves: 2 captures, A3 (1 step move), B2, C2, D2, E2, F2
	assert.Equal(t, 8, len(movesFromRook))
	assert.Equal(t, 2, numCaptures)
}

func TestMoveGenerationFromQueen(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|QUEEN_MASK)

	moves := generateMovesFromBoard(&testBoard)

	assert.Equal(t, 21, len(moves))
}

func TestMoveGenerationFromBishop(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_D2, WHITE_MASK|BISHOP_MASK)

	moves := generateMovesFromBoard(&testBoard)

	assert.Equal(t, 9, len(moves))
}

func TestMoveGenerationFromInitialBoard(t *testing.T) {
	var testBoard BoardState = CreateInitialBoardState()

	moves := generateMovesFromBoard(&testBoard)
	assert.Equal(t, 20, len(moves))
}

func TestMoveGenerationFromPawn(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_B3, BLACK_MASK|ROOK_MASK)

	moves := generateMovesFromBoard(&testBoard)
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

	moves := generateMovesFromBoard(&testBoard)
	assert.Equal(t, 32, len(moves))
}
func TestMoveGenerationFromPawnPromotionBlack(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, BLACK_MASK|PAWN_MASK)
	testBoard.sideToMove = BLACK_OFFSET

	moves := generateMovesFromBoard(&testBoard)
	assert.Equal(t, 4, len(moves))
}

func TestMoveGenerationFromPawnDoesNotMoveIntoOwnPiece(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A4, WHITE_MASK|ROOK_MASK)

	moves := generateMovesFromBoard(&testBoard)
	pawnMoves := filterMovesFrom(moves, SQUARE_A2)

	assert.Equal(t, 1, len(pawnMoves))
}

func TestEnPassantCaptureFromPawn(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A5, BLACK_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_B5, WHITE_MASK|PAWN_MASK)
	testBoard.boardInfo.enPassantTargetSquare = SQUARE_A6

	moves := generateMovesFromBoard(&testBoard)

	assert.Equal(t, 2, len(moves))
	assert.Equal(t, 1, len(filterCaptures(moves, &testBoard)))
}

func TestCreateMoveBitboardsPawnAsserts(t *testing.T) {
	moveBitboards := CreateMoveBitboards()
	assert.Equal(t, uint64(0x2020000), moveBitboards.pawnMoves[WHITE_OFFSET][SQUARE_B2])
	assert.Equal(t, uint64(0x50000), moveBitboards.pawnAttacks[WHITE_OFFSET][SQUARE_B2])

	assert.Equal(t, uint64(0x1000000), moveBitboards.pawnMoves[WHITE_OFFSET][SQUARE_A3])
	assert.Equal(t, uint64(0x2000000), moveBitboards.pawnAttacks[WHITE_OFFSET][SQUARE_A3])

	assert.Equal(t, uint64(0x80800000), moveBitboards.pawnMoves[WHITE_OFFSET][SQUARE_H2])
	assert.Equal(t, uint64(0x400000), moveBitboards.pawnAttacks[WHITE_OFFSET][SQUARE_H2])
}

func TestGetKingMoveBitboard(t *testing.T) {
	assert.Equal(t, 8, bits.OnesCount64(getKingMoveBitboard(SQUARE_D4)))
	assert.Equal(t, 5, bits.OnesCount64(getKingMoveBitboard(SQUARE_A2)))
	assert.Equal(t, 5, bits.OnesCount64(getKingMoveBitboard(SQUARE_H2)))
	assert.Equal(t, 8, bits.OnesCount64(getKingMoveBitboard(SQUARE_B4)))
	assert.Equal(t, 8, bits.OnesCount64(getKingMoveBitboard(SQUARE_G4)))
	assert.Equal(t, SetBitboard(SetBitboard(SetBitboard(0, SQUARE_A2), SQUARE_B2), SQUARE_B1),
		getKingMoveBitboard(SQUARE_A1))
}

func TestCreateMoveBitboardsKnightAsserts(t *testing.T) {
	// Test edges by traversing the board "up" and "down"
	for i := byte(0); i < 7; i++ {
		var expectedNumber int
		switch i {
		case 0:
			fallthrough
		case 7:
			expectedNumber = 4
		case 1:
			fallthrough
		case 6:
			expectedNumber = 6
		default:
			expectedNumber = 8
		}
		assert.Equal(t, expectedNumber, bits.OnesCount64(getKnightMoveBitboard(idx(i, 3))))
		assert.Equal(t, expectedNumber, bits.OnesCount64(getKnightMoveBitboard(idx(3, i))))
	}
	assert.Equal(t, SetBitboard(SetBitboard(0, SQUARE_B3), SQUARE_C2),
		getKnightMoveBitboard(SQUARE_A1))
}

func TestCreateMovesFromBitboard(t *testing.T) {
	var bitboard uint64
	bitboard = SetBitboard(bitboard, SQUARE_A5)
	bitboard = SetBitboard(bitboard, SQUARE_B3)
	bitboard = SetBitboard(bitboard, SQUARE_D4)

	moves := make([]Move, 64)
	end := CreateMovesFromBitboard(SQUARE_D3, bitboard, moves[:], 0, 0)
	assert.Equal(t, 3, end)
	assertMovePresent(t, moves[0:end], SQUARE_D3, SQUARE_A5)
	assertMovePresent(t, moves[0:end], SQUARE_D3, SQUARE_B3)
	assertMovePresent(t, moves[0:end], SQUARE_D3, SQUARE_D4)
}

func TestGenerateQuiescentMoves(t *testing.T) {
	var testBoard BoardState = CreateEmptyBoardState()
	testBoard.SetPieceAtSquare(SQUARE_A5, WHITE_MASK|ROOK_MASK)
	testBoard.SetPieceAtSquare(SQUARE_B5, BLACK_MASK|PAWN_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A1, BLACK_MASK|ROOK_MASK)
	testBoard.SetPieceAtSquare(SQUARE_A8, BLACK_MASK|QUEEN_MASK)

	moves := make([]Move, 64)
	moveScores := make([]int16, len(moves))
	end := GenerateQuiescentMoves(&testBoard, moves[:], moveScores[:], 0)
	assert.Equal(t, 3, end)

	// verify ordering of captures
	assert.Equal(t, CreateMoveWithFlags(SQUARE_A5, SQUARE_A8, CAPTURE_MASK), moves[0])
	assert.Equal(t, CreateMoveWithFlags(SQUARE_A5, SQUARE_A1, CAPTURE_MASK), moves[1])
	assert.Equal(t, CreateMoveWithFlags(SQUARE_A5, SQUARE_B5, CAPTURE_MASK), moves[2])
}
