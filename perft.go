package main

import (
	"fmt"
)

var _ = fmt.Println

type PerftInfo struct {
	nodes      uint
	captures   uint
	castles    uint
	promotions uint
	checks     uint
}

type PerftOptions struct {
	checks      bool
	sanityCheck bool
}

func Perft(boardState *BoardState, depth uint, options PerftOptions) PerftInfo {
	var perftInfo PerftInfo

	if options.checks && boardState.IsInCheck(boardState.whiteToMove) {
		perftInfo.checks += 1
	}

	if depth == 0 {
		perftInfo.nodes = 1
		return perftInfo
	}

	moves := GenerateMoves(boardState)
	captures := uint(0)
	castles := uint(0)
	promotions := uint(0)

	for _, move := range moves {
		if options.sanityCheck {
			testMoveLegality(boardState, move)
		}

		if move.IsCastle() && !boardState.TestCastleLegality(move) {
			continue
		}

		boardState.ApplyMove(move)

		if options.sanityCheck {
			if boardState.board[boardState.lookupInfo.whiteKingSquare] != WHITE_MASK|KING_MASK {
				fmt.Println(boardState.ToString())
				panic("white king not at expected position")
			}
			if boardState.board[boardState.lookupInfo.blackKingSquare] != BLACK_MASK|KING_MASK {
				fmt.Println(boardState.ToString())
				panic("black king not at expected position")
			}
		}

		if !boardState.IsInCheck(!boardState.whiteToMove) {
			if move.IsCapture() {
				captures++
			} else if move.IsKingsideCastle() || move.IsQueensideCastle() {
				castles++
			} else if move.IsPromotion() {
				promotions++
			}

			info := Perft(boardState, depth-1, options)
			addPerftInfo(&perftInfo, info)
		}

		boardState.UnapplyMove(move)
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
		fmt.Println(boardState.ToString())
		fmt.Println(MoveToString(move))
		panic("Illegal move")
	}
}
