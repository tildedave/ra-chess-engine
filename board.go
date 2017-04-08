package main

import (
	"fmt"
	"strconv"
)

const SENTINEL_MASK byte = 0xFF
const PAWN_MASK byte = 0x01
const KNIGHT_MASK byte = 0x02
const BISHOP_MASK byte = 0x03
const ROOK_MASK byte = 0x04
const QUEEN_MASK byte = 0x05
const KING_MASK byte = 0x06
const BLACK_MASK byte = 0x80
const WHITE_MASK byte = 0x00

var m map[byte]byte = make(map[byte]byte)

func isPieceBlack(p byte) bool {
	return p&BLACK_MASK == BLACK_MASK
}

func isPieceWhite(p byte) bool {
	return p&WHITE_MASK == WHITE_MASK
}

func isSentinel(p byte) bool {
	return p&SENTINEL_MASK == SENTINEL_MASK
}

func isPawn(p byte) bool {
	return p&PAWN_MASK == PAWN_MASK
}

func isBishop(p byte) bool {
	return p&BISHOP_MASK == BISHOP_MASK
}

func isKnight(p byte) bool {
	return p&KNIGHT_MASK == KNIGHT_MASK
}

func isRook(p byte) bool {
	return p&ROOK_MASK == ROOK_MASK
}

func isQueen(p byte) bool {
	return p&QUEEN_MASK == QUEEN_MASK
}

func isKing(p byte) bool {
	return p&KING_MASK == KING_MASK
}

func isSquareEmpty(p byte) bool {
	return p == 0x00
}

func PieceAtSquare(boardState BoardState, sq uint8) byte {
	// row is 0 - 7, col is 0 - 8
	// 10x12 board
	return boardState.board[sq]
}

func pieceToString(p byte) byte {
	if p == 0x00 {
		return '.'
	} else if p == PAWN_MASK|WHITE_MASK {
		return 'P'
	} else if p == KNIGHT_MASK|WHITE_MASK {
		return 'N'
	} else if p == BISHOP_MASK|WHITE_MASK {
		return 'B'
	} else if p == ROOK_MASK|WHITE_MASK {
		return 'R'
	} else if p == QUEEN_MASK|WHITE_MASK {
		return 'Q'
	} else if p == KING_MASK|WHITE_MASK {
		return 'K'
	} else if p == PAWN_MASK|BLACK_MASK {
		return 'p'
	} else if p == KNIGHT_MASK|BLACK_MASK {
		return 'n'
	} else if p == BISHOP_MASK|BLACK_MASK {
		return 'b'
	} else if p == ROOK_MASK|BLACK_MASK {
		return 'r'
	} else if p == QUEEN_MASK|BLACK_MASK {
		return 'q'
	} else if p == KING_MASK|BLACK_MASK {
		return 'k'
	}

	return '-'
}

func BoardToString(boardState BoardState) string {
	var s [9 * 8]byte

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			var p = PieceAtSquare(boardState, RowAndColToSquare(byte(i), byte(j)))
			s[(7-i)*9+j] = pieceToString(p)
		}
		s[(7-i)*9+8] = '\n'
	}
	return string(s[:9*8])
}

func BoardStateToFENString(boardState BoardState) string {
	var s string

	for i := 0; i < 8; i++ {
		var numEmpty = 0
		for j := 0; j < 8; j++ {
			p := PieceAtSquare(boardState, RowAndColToSquare(byte(7-i), byte(j)))
			if isSquareEmpty(p) {
				numEmpty++
			} else {
				if numEmpty > 0 {
					s += strconv.Itoa(numEmpty)
					numEmpty = 0
				}
				s += string(pieceToString(p))
				numEmpty = 0
			}
		}
		if numEmpty > 0 {
			s += strconv.Itoa(numEmpty)
			numEmpty = 0
		}
		if i < 7 {
			s += "/"
		} else {
			s += " "
		}
	}

	if boardState.whiteToMove {
		s += "w "
	} else {
		s += "b "
	}

	var hasCastleSquare = false
	if boardState.whiteCanCastleKingside {
		s += "K"
		hasCastleSquare = true
	}
	if boardState.whiteCanCastleQueenside {
		s += "Q"
		hasCastleSquare = true
	}
	if boardState.blackCanCastleKingside {
		s += "k"
		hasCastleSquare = true
	}
	if boardState.blackCanCastleQueenside {
		s += "q"
		hasCastleSquare = true
	}
	if !hasCastleSquare {
		s += "-"
	}
	s += " "
	if boardState.enPassantTargetSquare == 255 {
		s += "-"
	} else {
		s += SquareToAlgebraicString(boardState.enPassantTargetSquare)
		// TODO: need to convert a square number (10x12) to algebraic notation (yawn)
	}
	s += " " + strconv.Itoa(boardState.halfmoveClock) + " " + strconv.Itoa(boardState.fullmoveNumber)

	return s
}

func RowAndColToSquare(row uint8, col uint8) uint8 {
	return 20 + row*10 + 1 + col
}

func SquareToAlgebraicString(sq uint8) string {
	var row = sq / 10
	var col = sq % 10

	if row < 2 || row > 9 {
		return "??"
	}
	if col == 0 || col == 9 {
		return "??"
	}

	// No need to offset this as 1 is board is 1-idnexed
	var rowStr = strconv.Itoa(int(col))
	switch row {
	case 2:
		return "a" + rowStr
	case 3:
		return "b" + rowStr
	case 4:
		return "c" + rowStr
	case 5:
		return "d" + rowStr
	case 6:
		return "e" + rowStr
	case 7:
		return "f" + rowStr
	case 8:
		return "g" + rowStr
	case 9:
		return "h" + rowStr
	default:
	}
	return "-"

}

// https://chessprogramming.wikispaces.com/10x12+Board
var initialBoard []byte = []byte{
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0x04, 0x02, 0x03, 0x05, 0x06, 0x03, 0x02, 0x04, 0xFF,
	0xFF, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0xFF,
	0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF,
	0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF,
	0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF,
	0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF,
	0xFF, 0x81, 0x81, 0x81, 0x81, 0x81, 0x81, 0x81, 0x81, 0xFF,
	0xFF, 0x84, 0x82, 0x83, 0x85, 0x86, 0x83, 0x82, 0x84, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
}

type BoardState struct {
	board                   []byte
	whiteToMove             bool
	whiteCanCastleKingside  bool
	whiteCanCastleQueenside bool
	blackCanCastleKingside  bool
	blackCanCastleQueenside bool
	enPassantTargetSquare   uint8
	// number of moves since last capture or pawn advance
	halfmoveClock int
	// starts at 1, incremented after Black moves
	fullmoveNumber int
}

func CreateInitialBoardState() BoardState {
	var b BoardState
	b.board = initialBoard
	b.whiteToMove = true
	b.whiteCanCastleKingside = true
	b.whiteCanCastleQueenside = true
	b.blackCanCastleKingside = true
	b.blackCanCastleQueenside = true
	b.enPassantTargetSquare = 255
	b.halfmoveClock = 0
	b.fullmoveNumber = 1

	return b
}

func main() {
	fmt.Println("Hello, world")
}
