package main

const BITBOARD_ALL_ONES = 0xFFFFFFFFFFFFFFFF
const BITBOARD_ALL_ZEROS = 0

var BITBOARD_PIECES = [6]byte{1, 2, 3, 4, 5, 6}

type Bitboards struct {
	color [2]uint64 // 0 = WHITE, 1 = BLACK
	piece [7]uint64 // 0 = unused, 1 = PAWN, 2 = KNIGHT, 3 = BISHOP, 4 = ROOK, 5 = QUEEN, 6 = KING
}

func BitboardsToString(b Bitboards) string {
	str := ""
	str += "white:\n" + BitboardToString(b.color[0])
	str += "black:\n" + BitboardToString(b.color[1])
	str += "pawns:\n" + BitboardToString(b.piece[1])
	str += "knights:\n" + BitboardToString(b.piece[2])
	str += "bishops:\n" + BitboardToString(b.piece[3])
	str += "rooks:\n" + BitboardToString(b.piece[4])
	str += "queens:\n" + BitboardToString(b.piece[5])
	str += "kings:\n" + BitboardToString(b.piece[6])

	return str
}

func SetBitboard(bitboard uint64, sq byte) uint64 {
	return (bitboard | (1 << sq))
}

func UnsetBitboard(bitboard uint64, sq byte) uint64 {
	return (bitboard & (0xFFFFFFFFFFFFFFFF ^ (1 << sq)))
}

func FlipBitboard(bitboard uint64, sq byte) uint64 {
	return (bitboard ^ (1 << sq))
}

func FlipBitboard2(bitboard uint64, sq1 byte, sq2 byte) uint64 {
	return (bitboard ^ ((1 << sq1) | (1 << sq2)))
}

func IsBitboardSet(bitboard uint64, sq byte) bool {
	return (bitboard & (1 << sq)) != 0
}

func BitboardToString(bitboard uint64) string {
	var s [9 * 8]byte

	for row := byte(0); row < 8; row++ {
		for col := byte(0); col < 8; col++ {
			var sq = row*8 + col
			var r byte
			if IsBitboardSet(bitboard, sq) {
				r = 'x'
			} else {
				r = '.'
			}
			s[(7-row)*9+col] = r
		}
		s[(7-row)*9+8] = '\n'
	}
	return string(s[:9*8]) + "\n"
}
