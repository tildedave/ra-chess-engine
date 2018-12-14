package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
		perftResult := Perft(&board, spec.Depth, options)
		if perftResult.nodes != spec.Nodes {
			fmt.Printf("NOT OK: %s (depth=%d, expected nodes=%d, actual nodes=%d)\n", spec.Fen, spec.Depth, spec.Nodes, perftResult.nodes)
			allSuccess = false
		} else {
			fmt.Printf("OK: %s (depth=%d, nodes=%d)\n", spec.Fen, spec.Depth, spec.Nodes)
		}
	}

	if allSuccess {
		return true, nil
	}

	return false, nil
}

func RunPerft(startingFen string, depth uint, options PerftOptions) (bool, error) {
	for i := uint(0); i <= depth; i++ {
		board, err := CreateBoardStateFromFENString(startingFen)
		if err == nil {
			fmt.Println(Perft(&board, i, options))
		} else {
			fmt.Println(err)
		}
	}

	return true, nil
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

	listing := GenerateMoveListing(boardState)
	captures := uint(0)
	castles := uint(0)
	promotions := uint(0)

	for _, moveList := range [][]Move{listing.captures, listing.moves, listing.promotions} {
		for _, move := range moveList {
			var originalHashKey uint32
			if options.sanityCheck {
				testMoveLegality(boardState, move)
				originalHashKey = boardState.hashKey

				if boardState.board[boardState.lookupInfo.blackKingSquare] != BLACK_MASK|KING_MASK {
					fmt.Println(boardState.ToString())
					panic("black king not at expected position")
				}
				for row := byte(0); row < 8; row++ {
					for col := byte(0); col < 8; col++ {
						legacySq := RowAndColToSquare(row, col)
						sq := idx(col, row)
						piece := boardState.PieceAtSquare(legacySq)
						var isError = false
						var message string
						var bitboard uint64

						if piece == EMPTY_SQUARE {
							for _, colorOffset := range []int{WHITE_OFFSET, BLACK_OFFSET} {
								if IsBitboardSet(boardState.bitboards.color[colorOffset], sq) {
									isError = true
									message = "Empty square had occupancy set"
									bitboard = boardState.bitboards.color[colorOffset]
								}
							}
							for _, pieceOffset := range BITBOARD_PIECES {
								if IsBitboardSet(boardState.bitboards.piece[pieceOffset], sq) {
									isError = true
									message = "Empty square was set for piece offset"
									bitboard = boardState.bitboards.color[pieceOffset]
								}
							}
						} else {
							colorOffset := PieceToColorOffset(piece)
							var otherColorOffset = 1
							if colorOffset == 1 {
								otherColorOffset = 0
							}
							if !IsBitboardSet(boardState.bitboards.color[colorOffset], sq) {
								isError = true
								message = "Color occupancy bitboard was not set for piece on square"
								bitboard = boardState.bitboards.color[colorOffset]
							}
							if IsBitboardSet(boardState.bitboards.color[otherColorOffset], sq) {
								isError = true
								message = "Color occupancy bitboard was set for opposite color of piece on square"
								bitboard = boardState.bitboards.color[otherColorOffset]
							}
							for _, pieceOffset := range BITBOARD_PIECES {
								if pieceOffset == piece&0x0F {
									if !IsBitboardSet(boardState.bitboards.piece[pieceOffset], sq) {
										isError = true
										message = fmt.Sprintf("Piece occupancy bitboard was not set for piece on square (piece=%d)", pieceOffset)
										bitboard = boardState.bitboards.piece[pieceOffset]
									}
								} else if IsBitboardSet(boardState.bitboards.piece[pieceOffset], sq) {
									isError = true
									message = "Piece occupancy bitboard was set for wrong kind of piece"
									bitboard = boardState.bitboards.piece[pieceOffset]
								}
							}
						}
						if isError {
							fmt.Println(boardState.ToString())
							fmt.Println(BitboardToString(bitboard))
							fmt.Println(MoveToString(move))
							panic(message)
						}
					}
				}
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

			wasValid := false
			if !boardState.IsInCheck(!boardState.whiteToMove) {
				wasValid = true

				if move.IsCapture() {
					captures++
				}
				if move.IsKingsideCastle() || move.IsQueensideCastle() {
					castles++
				}
				if move.IsPromotion() {
					promotions++
				}

				info := Perft(boardState, depth-1, options)
				addPerftInfo(&perftInfo, info)
			}

			boardState.UnapplyMove(move)
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
		fmt.Println(boardState.ToString())
		fmt.Println(MoveToString(move))
		panic("Illegal move")
	}
}
