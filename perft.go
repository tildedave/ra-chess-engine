package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

type PerftSpecification struct {
	Depth uint   `json:depth`
	Nodes uint   `json:nodes`
	Fen   string `json:fen`
}

var _ = fmt.Println

type PerftInfo struct {
	nodes      uint
	captures   uint
	castles    uint
	promotions uint
	checks     uint
}

type PerftOptions struct {
	checks          bool
	sanityCheck     bool
	perftPrintMoves bool
	depth           uint
	divide          bool
}

func RunPerftJson(perftJsonFile string, options PerftOptions) (bool, error) {
	b, err := ioutil.ReadFile(perftJsonFile)
	if err != nil {
		panic(err)
	}

	var specs []PerftSpecification
	json.Unmarshal(b, &specs)

	allSuccess := true
	for _, spec := range specs {
		board, err := CreateBoardStateFromFENString(spec.Fen)
		if err != nil {
			fmt.Println("Unable to parse FEN " + spec.Fen + ", continuing")
			fmt.Println(err)
			continue
		}
		options.depth = spec.Depth

		start := time.Now()
		moves := make([]Move, 13824)
		var moveStart [64]int
		perftResult := Perft(&board, spec.Depth, options, moves[:], moveStart[:])
		elapsed := time.Since(start)
		if perftResult.nodes != spec.Nodes {
			fmt.Printf("NOT OK: %s (depth=%d, expected nodes=%d, actual nodes=%d; duration=%s)\n", spec.Fen, spec.Depth, spec.Nodes, perftResult.nodes, elapsed)
			allSuccess = false
		} else {
			fmt.Printf("OK: %s (depth=%d, nodes=%d; duration=%s)\n", spec.Fen, spec.Depth, spec.Nodes, elapsed)
		}
	}

	if allSuccess {
		return true, nil
	}

	return false, nil
}

func RunPerft(fen string, variation string, depth uint, options PerftOptions) (bool, error) {
	for i := uint(0); i <= depth; i++ {
		boardState, err := CreateBoardStateFromFENStringWithVariation(fen, variation)
		if err != nil {
			return false, err
		}

		if i == uint(0) {
			fmt.Println(boardState.String())
		}

		if err == nil {
			options.depth = i
			start := time.Now()
			moves := make([]Move, 13824)
			var moveStart [64]int
			result := Perft(&boardState, i, options, moves[:], moveStart[:])
			fmt.Printf("%d\t%10d\t%s\n", i, result.nodes, time.Since(start))
		} else {
			fmt.Println(err)
		}
	}

	return true, nil
}

func Perft(boardState *BoardState, depth uint, options PerftOptions, moves []Move, moveStart []int) PerftInfo {
	var perftInfo PerftInfo

	if options.checks && boardState.IsInCheck(boardState.sideToMove) {
		perftInfo.checks += 1
	}

	if depth == 0 {
		perftInfo.nodes = 1
		return perftInfo
	}

	currentDepth := options.depth - depth
	start := moveStart[currentDepth]
	end := GenerateMoves(boardState, moves, start)
	moveStart[currentDepth+1] = end
	captures := uint(0)
	castles := uint(0)
	promotions := uint(0)

	for i := start; i < end; i++ {
		move := moves[i]
		var originalHashKey uint64
		var originalPawnHashKey uint64
		if options.sanityCheck {
			testMoveLegality(boardState, move)
			originalHashKey = boardState.hashKey
			originalPawnHashKey = boardState.pawnHashKey
			sanityCheckBitboards(MoveToString(move, boardState), boardState)
		}

		if move.IsCastle() && !boardState.TestCastleLegality(move) {
			continue
		}

		boardState.ApplyMove(move)

		wasValid := false
		var otherOffset int
		switch boardState.sideToMove {
		case WHITE_OFFSET:
			otherOffset = BLACK_OFFSET
		case BLACK_OFFSET:
			otherOffset = WHITE_OFFSET
		}
		if boardState.IsInCheck(otherOffset) {
			boardState.UnapplyMove(move)
			continue
		}

		wasValid = true
		if boardState.wasCapture[boardState.moveIndex-1] {
			captures++
		}
		if move.IsCastle() {
			castles++
		}
		if move.IsPromotion() {
			promotions++
		}

		info := Perft(boardState, depth-1, options, moves, moveStart)
		boardState.UnapplyMove(move)

		if options.divide && depth == options.depth {
			fmt.Printf("%s %d\n", MoveToString(move, boardState), info.nodes)
		}
		addPerftInfo(&perftInfo, info)

		if depth == 1 && options.perftPrintMoves {
			if wasValid {
				fmt.Println(MoveToPrettyString(move, boardState))
			} else {
				fmt.Println("ILLEGAL: " + MoveToPrettyString(move, boardState))
			}
		}
		if options.sanityCheck {
			if boardState.hashKey != originalHashKey {
				fmt.Printf("Unapplying move did not restore original hash key: %s (%d vs %d)\n",
					MoveToPrettyString(move, boardState),
					boardState.hashKey,
					originalHashKey)
			}
			if boardState.pawnHashKey != originalPawnHashKey {
				fmt.Printf("Unapplying move did not restore original pawn hash key: %s (%d vs %d)\n",
					MoveToPrettyString(move, boardState),
					boardState.pawnHashKey,
					originalPawnHashKey)
			}
		}
	}

	perftInfo.captures += captures
	perftInfo.castles += castles
	perftInfo.promotions += promotions

	return perftInfo
}

func addPerftInfo(info1 *PerftInfo, info2 PerftInfo) {
	info1.nodes += info2.nodes
	info1.captures += info2.captures
	info1.castles += info2.castles
	info1.promotions += info2.promotions
	info1.checks += info2.checks
}

func testMoveLegality(boardState *BoardState, move Move) {
	legal, err := boardState.IsMoveLegal(move)
	if !legal {
		fmt.Println(err)
		fmt.Println(boardState.String())
		fmt.Println(MoveToString(move, boardState))
		panic("Illegal move")
	}
}
