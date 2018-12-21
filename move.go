package main

import (
	"errors"
	"fmt"
	"strings"
)

// Move will be encoded as 32 bits - should still be fast
// Eventually consider trying a smaller representation
type Move struct {
	from  uint8
	to    uint8
	flags uint8
}

const CAPTURE_MASK = 0x80
const PROMOTION_MASK = 0x40 // may not be needed
const SPECIAL1_MASK = 0x20
const SPECIAL2_MASK = 0x10

func (m Move) IsCapture() bool {
	return m.flags&CAPTURE_MASK == CAPTURE_MASK
}

func (m Move) IsQueensideCastle() bool {
	return m.flags == SPECIAL1_MASK|SPECIAL2_MASK
}

func (m Move) IsKingsideCastle() bool {
	return m.flags == SPECIAL1_MASK
}

func (m Move) IsCastle() bool {
	var flag = m.flags & 0xF0
	return flag == SPECIAL1_MASK || flag == SPECIAL1_MASK|SPECIAL2_MASK
}

func (m Move) IsPromotion() bool {
	var flag = m.flags & 0xF0
	return flag == PROMOTION_MASK || flag == PROMOTION_MASK|CAPTURE_MASK
}

func (m Move) IsEnPassantCapture() bool {
	var flag = m.flags & 0xF0
	return flag == CAPTURE_MASK|SPECIAL1_MASK
}

// GetPromotionPiece returns the piece that the promotion move will be returned to (colorless).
func (m Move) GetPromotionPiece() uint8 {
	return m.flags & 0x0F
}

func CreateMove(from uint8, to uint8) Move {
	var m Move
	m.from = from
	m.to = to

	return m
}

func CreateCapture(from uint8, to uint8) Move {
	var m Move
	m.from = from
	m.to = to
	m.flags |= CAPTURE_MASK

	return m
}

func CreatePromotionCapture(from uint8, to uint8, pieceMask uint8) Move {
	var m Move = CreateCapture(from, to)

	m.flags |= PROMOTION_MASK | pieceMask

	return m
}

func CreateEnPassantCapture(from uint8, to uint8) Move {
	var m Move = CreateCapture(from, to)
	m.flags |= SPECIAL1_MASK

	return m
}

func CreateKingsideCastle(from uint8, to uint8) Move {
	var m Move
	m.from = from
	m.to = to
	m.flags |= SPECIAL1_MASK

	return m
}

func CreateQueensideCastle(from uint8, to uint8) Move {
	var m Move
	m.from = from
	m.to = to
	m.flags |= SPECIAL1_MASK | SPECIAL2_MASK

	return m
}

func CreatePromotion(from uint8, to uint8, pieceMask uint8) Move {
	// Piece is stored in bottom half of the promotion
	var m Move
	m.from = from
	m.to = to
	m.flags |= PROMOTION_MASK | pieceMask

	return m
}

func MoveToString(move Move) string {
	if move.flags&SPECIAL1_MASK == SPECIAL1_MASK && !move.IsCapture() {
		if move.flags&SPECIAL2_MASK == SPECIAL2_MASK {
			return "O-O-O"
		}
		return "O-O"
	}

	var s string
	s += SquareToAlgebraicString(move.from)
	if move.IsCapture() {
		s += "x"
	} else {
		s += "-"
	}
	s += SquareToAlgebraicString(move.to)
	if move.IsPromotion() {
		s += "=" + string(pieceToString(move.flags|WHITE_MASK))
	}

	return s
}

func MoveToPrettyString(move Move, boardState *BoardState) string {
	if move.flags&SPECIAL1_MASK == SPECIAL1_MASK && !move.IsCapture() {
		if move.flags&SPECIAL2_MASK == SPECIAL2_MASK {
			return "O-O-O"
		}
		return "O-O"
	}

	var p byte = boardState.board[move.from]
	if p&0x0F == PAWN_MASK {
		if move.IsCapture() {
			return ColumnToAlgebraicNotation(move.from%8+1) + "x" + SquareToAlgebraicString(move.to)
		}

		s := SquareToAlgebraicString(move.to)
		if move.IsPromotion() {
			s += "=" + string(pieceToString(move.flags|WHITE_MASK))
		}
		return s
	}

	// TODO: handle ambiguity if there's another piece of that type that
	// has a valid legal move here

	var s string
	s += string(pieceToString((p & 0x0F) | WHITE_MASK))

	if move.IsCapture() {
		s += "x"
	}

	s += SquareToAlgebraicString(move.to)

	return s
}

func MoveArrayToPrettyString(moveArr []Move, boardState *BoardState) string {
	var s string
	for i, m := range moveArr {
		if i > 0 {
			s += " "
		}
		s += MoveToPrettyString(m, boardState)
		boardState.ApplyMove(m)
	}

	for i := len(moveArr) - 1; i >= 0; i-- {
		boardState.UnapplyMove(moveArr[i])
	}

	return s
}

func ParsePrettyMove(moveStr string, boardState *BoardState) (Move, error) {
	move := Move{}

	var err error
	var toSquare byte
	var isCapture bool
	var isPromotion bool
	var isKingsideCastle bool
	var isQueensideCastle bool
	var promotionPiece byte
	var piece byte

	// capture
	captureSplits := strings.Split(moveStr, "x")
	if len(captureSplits) == 2 {
		isCapture = true
		moveStr = strings.Replace(moveStr, "x", "", 1)
	} else if len(captureSplits) > 2 {
		return move, errors.New("String contained multiple captures")
	}

	// promotion
	promotionSplits := strings.Split(moveStr, "=")
	if len(promotionSplits) == 2 {
		isPromotion = true
		moveStr = promotionSplits[0]
		promotionPiece = CharToPieceMask(promotionSplits[1][1])
	}

	if moveStr == "O-O" {
		piece = KING_MASK
		isKingsideCastle = true
	} else if moveStr == "O-O-O" {
		piece = KING_MASK
		isQueensideCastle = true
	} else if len(moveStr) == 2 {
		piece = PAWN_MASK
		toSquare, err = ParseAlgebraicSquare(moveStr)
		if err != nil {
			return move, err
		}
	} else {
		piece = CharToPieceMask(moveStr[0])
		// TODO: handle disambiguating moves (Nbd2 R1f7 etc)

		toSquare, err = ParseAlgebraicSquare(moveStr[1:])
		if err != nil {
			return move, err
		}
	}

	for _, candidateMove := range GenerateMoves(boardState) {
		p := boardState.PieceAtSquare(candidateMove.from) & 0x0F
		if (candidateMove.to == toSquare || isKingsideCastle || isQueensideCastle) &&
			p == piece &&
			isCapture == candidateMove.IsCapture() &&
			isPromotion == candidateMove.IsPromotion() &&
			isKingsideCastle == candidateMove.IsKingsideCastle() &&
			isQueensideCastle == candidateMove.IsQueensideCastle() &&
			(!isPromotion || candidateMove.GetPromotionPiece() == promotionPiece) {
			return candidateMove, nil
		}
	}

	return move, fmt.Errorf("Could not find move %s in list of generated moves", moveStr)
}

func VariationToMoveList(variation string, boardState *BoardState) ([]Move, error) {
	moves := make([]Move, 0)
	var err error

	for _, moveStr := range strings.Split(variation, " ") {
		var move Move
		move, err = ParsePrettyMove(moveStr, boardState)

		if err != nil {
			break
		} else {
			moves = append(moves, move)
			boardState.ApplyMove(move)
		}
	}

	for i := len(moves) - 1; i >= 0; i-- {
		boardState.UnapplyMove(moves[i])
	}

	return moves, err
}
