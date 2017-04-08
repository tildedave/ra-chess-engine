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
	return m.flags&(SPECIAL1_MASK|SPECIAL2_MASK) == SPECIAL1_MASK|SPECIAL2_MASK
}

func (m Move) IsKingsideCastle() bool {
	return m.flags == SPECIAL1_MASK
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

func CreatePromotion(from uint8, to uint8, piece uint8) Move {
	// Piece is stored in bottom half of the promotion
	var m Move
	m.from = from
	m.to = to
	m.flags |= PROMOTION_MASK | (piece & 0x0F)

	return m
}

func MoveToString(move Move) string {
	if move.flags&SPECIAL1_MASK == SPECIAL1_MASK {
		if move.flags&SPECIAL2_MASK == SPECIAL2_MASK {
			return "O-O-O"
		}
		return "O-O"
	}

	var s string
	s += SquareToAlgebraicString(move.from)
	if move.flags&CAPTURE_MASK == CAPTURE_MASK {
		s += "x"
	} else {
		s += "-"
	}
	s += SquareToAlgebraicString(move.to)

	return s
}
