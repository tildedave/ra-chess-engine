package main

import (
	"fmt"
)

var _ = fmt.Println

type MoveListing struct {
	moves      []Move
	captures   []Move
	promotions []Move
}

type MoveBitboards struct {
	pawnMoves   [2][64]uint64
	pawnAttacks [2][64]uint64
	kingMoves   [64]uint64
}

func CreateMoveBitboards() MoveBitboards {
	var pawnMoves [2][64]uint64
	var pawnAttacks [2][64]uint64
	var kingMoves [64]uint64

	for row := byte(0); row < 8; row++ {
		for col := byte(0); col < 8; col++ {
			sq := idx(col, row)
			pawnMoves[WHITE_OFFSET][sq], pawnMoves[BLACK_OFFSET][sq] = createPawnMoveBitboards(col, row)
			pawnAttacks[WHITE_OFFSET][sq], pawnAttacks[BLACK_OFFSET][sq] = createPawnAttackBitboards(col, row)
			kingMoves[sq] = createKingBitboard(col, row)
		}
	}

	return MoveBitboards{pawnAttacks: pawnAttacks, pawnMoves: pawnMoves, kingMoves: kingMoves}
}

func createKingBitboard(col byte, row byte) uint64 {
	var kingMoveBitboard uint64
	if row > 0 {
		kingMoveBitboard = SetBitboard(kingMoveBitboard, idx(col, row-1))
		if col > 0 {
			kingMoveBitboard = SetBitboard(kingMoveBitboard, idx(col-1, row-1))
		}
		if col < 7 {
			kingMoveBitboard = SetBitboard(kingMoveBitboard, idx(col+1, row-1))
		}
	}
	if row < 7 {
		kingMoveBitboard = SetBitboard(kingMoveBitboard, idx(col, row+1))
		if col > 0 {
			kingMoveBitboard = SetBitboard(kingMoveBitboard, idx(col-1, row+1))
		}
		if col < 7 {
			kingMoveBitboard = SetBitboard(kingMoveBitboard, idx(col+1, row+1))
		}
	}
	if col > 0 {
		kingMoveBitboard = SetBitboard(kingMoveBitboard, idx(col-1, row))
	}
	if col < 7 {
		kingMoveBitboard = SetBitboard(kingMoveBitboard, idx(col+1, row))
	}

	return kingMoveBitboard
}

func createPawnMoveBitboards(col byte, row byte) (uint64, uint64) {
	if row == 0 || row == 7 {
		return 0, 0
	}

	var whitePawnMoveBitboard uint64
	var blackPawnMoveBitboard uint64

	whitePawnMoveBitboard = SetBitboard(whitePawnMoveBitboard, idx(col, row+1))
	blackPawnMoveBitboard = SetBitboard(blackPawnMoveBitboard, idx(col, row+1))
	if row == 1 {
		whitePawnMoveBitboard = SetBitboard(whitePawnMoveBitboard, idx(col, row+2))
	} else if row == 7 {
		blackPawnMoveBitboard = SetBitboard(blackPawnMoveBitboard, idx(col, row-2))
	}

	return whitePawnMoveBitboard, blackPawnMoveBitboard
}

func createPawnAttackBitboards(col byte, row byte) (uint64, uint64) {
	if row == 0 || row == 7 {
		return 0, 0
	}

	var whitePawnAttackBitboard uint64
	var blackPawnAttackBitboard uint64

	if col > 0 {
		whitePawnAttackBitboard = SetBitboard(whitePawnAttackBitboard, idx(col-1, row+1))
		blackPawnAttackBitboard = SetBitboard(blackPawnAttackBitboard, idx(col-1, row-1))
	}

	if col < 7 {
		whitePawnAttackBitboard = SetBitboard(whitePawnAttackBitboard, idx(col+1, row+1))
		blackPawnAttackBitboard = SetBitboard(blackPawnAttackBitboard, idx(col+1, row-1))
	}

	return whitePawnAttackBitboard, blackPawnAttackBitboard
}

func createMoveListing() MoveListing {
	listing := MoveListing{}
	listing.moves = make([]Move, 0)
	listing.captures = make([]Move, 0)
	listing.promotions = make([]Move, 0)

	return listing
}

func GenerateMoveListing(boardState *BoardState) MoveListing {
	listing := createMoveListing()

	for i := byte(0); i < 8; i++ {
		for j := byte(0); j < 8; j++ {
			sq := RowAndColToSquare(i, j)
			p := boardState.board[sq]
			if p == EMPTY_SQUARE {
				continue
			}

			isWhite := p&BLACK_MASK != BLACK_MASK
			if isWhite == boardState.whiteToMove {
				if !isPawn(p) {
					generatePieceMoves(boardState, p, sq, isWhite, &listing)
				} else {
					generatePawnMoves(boardState, p, sq, isWhite, &listing)
				}
			}
		}
	}

	return listing
}

