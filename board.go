package main

import "fmt"

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

func pieceAtSquare(board []byte, row byte, col byte) byte {
	// row is 0 - 7, col is 0 - 8
	// 10x12 board
	return board[20+row*10+1+col]
}

func toString(p byte) byte {
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

func boardToString(board []byte) string {
	var s [9 * 8]byte

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			fmt.Println(i, j)
			s[i*9+j] = toString(pieceAtSquare(board, byte(i), byte(j)))
		}
		s[i*9+8] = '\n'
	}
	return string(s[:9*8])

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

func main() {
	fmt.Println("Hello, world")
}
