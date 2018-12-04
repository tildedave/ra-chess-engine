package main

import (
	"fmt"
	"math/bits"
)

const (
	NORTH      = iota
	SOUTH      = iota
	WEST       = iota
	EAST       = iota
	NORTH_WEST = iota
	NORTH_EAST = iota
	SOUTH_EAST = iota
	SOUTH_WEST = iota
)

func CreateRay(col byte, row byte, direction int) uint64 {
	var bitboard uint64
	first := false
	for row >= 0 && col >= 0 && row <= 8 && col <= 8 {
		if !first {
			bitboard = SetBitboard(bitboard, row*8+col)
		}
		switch direction {
		case NORTH:
			row++
		case SOUTH:
			row--
		case WEST:
			col++
		case EAST:
			col--
		case NORTH_EAST:
			row++
			col--
		case NORTH_WEST:
			row++
			col++
		case SOUTH_EAST:
			row--
			col--
		case SOUTH_WEST:
			row--
			col++
		}
	}

	fmt.Println(BitboardToString(bitboard))

	return bitboard
}

func AllCombinations(bitboard uint64) []uint64 {
	// flip every bit in this board
	if bitboard == 0 {
		return []uint64{}
	}

	combinations := make([]uint64, 0)
	firstSet := bits.LeadingZeros64(bitboard)
	fmt.Println(firstSet)
	rbb := bitboard & (BITBOARD_ALL_ONES ^ (1 << uint(firstSet)))
	fmt.Println(BitboardToString(bitboard))
	fmt.Println(BitboardToString(rbb))
	recursiveCombinations := AllCombinations(rbb)
	combinations = append(combinations, recursiveCombinations...)
	for i := range recursiveCombinations {
		combinations = append(combinations, recursiveCombinations[i]|(1<<uint(firstSet)))
	}

	return combinations
}

func GenerateMagicBitboards() {
	for col := byte(0); col < 8; col++ {
		for row := byte(0); row < 8; row++ {
			for _, b := range AllCombinations(CreateRay(row, col, NORTH)) {
				fmt.Println(BitboardToString(b))
			}
			// // go in all four directions
			// for down := 0; int(row)-down >= 0; down++ {
			// 	for up := 0; int(row)+up < 8; up++ {
			// 		for left := 0; int(col)-left >= 0; left++ {
			// 			for right := 0; int(col)+right <= 8; right++ {

			// 				// up/down/left/right is the bounding box

			// 				combinationsDown := 1<<uint(row) - down
			// 				combinationsUp := 8 - (1<<uint(row) + up)

			// 				fmt.Printf("start: %d, up: %d (%d) down: %d (%d) left: %d right: %d\n",
			// 					row*8+col,
			// 					up,
			// 					combinationsUp,
			// 					down,
			// 					combinationsDown,
			// 					left,
			// 					right)

			// 				var bitboard uint64
			// 				bitboard = SetBitboard(bitboard, idx(col-byte(left), row))
			// 				bitboard = SetBitboard(bitboard, idx(col+byte(right), row))
			// 				bitboard = SetBitboard(bitboard, idx(col, row-byte(down)))
			// 				bitboard = SetBitboard(bitboard, idx(col, row+byte(up)))
			// 				fmt.Println(bitboard)
			// 				fmt.Println(BitboardToString(bitboard))
			// 			}
			// 		}
			// 	}
			// }
		}
	}
}

func idx(col byte, row byte) byte {
	return row*8 + col
}
