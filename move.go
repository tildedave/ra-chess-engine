package main

import (
	"errors"
	"fmt"
	"strings"
)

// Move will be encoded as 32 bits - should still be fast
// 8 bits for "from"
// 8 bits for "to"
// 8 bits for masks
//
// In theory, we could do better, and stuff it in an int16.
// 6 bits are needed for "from" (0-63)
// 6 bits are needed for "to" (0-63)
// This leaves 4 bits left.  In this we need to store the normal flags
// for kingside/queenside castling and, for promotions, the piece that's being
// promoted to.
// We'll outline promotions first as it's simplest. There are 6 unique piece
// types, of which only 4 can be promoted to. So 2 bits for the promotion piece,
// understanding that we're no longer using KNIGHT_MASK &c (all the create and
// "which piece" logic needs to be changed).
// OK so this gives us 2 bits to encode the rest of the masks.
// Kingside/Queenside castling is a trinary state, castle (y/n), which type.
// 2 bits.  If we say we'll update the code that checks if a move is
// kingside/queenside to ALSO check if the king is the piece being moved, we can
// use these 2 bits for other flags.
// Pawn moves, en passant capture and promotion are 2 bits.  These flags also
// need to be checked based on the piece type.
// That's it.  Everything stored in 16 bits.
type Move uint32

const CAPTURE_MASK = 0x80
const PROMOTION_MASK = 0x40 // may not be needed
const SPECIAL1_MASK = 0x20
const SPECIAL2_MASK = 0x10

// IsCapture returns whether this move is a capture.
// Due to it checking the board array, it should not be used in any performance-critical places.
func (m Move) IsCapture(boardState *BoardState) bool {
	return m.IsEnPassantCapture() || boardState.board[m.To()] != EMPTY_SQUARE
}

func (m Move) IsQueensideCastle() bool {
	return m.Flags() == SPECIAL1_MASK|SPECIAL2_MASK
}

func (m Move) IsKingsideCastle() bool {
	return m.Flags() == SPECIAL1_MASK
}

func (m Move) IsCastle() bool {
	var flag = m.Flags() & 0xF0
	return flag == SPECIAL1_MASK || flag == SPECIAL1_MASK|SPECIAL2_MASK
}

func (m Move) IsPromotion() bool {
	var flag = m.Flags() & 0xF0
	return flag == PROMOTION_MASK || flag == PROMOTION_MASK|CAPTURE_MASK
}

func (m Move) IsEnPassantCapture() bool {
	var flag = m.Flags() & 0xF0
	return flag == CAPTURE_MASK|SPECIAL1_MASK
}

// GetPromotionPiece returns the piece that the promotion move will be returned to (colorless).
func (m Move) GetPromotionPiece() uint8 {
	return m.Flags() & 0x0F
}

func (m Move) From() uint8 {
	return uint8(m >> (32 - 8))
}

func (m Move) To() uint8 {
	return uint8((m >> (32 - 16)) & 0xFF)
}

func (m Move) Flags() uint8 {
	return uint8(m & 0xFF)
}

// Hopefully these all get inlined

func SetFrom(m Move, from uint8) Move {
	return Move(uint32(m) | uint32(from)<<(32-8))
}

func SetTo(m Move, to uint8) Move {
	return Move(uint32(m) | uint32(to)<<(32-16))
}

func SetFlags(m Move, flags uint8) Move {
	return Move(uint32(m) | uint32(flags))
}

func CreateMove(from uint8, to uint8) Move {
	var m Move
	m = SetFrom(m, from)
	m = SetTo(m, to)
	return m
}

func CreateMoveWithFlags(from uint8, to uint8, flags uint8) Move {
	var m Move
	m = SetFrom(m, from)
	m = SetTo(m, to)
	m = SetFlags(m, flags)
	return m
}

func CreatePromotionCapture(from uint8, to uint8, pieceMask uint8) Move {
	return CreateMoveWithFlags(from, to, PROMOTION_MASK|pieceMask)
}

func CreateEnPassantCapture(from uint8, to uint8) Move {
	return CreateMoveWithFlags(from, to, SPECIAL1_MASK|CAPTURE_MASK)
}

