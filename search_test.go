package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Checkmate is Queen C3 to C1
func CreateMateInOneBoard() BoardState {
	boardState := CreateEmptyBoardState()
	boardState.SetPieceAtSquare(SQUARE_A1, BLACK_MASK|KING_MASK)
	boardState.SetPieceAtSquare(SQUARE_D5, WHITE_MASK|QUEEN_MASK)
	boardState.SetPieceAtSquare(SQUARE_B4, WHITE_MASK|KNIGHT_MASK)
	boardState.SetPieceAtSquare(SQUARE_H8, WHITE_MASK|KING_MASK)
	generateBoardLookupInfo(&boardState)

	return boardState
}

func TestSearchStartingPosition(t *testing.T) {
	boardState := CreateInitialBoardState()

	originalKey := boardState.hashKey
	Search(&boardState, 4)

	assert.Equal(t, originalKey, boardState.hashKey)
}

func TestSearchMateInOne(t *testing.T) {
	boardState := CreateMateInOneBoard()

	result := Search(&boardState, 2)

	assert.Equal(t, CHECKMATE_SCORE, result.value)
	assert.Equal(t, Move{from: SQUARE_D5, to: SQUARE_A2}, result.move)
}

func TestSearchMateInOneBlack(t *testing.T) {
	boardState := CreateMateInOneBoard()
	FlipBoardColors(&boardState)

	result := Search(&boardState, 2)

	assert.Equal(t, -CHECKMATE_SCORE, result.value)
	assert.Equal(t, Move{from: SQUARE_D5, to: SQUARE_A2}, result.move)
}

func TestSearchAvoidMateInOne(t *testing.T) {
	boardState := CreateMateInOneBoard()
	boardState.whiteToMove = false
	boardState.SetPieceAtSquare(SQUARE_F5, BLACK_MASK|ROOK_MASK)

	result := Search(&boardState, 1)

	assert.True(t, result.move.IsCapture())
	assert.Equal(t, Move{from: SQUARE_F5, to: SQUARE_D5, flags: CAPTURE_MASK}, result.move)
}

func TestSearchWhiteForcesPawnPromotion(t *testing.T) {
	boardState := CreateEmptyBoardState()
	boardState.SetPieceAtSquare(SQUARE_B7, WHITE_MASK|PAWN_MASK)
	boardState.SetPieceAtSquare(SQUARE_B8, BLACK_MASK|KING_MASK)
	boardState.SetPieceAtSquare(SQUARE_B5, WHITE_MASK|KING_MASK)
	generateBoardLookupInfo(&boardState)

	result := Search(&boardState, 5)

	assert.Equal(t, QUEEN_EVAL_SCORE, result.value)
	assert.Equal(t, SQUARE_B5, result.move.from)
	assert.True(t, result.move.to == SQUARE_C6 || result.move.to == SQUARE_A6)
}

func TestSearchBlackStopsPromotion(t *testing.T) {
	boardState := CreateEmptyBoardState()
	boardState.SetPieceAtSquare(SQUARE_D5, WHITE_MASK|PAWN_MASK)
	boardState.SetPieceAtSquare(SQUARE_D6, BLACK_MASK|KING_MASK)
	boardState.SetPieceAtSquare(SQUARE_D3, WHITE_MASK|KING_MASK)
	generateBoardLookupInfo(&boardState)

	result := Search(&boardState, 5)

	assert.Equal(t, PAWN_EVAL_SCORE, result.value)
}

func TestSearchWhiteForcesPromotion(t *testing.T) {
	boardState := CreateEmptyBoardState()
	boardState.SetPieceAtSquare(SQUARE_D6, WHITE_MASK|KING_MASK)
	boardState.SetPieceAtSquare(SQUARE_D8, BLACK_MASK|KING_MASK)
	boardState.SetPieceAtSquare(SQUARE_D4, WHITE_MASK|PAWN_MASK)
	generateBoardLookupInfo(&boardState)

	result := Search(&boardState, 12)

	assert.Equal(t, QUEEN_EVAL_SCORE, result.value)
}

func TestSearchWhiteSavesKnightFromCapture(t *testing.T) {
	boardState, _ := CreateBoardStateFromFENString("rnbqkbnr/ppp1pppp/8/8/3p4/2N5/PPPPPPPP/1RBQKBNR w Kkq - 0 3")

	result := Search(&boardState, 2)

	assert.Equal(t, result.move.from, SQUARE_C3)
}

func TestDoesNotHangCheckmate(t *testing.T) {
	boardState, _ := CreateBoardStateFromFENString("5rk1/B3bppp/8/6P1/b1p1pP2/2P5/Pr4P1/R3K1NR w KQ - 0 24")

	result := Search(&boardState, 2)
	fmt.Println(SearchResultToString(result))
}
