package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/bits"
	"math/rand"
	"time"
)

type CollisionEntry struct {
	set       bool
	moveBoard uint64
}

type Magic struct {
	Magic uint64 `json:magic`
	Sq    byte   `json:sq`
}

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

func FollowRay(bitboard uint64, col byte, row byte, direction int, distance int) uint64 {
	for distance > 0 {
		switch direction {
		case NORTH:
			row++
		case SOUTH:
			row--
		case WEST:
			col--
		case EAST:
			col++
		case NORTH_EAST:
			row++
			col++
		case NORTH_WEST:
			row++
			col--
		case SOUTH_EAST:
			row--
			col++
		case SOUTH_WEST:
			row--
			col--
		}

		if distance%2 == 1 {
			bitboard = SetBitboard(bitboard, row*8+col)
		}
		distance = distance >> 1
	}

	return bitboard
}

func RookMask(sq byte) uint64 {
	col := sq % 8
	row := sq / 8

	var bitboard uint64
	// extra bounds check on the negative values is to prevent overflows
	for wcol := col - 1; wcol > 0 && wcol < 8; wcol-- {
		bitboard = SetBitboard(bitboard, row*8+wcol)
	}
	for ecol := col + 1; ecol < 7; ecol++ {
		bitboard = SetBitboard(bitboard, row*8+ecol)
	}
	for nrow := row + 1; nrow < 7; nrow++ {
		bitboard = SetBitboard(bitboard, nrow*8+col)
	}
	for srow := row - 1; srow > 0 && srow < 8; srow-- {
		bitboard = SetBitboard(bitboard, srow*8+col)
	}

	return bitboard
}

func RookMoveBoard(sq byte, occupancies uint64) uint64 {
	col := sq % 8
	row := sq / 8

	var bitboard uint64
	for wcol := col - 1; wcol > 0 && wcol < 8; wcol-- {
		sq := row*8 + wcol
		if IsBitboardSet(occupancies, sq) {
			break
		}
		bitboard = SetBitboard(bitboard, sq)
	}
	for ecol := col + 1; ecol < 7; ecol++ {
		sq := row*8 + ecol
		if IsBitboardSet(occupancies, sq) {
			break
		}
		bitboard = SetBitboard(bitboard, sq)
	}
	for nrow := row + 1; nrow < 7; nrow++ {
		sq := nrow*8 + col
		if IsBitboardSet(occupancies, sq) {
			break
		}
		bitboard = SetBitboard(bitboard, sq)
	}
	for srow := row - 1; srow > 0 && srow < 8; srow-- {
		sq := srow*8 + col
		if IsBitboardSet(occupancies, sq) {
			break
		}
		bitboard = SetBitboard(bitboard, sq)
	}

	return bitboard
}

func BishopMask(sq byte) uint64 {
	col := sq % 8
	row := sq / 8

	var bitboard uint64
	for wcol, nrow := col-1, row+1; wcol > 0 && wcol < 8 && nrow < 7; wcol, nrow = wcol-1, nrow+1 {
		bitboard = SetBitboard(bitboard, nrow*8+wcol)
	}
	for ecol, nrow := col+1, row+1; ecol < 7 && nrow < 7; ecol, nrow = ecol+1, nrow+1 {
		bitboard = SetBitboard(bitboard, nrow*8+ecol)
	}
	for wcol, srow := col-1, row-1; wcol > 0 && wcol < 8 && srow > 0 && srow < 8; wcol, srow = wcol-1, srow-1 {
		bitboard = SetBitboard(bitboard, srow*8+wcol)
	}
	for ecol, srow := col+1, row-1; ecol < 7 && srow > 0 && srow < 8; ecol, srow = ecol+1, srow-1 {
		bitboard = SetBitboard(bitboard, srow*8+ecol)
	}

	return bitboard
}

func BishopMoveBoard(sq byte, occupancies uint64) uint64 {
	col := sq % 8
	row := sq / 8

	var bitboard uint64
	for wcol, nrow := col-1, row+1; wcol > 0 && wcol < 8 && nrow < 7; wcol, nrow = wcol-1, nrow+1 {
		sq := nrow*8 + wcol
		if IsBitboardSet(occupancies, sq) {
			break
		}
		bitboard = SetBitboard(bitboard, nrow*8+wcol)
	}
	for ecol, nrow := col+1, row+1; ecol < 7 && nrow < 7; ecol, nrow = ecol+1, nrow+1 {
		sq := nrow*8 + ecol
		if IsBitboardSet(occupancies, sq) {
			break
		}
		bitboard = SetBitboard(bitboard, sq)
	}
	for wcol, srow := col-1, row-1; wcol > 0 && wcol < 8 && srow > 0 && srow < 8; wcol, srow = wcol-1, srow-1 {
		sq := srow*8 + wcol
		if IsBitboardSet(occupancies, sq) {
			break
		}
		bitboard = SetBitboard(bitboard, sq)
	}
	for ecol, srow := col+1, row-1; ecol < 7 && srow > 0 && srow < 8; ecol, srow = ecol+1, srow-1 {
		sq := srow*8 + ecol
		if IsBitboardSet(occupancies, sq) {
			break
		}

		bitboard = SetBitboard(bitboard, sq)
	}

	return bitboard
}