func GenerateMoves(boardState *BoardState) []Move {
	listing := GenerateMoveListing(boardState)

	l1 := len(listing.promotions)
	l2 := len(listing.captures)
	l3 := len(listing.moves)

	var moves = make([]Move, l1+l2+l3)
	var i int
	i += copy(moves[i:], listing.promotions)
	i += copy(moves[i:], listing.captures)
	i += copy(moves[i:], listing.moves)

	return moves
}

// These arrays are used for both move generation + king in check detection,
// so we'll pull them out of function scope

// ray starts from going up, then clockwise around
var offsetArr = [7][8]int8{
	[8]int8{0, 0, 0, 0, 0, 0, 0, 0},
	[8]int8{0, 0, 0, 0, 0, 0, 0, 0},
	[8]int8{-19, -8, 12, 21, 19, 8, -12, -21},
	[8]int8{-11, -9, 9, 11, 0, 0, 0, 0},
	[8]int8{-10, 1, 10, -1, 0, 0, 0, 0},
	[8]int8{-10, -9, 1, 11, 10, 9, -1, -11},
	[8]int8{-10, -9, 1, 11, 10, 9, -1, -11},
}

var slidingPieces = [7]bool{false, false, false, true, true, true, false}

// black is negative
var whitePawnCaptureOffsetArr = [2]int8{9, 11}
var blackPawnCaptureOffsetArr = [2]int8{-9, -11}

func generatePieceMoves(boardState *BoardState, p byte, sq byte, isWhite bool, listing *MoveListing) {
	// 8 possible directions to go (for the queen + king + knight)
	// 4 directions for bishop + rook
	// queen, bishop, and rook will continue along a "ray" until they find
	// a stopping point
	// Knight = 2, Bishop = 3, Rook = 4, Queen = 5, King = 6

	if p&0x0F == KING_MASK {
		var offset int
		if isWhite {
			offset = WHITE_OFFSET
		} else {
			offset = BLACK_OFFSET
		}
		bb := boardState.moveBitboards.kingMoves[legacySquareToBitboardSquare(sq)] & ^boardState.bitboards.color[offset]
		moves := CreateMovesFromBitboard(sq, bb)

		// eventually this will need to be done at the very end ...
		// not sure the .moves/.captures works well for the bitboard setup.  need to think more

		for i := range moves {
			move := moves[i]
			move.to = bitboardSquareToLegacySquare(move.to)

			oppositePiece := boardState.PieceAtSquare(move.to)
			if oppositePiece != EMPTY_SQUARE && oppositePiece&0xF0 != p&0xF0 {
				move.flags |= CAPTURE_MASK
				listing.captures = append(listing.captures, move)
			} else {
				listing.moves = append(listing.moves, move)
			}
		}
		// TODO must also handle captures
	} else {
		offsets := offsetArr[p&0x0F]

		var captureMask byte
		if isWhite {
			captureMask = BLACK_MASK
		} else {
			captureMask = WHITE_MASK
		}

		for _, value := range offsets {
			var dest byte = sq
			if value == 0 {
				continue
			}

			// we'll continue until we have to stop
			for true {
				dest = uint8(int8(dest) + value)
				destPiece := boardState.board[dest]

				if destPiece == SENTINEL_MASK {
					// stop - end of the board
					break
				} else if destPiece == EMPTY_SQUARE {
					// keep moving
					listing.moves = append(listing.moves, CreateMove(sq, dest))
				} else {
					if destPiece&captureMask == captureMask {
						listing.captures = append(listing.captures, CreateCapture(sq, dest))
					}

					// stop - hit another piece or made a capture
					break
				}

				// stop - piece type only gets one move
				if !slidingPieces[p&0x0F] {
					break
				}
			}
		}
	}

	// if piece is a king, castle logic
	// this doesn't have the provision against 'castle through check' or
	// 'castle outside of check' which we'll do later outside of move generation
	if p&0x0F == KING_MASK {
		if boardState.whiteToMove {
			if boardState.boardInfo.whiteCanCastleKingside &&
				boardState.board[SQUARE_F1] == EMPTY_SQUARE &&
				boardState.board[SQUARE_G1] == EMPTY_SQUARE &&
				boardState.board[SQUARE_H1] == WHITE_MASK|ROOK_MASK {
				listing.moves = append(listing.moves, CreateKingsideCastle(sq, SQUARE_G1))
			}
			if boardState.boardInfo.whiteCanCastleQueenside &&
				boardState.board[SQUARE_D1] == EMPTY_SQUARE &&
				boardState.board[SQUARE_C1] == EMPTY_SQUARE &&
				boardState.board[SQUARE_B1] == EMPTY_SQUARE &&
				boardState.board[SQUARE_A1] == WHITE_MASK|ROOK_MASK {
				listing.moves = append(listing.moves, CreateQueensideCastle(sq, SQUARE_C1))
			}
		} else {
			if boardState.boardInfo.blackCanCastleKingside &&
				boardState.board[SQUARE_F8] == EMPTY_SQUARE &&
				boardState.board[SQUARE_G8] == EMPTY_SQUARE &&
				boardState.board[SQUARE_H8] == BLACK_MASK|ROOK_MASK {
				listing.moves = append(listing.moves, CreateKingsideCastle(sq, SQUARE_G8))
			}
			if boardState.boardInfo.blackCanCastleQueenside &&
				boardState.board[SQUARE_D8] == EMPTY_SQUARE &&
				boardState.board[SQUARE_C8] == EMPTY_SQUARE &&
				boardState.board[SQUARE_B8] == EMPTY_SQUARE &&
				boardState.board[SQUARE_A8] == BLACK_MASK|ROOK_MASK {
				listing.moves = append(listing.moves, CreateQueensideCastle(sq, SQUARE_C8))
			}
		}
	}
}

