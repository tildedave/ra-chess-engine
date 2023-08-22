package main

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestCreateMove(t *testing.T) {
	var m = CreateMove(31, 51)

	assert.Equal(t, uintptr(4), unsafe.Sizeof(m))
	assert.Equal(t, CreateMove(31, 51), m)
}

func TestCreateCapturePromotion(t *testing.T) {
	var m = CreatePromotionCapture(SQUARE_A7, SQUARE_B8, QUEEN_MASK)
	fmt.Printf("move: %d hex: %x\n", m, m)

	assert.Equal(t, CreateMoveWithFlags(SQUARE_A7, SQUARE_B8, 0x45), m)
	assert.True(t, m.IsPromotion())
}

func TestCreateEnPassantCapture(t *testing.T) {
	var m = CreateEnPassantCapture(31, 51)

	assert.Equal(t, CreateMoveWithFlags(31, 51, 0xA0), m)
	assert.True(t, m.IsEnPassantCapture())
}

func TestCreateCastle(t *testing.T) {
	assert.True(t, CreateKingsideCastle(SQUARE_E1, SQUARE_G1).IsCastle())
	assert.True(t, CreateQueensideCastle(SQUARE_E1, SQUARE_C1).IsCastle())
}

func TestMoveToString(t *testing.T) {
	boardState := CreateEmptyBoardState()

	assert.Equal(t, "a2-a4", MoveToString(CreateMove(SQUARE_A2, SQUARE_A4), &boardState))
	boardState.board[SQUARE_A4] = WHITE_MASK | PAWN_MASK
	assert.Equal(t, "a2xa4", MoveToString(CreateMove(SQUARE_A2, SQUARE_A4), &boardState))
	assert.Equal(t, "a2xa4", MoveToString(CreateEnPassantCapture(SQUARE_A2, SQUARE_A4), &boardState))
	assert.Equal(t, "O-O", MoveToString(CreateKingsideCastle(25, 27), &boardState))
	assert.Equal(t, "O-O-O", MoveToString(CreateQueensideCastle(25, 23), &boardState))
}

func TestMoveToPrettyString(t *testing.T) {
	var boardState BoardState = CreateEmptyBoardState()

	boardState.SetPieceAtSquare(SQUARE_A2, WHITE_MASK|PAWN_MASK)
	boardState.SetPieceAtSquare(SQUARE_B4, BLACK_MASK|PAWN_MASK)
	boardState.SetPieceAtSquare(SQUARE_F5, WHITE_MASK|KNIGHT_MASK)
	boardState.SetPieceAtSquare(SQUARE_G7, BLACK_MASK|ROOK_MASK)

	assert.Equal(t, "axb4", MoveToPrettyString(CreateMove(SQUARE_A2, SQUARE_B4), &boardState))
	assert.Equal(t, "axb4", MoveToPrettyString(CreateEnPassantCapture(SQUARE_A2, SQUARE_B4), &boardState))
	assert.Equal(t, "a4", MoveToPrettyString(CreateMove(SQUARE_A2, SQUARE_A4), &boardState))
	assert.Equal(t, "Nxg7", MoveToPrettyString(CreateMove(SQUARE_F5, SQUARE_G7), &boardState))
}

func TestParsePrettyMoveFromInitial(t *testing.T) {
	var boardState BoardState = CreateInitialBoardState()

	move, err := ParsePrettyMove("Nf3", &boardState)
	assert.Nil(t, err)
	assert.Equal(t, CreateMove(SQUARE_G1, SQUARE_F3), move)

	move, err = ParsePrettyMove("e4", &boardState)
	assert.Nil(t, err)
	assert.Equal(t, CreateMove(SQUARE_E2, SQUARE_E4), move)

	move, err = ParsePrettyMove("e6", &boardState)
	assert.NotNil(t, err)
}

func TestParsePrettyMoveCastle(t *testing.T) {
	var boardState BoardState = CreateEmptyBoardState()
	boardState.boardInfo.whiteCanCastleKingside = true
	boardState.boardInfo.whiteCanCastleQueenside = true

	boardState.SetPieceAtSquare(SQUARE_E1, KING_MASK|WHITE_MASK)
	boardState.SetPieceAtSquare(SQUARE_A1, ROOK_MASK|WHITE_MASK)
	boardState.SetPieceAtSquare(SQUARE_H1, ROOK_MASK|WHITE_MASK)

	move, err := ParsePrettyMove("O-O", &boardState)
	assert.Nil(t, err)
	assert.True(t, move.IsKingsideCastle())

	move, err = ParsePrettyMove("O-O-O", &boardState)
	assert.Nil(t, err)
	assert.True(t, move.IsQueensideCastle())
}
