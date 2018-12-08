package main

const BITBOARD_ALL_ONES = 0xFFFFFFFFFFFFFFFF
const BITBOARD_ALL_ZEROS = 0

type Bitboards struct {
	// 0 = WHITE, 1 = BLACK
	color [2]uint64
	// 0 = unused, 1 = PAWN, 2 = KNIGHT, 3 = BISHOP, 4 = ROOK, 5 = QUEEN, 6 = KING
	piece [7]uint64
}

func legacySquareToBitboardSquare(sq byte) byte {
	var row = sq / 10
	var col = sq % 10

	row = row - 2
	col = col - 1

	return row*8 + col
}

func SetBitboard(bitboard uint64, sq byte) uint64 {
	return (bitboard | (1 << sq))
}

func UnsetBitboard(bitboard uint64, sq byte) uint64 {
	return (bitboard & (0xFFFFFFFFFFFFFFFF ^ (1 << sq)))
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