func CreateKingsideCastle(from uint8, to uint8) Move {
	return CreateMoveWithFlags(from, to, SPECIAL1_MASK)
}

func CreateQueensideCastle(from uint8, to uint8) Move {
	return CreateMoveWithFlags(from, to, SPECIAL1_MASK|SPECIAL2_MASK)
}

func CreatePromotion(from uint8, to uint8, pieceMask uint8) Move {
	// Piece is stored in bottom half of the promotion
	return CreateMoveWithFlags(from, to, PROMOTION_MASK|pieceMask)
}

func MoveToDebugString(move Move) string {
	return fmt.Sprintf("%s-%s (%d)",
		SquareToAlgebraicString(move.From()),
		SquareToAlgebraicString(move.To()),
		move.Flags())
}

func MoveToString(move Move, boardState *BoardState) string {
	if move.Flags()&SPECIAL1_MASK == SPECIAL1_MASK && !move.IsCapture(boardState) {
		if move.Flags()&SPECIAL2_MASK == SPECIAL2_MASK {
			return "O-O-O"
		}
		return "O-O"
	}

	var s string
	s += SquareToAlgebraicString(move.From())
	if move.IsCapture(boardState) {
		s += "x"
	} else {
		s += "-"
	}
	s += SquareToAlgebraicString(move.To())
	if move.IsPromotion() {
		s += "=" + string(pieceToString(move.Flags()|WHITE_MASK))
	}

	return s
}

func MoveToPrettyString(move Move, boardState *BoardState) string {
	if move.Flags()&SPECIAL1_MASK == SPECIAL1_MASK && !move.IsCapture(boardState) {
		if move.Flags()&SPECIAL2_MASK == SPECIAL2_MASK {
			return "O-O-O"
		}
		return "O-O"
	}

	var p byte = boardState.board[move.From()]
	if p&0x0F == PAWN_MASK {
		if move.IsCapture(boardState) {
			return ColumnToAlgebraicNotation(move.From()%8+1) + "x" + SquareToAlgebraicString(move.To())
		}

		s := SquareToAlgebraicString(move.To())
		if move.IsPromotion() {
			s += "=" + string(pieceToString(move.Flags()|WHITE_MASK))
		}
		return s
	}

	// TODO: handle ambiguity if there's another piece of that type that
	// has a valid legal move here

	var s string
	s += string(pieceToString((p & 0x0F) | WHITE_MASK))

	if move.IsCapture(boardState) {
		s += "x"
	}

	s += SquareToAlgebraicString(move.To())

	return s
}

func MoveArrayToPrettyString(moveArr []Move, boardState *BoardState) (string, error) {
	var s string
	var err error
	moves := make([]Move, 0)

	for _, m := range moveArr {
		if _, err = boardState.IsMoveLegal(m); err != nil {
			break
		}
		s += MoveToPrettyString(m, boardState) + " "
		boardState.ApplyMove(m)
		moves = append(moves, m)

		if boardState.IsCheckmate() {
			s = strings.Trim(s, " ") + "#"
		}
	}

	for i := len(moves) - 1; i >= 0; i-- {
		boardState.UnapplyMove(moves[i])
	}

	return strings.Trim(s, " "), err
}

func ParsePrettyMove(moveStr string, boardState *BoardState) (Move, error) {
	move := Move(0)

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
		promotionPiece = CharToPieceMask(promotionSplits[1][0])
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

	moves := make([]Move, 256)
	end := GenerateMoves(boardState, moves[:], 0)
	for _, candidateMove := range moves[0:end] {
		p := boardState.PieceAtSquare(candidateMove.From()) & 0x0F
		if (candidateMove.To() == toSquare || isKingsideCastle || isQueensideCastle) &&
			p == piece &&
			isCapture == candidateMove.IsCapture(boardState) &&
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
			move, err = ParseXboardMove(moveStr, boardState)
		}

		if err != nil {
			err = fmt.Errorf("Unable to parse %s into a move (attempted pretty move and xboard move)",
				moveStr)
			break
		}

		moves = append(moves, move)
		boardState.ApplyMove(move)
	}

	for i := len(moves) - 1; i >= 0; i-- {
		boardState.UnapplyMove(moves[i])
	}

	return moves, err
}
