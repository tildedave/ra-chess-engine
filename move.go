package main

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

func MoveToPrettyString(move Move, boardState BoardState) string {
	if move.flags&SPECIAL1_MASK == SPECIAL1_MASK && !move.IsCapture() {
		if move.flags&SPECIAL2_MASK == SPECIAL2_MASK {
			return "O-O-O"
		}
		return "O-O"
	}

	var p byte = boardState.board[move.from]
	if p&0x0F == PAWN_MASK {
		if move.IsCapture() {
			return SquareToAlgebraicString(move.from) + "x" + SquareToAlgebraicString(move.to)
		}

		return SquareToAlgebraicString(move.to)
	}

	// TODO: handle ambiguity if there's another piece of that type that
	// has a valid legal move here

	var s string
	s += string(pieceToString(p))

	if move.IsCapture() {
		s += "x"
	}
	s += SquareToAlgebraicString(move.to)

	return s
}

func MoveArrayToPrettyString(moveArr []Move, boardState BoardState) string {
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
