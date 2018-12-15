package main

import (
	"math/bits"
)

type MoveListing struct {
	moves      []Move
	captures   []Move
	promotions []Move
}

var moveBitboards *MoveBitboards

type MoveBitboards struct {
	pawnMoves          [2][64]uint64
	pawnAttacks        [2][64]uint64
	kingMoves          [64][]Move
	knightMoves        [64][]Move
	bishopMagics       map[byte]Magic
	rookMagics         map[byte]Magic
	rookSlidingMoves   map[byte]map[uint16][]Move
	bishopSlidingMoves map[byte]map[uint16][]Move
}

func CreateMoveBitboards() MoveBitboards {
	var pawnMoves [2][64]uint64
	var pawnAttacks [2][64]uint64
	var kingMoves [64][]Move
	var knightMoves [64][]Move

	for row := byte(0); row < 8; row++ {
		for col := byte(0); col < 8; col++ {
			sq := idx(col, row)

			pawnMoves[WHITE_OFFSET][sq], pawnMoves[BLACK_OFFSET][sq] = createPawnMoveBitboards(col, row)
			pawnAttacks[WHITE_OFFSET][sq], pawnAttacks[BLACK_OFFSET][sq] = createPawnAttackBitboards(col, row)
			kingMoves[sq] = CreateMovesFromBitboard(sq, getKingMoveBitboard(sq))
			knightMoves[sq] = CreateMovesFromBitboard(sq, getKnightMoveBitboard(sq))
		}
	}

	rookMagics, err := inputMagicFile("rook-magics.json")
	if err != nil {
		panic(err)
	}

	bishopMagics, err := inputMagicFile("bishop-magics.json")
	if err != nil {
		panic(err)
	}

	rookSlidingMoves, bishopSlidingMoves := GenerateSlidingMoves(rookMagics, bishopMagics)

	return MoveBitboards{
		pawnAttacks:        pawnAttacks,
		pawnMoves:          pawnMoves,
		kingMoves:          kingMoves,
		knightMoves:        knightMoves,
		bishopMagics:       bishopMagics,
		rookMagics:         rookMagics,
		rookSlidingMoves:   rookSlidingMoves,
		bishopSlidingMoves: bishopSlidingMoves,
	}
}

func getKnightMoveBitboard(sq byte) uint64 {
	col := sq % 8
	row := sq / 8

	var knightMoveBitboard uint64
	// 8 possibilities

	if row < 6 {
		// North
		if col > 0 {
			// East
			knightMoveBitboard = SetBitboard(knightMoveBitboard, idx(col-1, row+2))
		}
		if col < 7 {
			// West
			knightMoveBitboard = SetBitboard(knightMoveBitboard, idx(col+1, row+2))
		}
	}

	if row > 1 {
		// South
		if col > 0 {
			// West
			knightMoveBitboard = SetBitboard(knightMoveBitboard, idx(col-1, row-2))
		}
		if col < 7 {
			// East
			knightMoveBitboard = SetBitboard(knightMoveBitboard, idx(col+1, row-2))
		}
	}

	if col > 1 {
		// West
		if row > 0 {
			// South
			knightMoveBitboard = SetBitboard(knightMoveBitboard, idx(col-2, row-1))
		}
		if row < 7 {
			// North
			knightMoveBitboard = SetBitboard(knightMoveBitboard, idx(col-2, row+1))
		}
	}

	if col < 6 {
		// East
		if row > 0 {
			// South
			knightMoveBitboard = SetBitboard(knightMoveBitboard, idx(col+2, row-1))
		}
		if row < 7 {
			// North
			knightMoveBitboard = SetBitboard(knightMoveBitboard, idx(col+2, row+1))
		}
	}

	return knightMoveBitboard
}

