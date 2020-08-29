package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/bits"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"time"
)

type CollisionEntry struct {
	set       bool
	moveBoard uint64
}

type Magic struct {
	Magic uint64 `json:"magic"`
	Sq    byte   `json:"sq"`
	Bits  uint   `json:"bits"`
	Mask  uint64 `json:"mask"`
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
	for wcol := col - 1; wcol >= 0 && wcol < 8; wcol-- {
		sq := row*8 + wcol
		bitboard = SetBitboard(bitboard, sq)
		if IsBitboardSet(occupancies, sq) {
			break
		}
	}
	for ecol := col + 1; ecol < 8; ecol++ {
		sq := row*8 + ecol
		bitboard = SetBitboard(bitboard, sq)
		if IsBitboardSet(occupancies, sq) {
			break
		}
	}
	for nrow := row + 1; nrow < 8; nrow++ {
		sq := nrow*8 + col
		bitboard = SetBitboard(bitboard, sq)
		if IsBitboardSet(occupancies, sq) {
			break
		}
	}
	for srow := row - 1; srow >= 0 && srow < 8; srow-- {
		sq := srow*8 + col
		bitboard = SetBitboard(bitboard, sq)
		if IsBitboardSet(occupancies, sq) {
			break
		}
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
	for wcol, nrow := col-1, row+1; wcol >= 0 && wcol < 8 && nrow < 8; wcol, nrow = wcol-1, nrow+1 {
		sq := nrow*8 + wcol
		bitboard = SetBitboard(bitboard, nrow*8+wcol)
		if IsBitboardSet(occupancies, sq) {
			break
		}
	}
	for ecol, nrow := col+1, row+1; ecol < 8 && nrow < 8; ecol, nrow = ecol+1, nrow+1 {
		sq := nrow*8 + ecol
		bitboard = SetBitboard(bitboard, sq)
		if IsBitboardSet(occupancies, sq) {
			break
		}
	}
	for wcol, srow := col-1, row-1; wcol >= 0 && wcol < 8 && srow >= 0 && srow < 8; wcol, srow = wcol-1, srow-1 {
		sq := srow*8 + wcol
		bitboard = SetBitboard(bitboard, sq)
		if IsBitboardSet(occupancies, sq) {
			break
		}
	}
	for ecol, srow := col+1, row-1; ecol < 8 && srow >= 0 && srow < 8; ecol, srow = ecol+1, srow-1 {
		sq := srow*8 + ecol
		bitboard = SetBitboard(bitboard, sq)
		if IsBitboardSet(occupancies, sq) {
			break
		}
	}

	return bitboard
}

func GenerateRookOccupancies(sq byte, includeEdges bool) []uint64 {
	col := sq % 8
	row := sq / 8
	occupancies := make([]uint64, 0)

	above := 8 - int(row) - 1
	below := int(row)
	west := int(col)
	east := 8 - int(col) - 1
	if !includeEdges {
		above = Max(above-1, 0)
		below = Max(below-1, 0)
		west = Max(west-1, 0)
		east = Max(east-1, 0)
	}

	for i := 0; i < 1<<uint(above); i++ {
		for j := 0; j < 1<<uint(below); j++ {
			for k := 0; k < 1<<uint(west); k++ {
				for l := 0; l < 1<<uint(east); l++ {
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

func GenerateBishopOccupancies(sq byte, includeEdges bool) []uint64 {
	col := sq % 8
	row := sq / 8
	occupancies := make([]uint64, 0)

	above := 8 - int(row) - 1
	below := int(row)
	west := int(col)
	east := 8 - int(col) - 1

	northWest := Min(above, west)
	northEast := Min(above, east)
	southWest := Min(below, west)
	southEast := Min(below, east)

	if !includeEdges {
		northWest = Max(northWest-1, 0)
		northEast = Max(northEast-1, 0)
		southWest = Max(southWest-1, 0)
		southEast = Max(southEast-1, 0)
	}

	for i := 0; i < 1<<uint(northWest); i++ {
		for j := 0; j < 1<<uint(northEast); j++ {
			for k := 0; k < 1<<uint(southWest); k++ {
				for l := 0; l < 1<<uint(southEast); l++ {
					var bitboard uint64
					bitboard = FollowRay(bitboard, col, row, NORTH_WEST, i)
					bitboard = FollowRay(bitboard, col, row, NORTH_EAST, j)
					bitboard = FollowRay(bitboard, col, row, SOUTH_WEST, k)
					bitboard = FollowRay(bitboard, col, row, SOUTH_EAST, l)
					occupancies = append(occupancies, bitboard)
				}
			}
		}
	}

	return occupancies
}

func GenerateRookMagic(sq byte, r *rand.Rand) (Magic, int) {
	mask := RookMask(sq)
	numBits := uint(bits.OnesCount64(mask))
	occupancies := GenerateRookOccupancies(sq, false)

	occupancyMoves := make(map[uint64]uint64)
	for _, occupancy := range occupancies {
		occupancyMoves[occupancy] = RookMoveBoard(sq, occupancy)
	}

	magicNumber, iterations := TrialAndErrorMagic(sq, r, numBits, occupancies, occupancyMoves)
	return Magic{Magic: magicNumber, Bits: numBits, Mask: mask, Sq: sq}, iterations
}

func GenerateBishopMagic(sq byte, r *rand.Rand) (Magic, int) {
	mask := BishopMask(sq)
	numBits := uint(bits.OnesCount64(mask))
	occupancies := GenerateBishopOccupancies(sq, false)

	occupancyMoves := make(map[uint64]uint64)
	for _, occupancy := range occupancies {
		occupancyMoves[occupancy] = BishopMoveBoard(sq, occupancy)
	}

	magicNumber, iterations := TrialAndErrorMagic(sq, r, numBits, occupancies, occupancyMoves)
	return Magic{Magic: magicNumber, Bits: numBits, Mask: mask, Sq: sq}, iterations
}

func TrialAndErrorMagic(
	sq byte,
	r *rand.Rand,
	numBits uint,
	occupancies []uint64,
	occupancyMoves map[uint64]uint64,
) (uint64, int) {
	var candidate uint64
	total := 0

TrialAndError:
	for {
		// try candidates until one of them is
		collisionMap := make(map[uint16]CollisionEntry)
		// overly bias zeros in the candidate - trick from https://github.com/goutham/magic-bits
		candidate = r.Uint64() & r.Uint64() & r.Uint64()
		total++

		collision := false
		for _, occupancy := range occupancies {
			hashKey := uint16((occupancy * candidate) >> uint(64-numBits))
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
	var rookMagics [64]Magic
	var bishopMagics [64]Magic

	for row := byte(0); row < 8; row++ {
		for col := byte(0); col < 8; col++ {
			sq := idx(col, row)
			rookMagic, rookIterations := GenerateRookMagic(idx(col, row), r)
			bishopMagic, bishopIterations := GenerateBishopMagic(idx(col, row), r)

			rookMagics[sq] = rookMagic
			bishopMagics[sq] = bishopMagic

			fmt.Printf("[%d] rook=%d (%d iterations), bishop = %d (%d iterations)\n",
				sq,
				rookMagic.Magic,
				rookIterations,
				bishopMagic.Magic,
				bishopIterations)
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

func GenerateSlidingMoves(
	rookMagics [64]Magic,
	bishopMagics [64]Magic,
) ([64][]SquareAttacks, [64][]SquareAttacks) {
	var rookAttacks [64][]SquareAttacks
	var bishopAttacks [64][]SquareAttacks

	for row := byte(0); row < 8; row++ {
		for col := byte(0); col < 8; col++ {
			sq := idx(col, row)

			rookAttacks[sq] = make([]SquareAttacks, 1<<(rookMagics[sq].Bits+1))
			bishopAttacks[sq] = make([]SquareAttacks, 1<<(bishopMagics[sq].Bits+1))

			GenerateRookSlidingMoves(sq, rookMagics[sq], rookAttacks[sq])
			GenerateBishopSlidingMoves(sq, bishopMagics[sq], bishopAttacks[sq])
		}
	}

	return rookAttacks, bishopAttacks
}

func GenerateRookSlidingMoves(
	sq byte,
	magic Magic,
	attacks []SquareAttacks,
) {
	occupancies := GenerateRookOccupancies(sq, true)
	for _, occupancy := range occupancies {
		key := hashKey(occupancy, magic)
		attackBoard := RookMoveBoard(sq, occupancy)
		moves := make([]Move, 64)
		end := CreateMovesFromBitboard(sq, attackBoard, moves, 0, 0)
		attacks[key] = SquareAttacks{
			board: attackBoard,
			moves: moves[0:end],
		}
	}
}

func GenerateBishopSlidingMoves(
	sq byte,
	magic Magic,
	attacks []SquareAttacks,
) {
	occupancies := GenerateBishopOccupancies(sq, true)
	for _, occupancy := range occupancies {
		key := hashKey(occupancy, magic)
		attackBoard := BishopMoveBoard(sq, occupancy)
		moves := make([]Move, 64)
		end := CreateMovesFromBitboard(sq, attackBoard, moves, 0, 0)
		attacks[key] = SquareAttacks{
			board: attackBoard,
			moves: moves[0:end],
		}
	}
}

func hashKey(occupancy uint64, magic Magic) uint16 {
	return uint16(((occupancy & magic.Mask) * magic.Magic) >> (64 - magic.Bits))
}

func outputMagicFile(magics [64]Magic, filename string) error {
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

func inputMagicFile(filename string) ([64]Magic, error) {
	var magics [64]Magic

	// Bad fix for finding the magic file in both xboard + test mode
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		ex, err := os.Executable()
		if err != nil {
			return magics, err
		}
		exPath := filepath.Dir(ex)

		filename = path.Join(exPath, filename)
	}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return magics, err
	}

	json.Unmarshal(b, &magics)
	return magics, nil
}

func idx(col byte, row byte) byte {
	return row*8 + col
}
