package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _ = fmt.Println

func TestProcessNewCommand(t *testing.T) {
	var state XboardState

	action, state := ProcessXboardCommand("new", state)

	assert.Equal(t, ACTION_HALT, action)
	assert.Equal(t, state.boardState.ToFENString(),
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
}

func TestProcessMoveCommand(t *testing.T) {
	var state XboardState
	var action int

	_, state = ProcessXboardCommand("new", state)
	action, state = ProcessXboardCommand("e2e4", state)

	assert.Equal(t, ACTION_THINK_AND_MOVE, action)
	assert.Equal(t, uint8(0), state.boardState.PieceAtSquare(SQUARE_E2))
	assert.Equal(t, WHITE_MASK|PAWN_MASK, state.boardState.PieceAtSquare(SQUARE_E4))
}

func TestProcessMoveCommandInForceMode(t *testing.T) {
	var state XboardState
	var action int

	_, state = ProcessXboardCommand("new", state)
	_, state = ProcessXboardCommand("force", state)
	action, state = ProcessXboardCommand("g1f3", state)

	assert.Equal(t, ACTION_NOTHING, action)
	assert.Equal(t, uint8(0), state.boardState.PieceAtSquare(SQUARE_G1))
	assert.Equal(t, WHITE_MASK|KNIGHT_MASK, state.boardState.PieceAtSquare(SQUARE_F3))
}

func TestProcessInvalidFENState(t *testing.T) {
	var state XboardState
	var action int

	action, state = ProcessXboardCommand("setboard thisisnotafenstring", state)
	assert.Equal(t, ACTION_ERROR, action)
	assert.NotNil(t, state.err)

	action, state = ProcessXboardCommand("e2e4", state)
	assert.Equal(t, ACTION_ERROR, action)
	assert.NotNil(t, state.err)
}

func TestProcessUndoCommand(t *testing.T) {
	var state XboardState
	var action int

	_, state = ProcessXboardCommand("new", state)
	_, state = ProcessXboardCommand("force", state)
	_, state = ProcessXboardCommand("e2e4", state)
	_, state = ProcessXboardCommand("e5e7", state)
	action, state = ProcessXboardCommand("undo", state)

	assert.Equal(t, BLACK_MASK|PAWN_MASK, state.boardState.PieceAtSquare(SQUARE_E7))
	assert.Equal(t, uint8(0), state.boardState.PieceAtSquare(SQUARE_E5))
	assert.Equal(t, BLACK_OFFSET, state.boardState.sideToMove)
	assert.Equal(t, ACTION_NOTHING, action)
}

func TestProcessRemoveCommand(t *testing.T) {
	var state XboardState
	var action int

	_, state = ProcessXboardCommand("new", state)
	_, state = ProcessXboardCommand("force", state)
	_, state = ProcessXboardCommand("e2e4", state)
	_, state = ProcessXboardCommand("e7e5", state)
	action, state = ProcessXboardCommand("remove", state)

	assert.Equal(t, BLACK_MASK|PAWN_MASK, state.boardState.PieceAtSquare(SQUARE_E7))
	assert.Equal(t, WHITE_MASK|PAWN_MASK, state.boardState.PieceAtSquare(SQUARE_E2))
	assert.Equal(t, uint8(0), state.boardState.PieceAtSquare(SQUARE_E5))
	assert.Equal(t, uint8(0), state.boardState.PieceAtSquare(SQUARE_E4))
	assert.Equal(t, WHITE_OFFSET, state.boardState.sideToMove)
	assert.Equal(t, ACTION_THINK_AND_MOVE, action)
}

func TestProcessResultCommand(t *testing.T) {
	var state XboardState
	var action int

	_, state = ProcessXboardCommand("new", state)
	action, state = ProcessXboardCommand("result 1/2-1/2 {Forgot how to play}", state)

	assert.Equal(t, ACTION_GAME_OVER, action)
}

func TestProcessNameCommand(t *testing.T) {
	var state XboardState
	var action int

	action, state = ProcessXboardCommand("name bob", state)

	assert.Equal(t, ACTION_NOTHING, action)
	assert.Equal(t, "bob", state.opponentName)
}

func TestParseXboardCommandSetboard(t *testing.T) {
	var state XboardState
	var action int

	fenString := "r6r/1b2k1bq/8/8/7B/8/8/R3K2R b KQ - 3 2"
	action, state = ProcessXboardCommand("setboard "+fenString, state)

	assert.Equal(t, ACTION_HALT, action)
	assert.Equal(t, fenString, state.boardState.ToFENString())

	// assert.True(t, move.IsCapture())
	// assert.Equal(t, SQUARE_E5, move.From())
	// assert.Equal(t, SQUARE_D6, move.To())
}

func TestParseXboardMoveMove(t *testing.T) {
	boardState := CreateInitialBoardState()
	move, err := ParseXboardMove("e2e4", &boardState)

	assert.Nil(t, err)
	assert.Equal(t, CreateMove(SQUARE_E2, SQUARE_E4), move)
}

func TestParseXboardMoveCapture(t *testing.T) {
	boardState := CreateEmptyBoardState()
	boardState.SetPieceAtSquare(SQUARE_A1, ROOK_MASK|WHITE_MASK)
	boardState.SetPieceAtSquare(SQUARE_A2, ROOK_MASK|BLACK_MASK)

	move, err := ParseXboardMove("a1a2", &boardState)

	assert.Nil(t, err)
	assert.Equal(t, CreateMove(SQUARE_A1, SQUARE_A2), move)
}

func TestParseXboardMovePromotion(t *testing.T) {
	boardState := CreateEmptyBoardState()
	boardState.SetPieceAtSquare(SQUARE_H7, PAWN_MASK|WHITE_MASK)

	move, err := ParseXboardMove("h7h8q", &boardState)

	assert.Nil(t, err)
	assert.True(t, move.IsPromotion())
	assert.Equal(t, move.From(), SQUARE_H7)
	assert.Equal(t, move.To(), SQUARE_H8)
	assert.Equal(t, move.Flags(), PROMOTION_MASK|QUEEN_MASK)
}

func TestParseXboardMoveRegexp(t *testing.T) {
	assert.True(t, moveRegexp.MatchString("e2e1q"))
	assert.True(t, moveRegexp.MatchString("e2e1"))
}

func TestParseXboardMovePromotionCapture(t *testing.T) {
	boardState := CreateEmptyBoardState()
	boardState.SetPieceAtSquare(SQUARE_H7, PAWN_MASK|WHITE_MASK)
	boardState.SetPieceAtSquare(SQUARE_G8, ROOK_MASK|BLACK_MASK)

	move, err := ParseXboardMove("h7g8r", &boardState)

	assert.Nil(t, err)
	assert.True(t, move.IsPromotion())
	assert.Equal(t, move.From(), SQUARE_H7)
	assert.Equal(t, move.To(), SQUARE_G8)
	assert.Equal(t, move.Flags(), PROMOTION_MASK|ROOK_MASK)
}

func TestParseXboardMoveKingsideCastle(t *testing.T) {
	boardState := CreateInitialBoardState()
	boardState.SetPieceAtSquare(SQUARE_F1, 0x00)
	boardState.SetPieceAtSquare(SQUARE_G1, 0x00)

	move, err := ParseXboardMove("e1g1", &boardState)

	assert.Nil(t, err)
	assert.True(t, move.IsKingsideCastle())
	assert.Equal(t, SQUARE_E1, move.From())
	assert.Equal(t, SQUARE_G1, move.To())
}

func TestParseXboardMoveQueensideCastle(t *testing.T) {
	boardState := CreateInitialBoardState()
	boardState.SetPieceAtSquare(SQUARE_B1, 0x00)
	boardState.SetPieceAtSquare(SQUARE_C1, 0x00)
	boardState.SetPieceAtSquare(SQUARE_D1, 0x00)
	boardState.sideToMove = BLACK_OFFSET

	move, err := ParseXboardMove("e8c8", &boardState)

	assert.Nil(t, err)
	assert.True(t, move.IsQueensideCastle())
	assert.Equal(t, SQUARE_E8, move.From())
	assert.Equal(t, SQUARE_C8, move.To())
}

func TestParseXboardMoveEnPassantCapture(t *testing.T) {
	boardState := CreateEmptyBoardState()
	boardState.SetPieceAtSquare(SQUARE_E5, WHITE_MASK|PAWN_MASK)
	boardState.SetPieceAtSquare(SQUARE_D5, BLACK_MASK|PAWN_MASK)
	boardState.boardInfo.enPassantTargetSquare = SQUARE_D6

	move, err := ParseXboardMove("e5d6", &boardState)

	assert.Nil(t, err)
	assert.True(t, move.IsEnPassantCapture())
	assert.Equal(t, SQUARE_E5, move.From())
	assert.Equal(t, SQUARE_D6, move.To())
}