func getKingMoveBitboard(sq byte) uint64 {
	col := sq % 8
	row := sq / 8
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

	var isWhite = boardState.whiteToMove
	var offset int
	if isWhite {
		offset = WHITE_OFFSET
	} else {
		offset = BLACK_OFFSET
	}

	occupancy := boardState.bitboards.color[offset]
	for occupancy != 0 {
		bbSq := byte(bits.TrailingZeros64(occupancy))
		sq := bitboardSquareToLegacySquare(bbSq)

		p := boardState.board[sq]

		if !isPawn(p) {
			generatePieceMoves(boardState, p, sq, isWhite, &listing)
		} else {
			generatePawnMoves(boardState, p, sq, isWhite, &listing)
		}

		occupancy ^= 1 << bbSq
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

// black is negative
var whitePawnCaptureOffsetArr = [2]int8{9, 11}
var blackPawnCaptureOffsetArr = [2]int8{-9, -11}

func generatePieceMoves(boardState *BoardState, p byte, sq byte, isWhite bool, listing *MoveListing) {
	// 8 possible directions to go (for the queen + king + knight)
	// 4 directions for bishop + rook
	// queen, bishop, and rook will continue along a "ray" until they find
	// a stopping point
	// Knight = 2, Bishop = 3, Rook = 4, Queen = 5, King = 6

	pieceType := p & 0x0F
	var offset int
	var otherOffset int
	if isWhite {
		offset = WHITE_OFFSET
		otherOffset = BLACK_OFFSET
	} else {
		offset = BLACK_OFFSET
		otherOffset = WHITE_OFFSET
	}

	moveBitboards := boardState.moveBitboards
	occupancy := boardState.bitboards.color[offset]
	bbSq := legacySquareToBitboardSquare(sq)

	var moves []Move

	switch pieceType {
	case KING_MASK:
		moves = moveBitboards.kingMoves[bbSq]
	case KNIGHT_MASK:
		moves = moveBitboards.knightMoves[bbSq]
	case BISHOP_MASK:
		otherOccupancy := boardState.bitboards.color[otherOffset]
		magic := moveBitboards.bishopMagics[bbSq]
		key := hashKey(occupancy|otherOccupancy, magic)
		moves = moveBitboards.bishopSlidingMoves[bbSq][key]
	case ROOK_MASK:
		otherOccupancy := boardState.bitboards.color[otherOffset]
		magic := moveBitboards.rookMagics[bbSq]
		key := hashKey(occupancy|otherOccupancy, magic)
		moves = moveBitboards.rookSlidingMoves[bbSq][key]
	case QUEEN_MASK:
		otherOccupancy := boardState.bitboards.color[otherOffset]
		bishopKey := hashKey(occupancy|otherOccupancy, moveBitboards.bishopMagics[bbSq])
		moves = moveBitboards.bishopSlidingMoves[bbSq][bishopKey]
		rookKey := hashKey(occupancy|otherOccupancy, moveBitboards.rookMagics[bbSq])
		moves = append(moves, moveBitboards.rookSlidingMoves[bbSq][rookKey]...)
	}

	for i := range moves {
		move := moves[i]
		move.to = bitboardSquareToLegacySquare(move.to)
		move.from = bitboardSquareToLegacySquare(move.from)

		oppositePiece := boardState.PieceAtSquare(move.to)
		if oppositePiece != EMPTY_SQUARE {
			if oppositePiece&0xF0 != p&0xF0 {
				move.flags |= CAPTURE_MASK
				listing.captures = append(listing.captures, move)
			} else {
				// same color, just skip it
				// I think this is how we need to filter out the precomputed sliding moves
			}
		} else {
			listing.moves = append(listing.moves, move)
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
	var otherOffset int
	var offset int
	if isWhite {
		offset = WHITE_OFFSET
		otherOffset = BLACK_OFFSET
	} else {
		offset = BLACK_OFFSET
		otherOffset = WHITE_OFFSET
	}

	bbSq := legacySquareToBitboardSquare(sq)
	otherOccupancies := boardState.bitboards.color[otherOffset]
	if boardState.boardInfo.enPassantTargetSquare != 0 {
		epSquare := legacySquareToBitboardSquare(boardState.boardInfo.enPassantTargetSquare)
		otherOccupancies = SetBitboard(otherOccupancies, epSquare)
	}
	pawnAttacks := boardState.moveBitboards.pawnAttacks[offset][bbSq]
	captures := CreateCapturesFromBitboard(bbSq, pawnAttacks&otherOccupancies)

	for _, capture := range captures {
		capture.to = bitboardSquareToLegacySquare(capture.to)
		capture.from = bitboardSquareToLegacySquare(capture.from)

		var destRank = Rank(capture.to)
		if (isWhite && destRank == RANK_8) || (!isWhite && destRank == RANK_1) {
			// promotion time
			var flags byte = PROMOTION_MASK | CAPTURE_MASK
			capture.flags = flags | QUEEN_MASK
			listing.promotions = append(listing.promotions, capture)
			capture.flags = flags | ROOK_MASK
			listing.promotions = append(listing.promotions, capture)
			capture.flags = flags | BISHOP_MASK
			listing.promotions = append(listing.promotions, capture)
			capture.flags = flags | KNIGHT_MASK
			listing.promotions = append(listing.promotions, capture)
		} else {
			if capture.to == boardState.boardInfo.enPassantTargetSquare {
				capture.flags |= SPECIAL1_MASK
			}
			listing.captures = append(listing.captures, capture)
		}
	}

	// TODO: must handle the double moves and being blocked by another piece
	var sqOffset int8
	if isWhite {
		sqOffset = 10
	} else {
		sqOffset = -10
	}

	var dest byte = uint8(int8(sq) + sqOffset)

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
				dest = uint8(int8(dest) + sqOffset)
				if boardState.board[dest] == EMPTY_SQUARE {
					listing.moves = append(listing.moves, CreateMove(sq, dest))
				}
			}
		}
	}
}