func generatePawnMoves(boardState *BoardState, p byte, sq byte, isWhite bool, listing *MoveListing) {
	var offset int8
	if isWhite {
		offset = 10
	} else {
		offset = -10
	}

	var dest byte = uint8(int8(sq) + offset)

	if boardState.board[dest] == EMPTY_SQUARE {
		var sourceRank byte = Rank(sq)
		// promotion
		if (isWhite && sourceRank == RANK_7) || (!isWhite && sourceRank == RANK_2) {
			// promotions are color-maskless
			listing.promotions = append(listing.promotions, CreatePromotion(sq, dest, QUEEN_MASK))
			listing.promotions = append(listing.promotions, CreatePromotion(sq, dest, BISHOP_MASK))
			listing.promotions = append(listing.promotions, CreatePromotion(sq, dest, KNIGHT_MASK))
			listing.promotions = append(listing.promotions, CreatePromotion(sq, dest, ROOK_MASK))
		} else {
			// empty square
			listing.moves = append(listing.moves, CreateMove(sq, dest))

			if (isWhite && sourceRank == RANK_2) ||
				(!isWhite && sourceRank == RANK_7) {
				// home row for white so we can move one more
				dest = uint8(int8(dest) + offset)
				if boardState.board[dest] == EMPTY_SQUARE {
					listing.moves = append(listing.moves, CreateMove(sq, dest))
				}
			}
		}
	}

	var pawnCaptureOffsetArr [2]int8
	if isWhite {
		pawnCaptureOffsetArr = whitePawnCaptureOffsetArr
	} else {
		pawnCaptureOffsetArr = blackPawnCaptureOffsetArr
	}

	for _, offset := range pawnCaptureOffsetArr {
		var dest byte = uint8(int8(sq) + offset)

		if boardState.boardInfo.enPassantTargetSquare == dest {
			listing.captures = append(listing.captures, CreateEnPassantCapture(sq, dest))
			continue
		}

		destPiece := boardState.board[dest]
		if destPiece == SENTINEL_MASK || destPiece == EMPTY_SQUARE {
			continue
		}

		isDestPieceWhite := destPiece&BLACK_MASK != BLACK_MASK
		if isWhite != isDestPieceWhite {
			var destRank byte = Rank(dest)
			if (isWhite && destRank == RANK_8) || (!isWhite && destRank == RANK_1) {
				// promotion time
				listing.promotions = append(listing.promotions, CreatePromotionCapture(sq, dest, QUEEN_MASK))
				listing.promotions = append(listing.promotions, CreatePromotionCapture(sq, dest, ROOK_MASK))
				listing.promotions = append(listing.promotions, CreatePromotionCapture(sq, dest, BISHOP_MASK))
				listing.promotions = append(listing.promotions, CreatePromotionCapture(sq, dest, KNIGHT_MASK))
			} else {
				listing.captures = append(listing.captures, CreateCapture(sq, dest))
			}
		}
	}
}