func GenerateRookOccupancies(sq byte) []uint64 {
	col := sq % 8
	row := sq / 8
	occupancies := make([]uint64, 0)

	above := 8 - int(row) - 1
	below := int(row)
	west := col
	east := 8 - int(col) - 1
	for i := 0; i < 1<<(uint(above)-1); i++ {
		for j := 0; j < 1<<(uint(below)-1); j++ {
			for k := 0; k < 1<<(uint(west)-1); k++ {
				for l := 0; l < 1<<(uint(east)-1); l++ {
					var bitboard uint64
					bitboard = FollowRay(bitboard, col, row, NORTH, i)
					bitboard = FollowRay(bitboard, col, row, SOUTH, j)
					bitboard = FollowRay(bitboard, col, row, WEST, k)
					bitboard = FollowRay(bitboard, col, row, EAST, l)
					occupancies = append(occupancies, bitboard)
				}
			}
		}
	}

	return occupancies
}

func GenerateRookMagic(sq byte, r *rand.Rand) (uint64, int) {
	numBits := bits.OnesCount64(RookMask(sq))
	occupancies := GenerateRookOccupancies(sq)

	occupancyMoves := make(map[uint64]uint64)
	for _, occupancy := range occupancies {
		occupancyMoves[occupancy] = RookMoveBoard(sq, occupancy)
	}

	return TrialAndErrorMagic(sq, r, numBits, occupancies, occupancyMoves)
}

func GenerateBishopMagic(sq byte, r *rand.Rand) (uint64, int) {
	numBits := bits.OnesCount64(BishopMask(sq))
	occupancies := GenerateRookOccupancies(sq)

	occupancyMoves := make(map[uint64]uint64)
	for _, occupancy := range occupancies {
		occupancyMoves[occupancy] = BishopMoveBoard(sq, occupancy)
	}

	return TrialAndErrorMagic(sq, r, numBits, occupancies, occupancyMoves)
}

func TrialAndErrorMagic(
	sq byte,
	r *rand.Rand,
	numBits int,
	occupancies []uint64,
	occupancyMoves map[uint64]uint64,
) (uint64, int) {
	var candidate uint64
	total := 0

TrialAndError:
	for {
		// try candidates until one of them is
		collisionMap := make(map[uint]CollisionEntry)
		// overly bias zeros in the candidate - trick from https://github.com/goutham/magic-bits
		candidate = r.Uint64() & r.Uint64() & r.Uint64()
		total++

		collision := false
		for _, occupancy := range occupancies {
			hashKey := uint((occupancy * candidate) >> uint(64-numBits))
			moveBoard := occupancyMoves[occupancy]

			if !collisionMap[hashKey].set {
				collisionMap[hashKey] = CollisionEntry{moveBoard: moveBoard, set: true}
			} else if collisionMap[hashKey].moveBoard != moveBoard {
				collision = true
				break
			}
		}

		if !collision {
			break TrialAndError
		}
	}

	return candidate, total
}

// GenerateMagicBitboards generates the magic values for both rook and bishop squares.
// It then creates two files:
// - rook-magics.json
// - bishop-magics.json
// These files will be read on engine initialization for pre-computing sliding piece moves based
// on an occupancy map for a given square.
func GenerateMagicBitboards() error {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	rookMagics := make([]Magic, 64)
	bishopMagics := make([]Magic, 64)

	for row := byte(0); row < 8; row++ {
		for col := byte(0); col < 8; col++ {
			sq := idx(col, row)
			rookMagic, _ := GenerateRookMagic(idx(col, row), r)
			bishopMagic, _ := GenerateBishopMagic(idx(col, row), r)

			rookMagics[sq] = Magic{Sq: sq, Magic: rookMagic}
			bishopMagics[sq] = Magic{Sq: sq, Magic: bishopMagic}
		}
	}

	err := outputMagicFile(rookMagics, "rook-magics.json")
	if err != nil {
		return err
	}
	err = outputMagicFile(bishopMagics, "bishop-magics.json")
	if err != nil {
		return err
	}
	return nil
}

func outputMagicFile(magics []Magic, filename string) error {
	magicJSON, err := json.Marshal(magics)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, magicJSON, 0644)
	if err != nil {
		return err
	}
	fmt.Printf("Wrote magic numbers to %s\n", filename)

	return nil
}

func idx(col byte, row byte) byte {
	return row*8 + col
}
