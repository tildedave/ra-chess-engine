package main

import (
	"math/bits"
)

type MoveListing struct {
	moves      []Move
	captures   []Move
	promotions []Move
}

type PrecomputedInfo struct {
	offset         int
	otherOffset    int
	ourOccupancy   uint64
	otherOccupancy uint64
	allOccupancy   uint64
}

type SquareAttacks struct {
	moves []Move
	board uint64
}

type MoveSizeHint struct {
	numMoves      int
	numCaptures   int
	numPromotions int
}

var moveBitboards *MoveBitboards

type MoveBitboards struct {
	pawnMoves   [2][64]uint64
	pawnAttacks [2][64]uint64

	knightAttacks [64]SquareAttacks
	kingAttacks   [64]SquareAttacks

	bishopMagics  [64]Magic
	rookMagics    [64]Magic
	rookAttacks   [64][]SquareAttacks
	bishopAttacks [64][]SquareAttacks
}

func CreateMoveBitboards() MoveBitboards {
	var pawnMoves [2][64]uint64
	var pawnAttacks [2][64]uint64
	var kingAttacks [64]SquareAttacks
	var knightAttacks [64]SquareAttacks

	for row := byte(0); row < 8; row++ {
		for col := byte(0); col < 8; col++ {
			sq := idx(col, row)

			pawnMoves[WHITE_OFFSET][sq], pawnMoves[BLACK_OFFSET][sq] = createPawnMoveBitboards(col, row)
			pawnAttacks[WHITE_OFFSET][sq], pawnAttacks[BLACK_OFFSET][sq] = createPawnAttackBitboards(col, row)

			kingSqAttacks := SquareAttacks{}
			kingSqAttacks.board = getKingMoveBitboard(sq)
			kingSqAttacks.moves = CreateMovesFromBitboard(sq, kingSqAttacks.board)

			knightSqAttacks := SquareAttacks{}
			knightSqAttacks.board = getKnightMoveBitboard(sq)
			knightSqAttacks.moves = CreateMovesFromBitboard(sq, knightSqAttacks.board)

			kingAttacks[sq] = kingSqAttacks
			knightAttacks[sq] = knightSqAttacks
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

	rookAttacks, bishopAttacks := GenerateSlidingMoves(rookMagics, bishopMagics)

	return MoveBitboards{
		pawnAttacks:   pawnAttacks,
		pawnMoves:     pawnMoves,
		kingAttacks:   kingAttacks,
		knightAttacks: knightAttacks,
		bishopMagics:  bishopMagics,
		rookMagics:    rookMagics,
		rookAttacks:   rookAttacks,
		bishopAttacks: bishopAttacks,
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
	// We must create pawn attack bitboards for all squares on the board because they
	// get used in check detection by having the king pretend to be a pawn of that color.

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

func createMoveListing(hint MoveSizeHint) MoveListing {
	listing := MoveListing{}
	listing.moves = make([]Move, 0, hint.numMoves)
	listing.captures = make([]Move, 0, hint.numCaptures)
	listing.promotions = make([]Move, 0, hint.numPromotions)

	return listing
}

func GenerateMoveListing(boardState *BoardState, hint MoveSizeHint) (MoveListing, MoveSizeHint) {
	listing := createMoveListing(hint)
	precomputedInfo := generatePrecomputedInfo(boardState)
	occupancy := precomputedInfo.ourOccupancy
	for occupancy != 0 {
		sq := byte(bits.TrailingZeros64(occupancy))
		GenerateMovesFromSquare(boardState, sq, boardState.offsetToMove, &listing, precomputedInfo)

		occupancy ^= 1 << sq
	}

	hint = MoveSizeHint{
		numMoves:      len(listing.moves),
		numCaptures:   len(listing.captures),
		numPromotions: len(listing.promotions),
	}
	return listing, hint
}

func generatePrecomputedInfo(boardState *BoardState) *PrecomputedInfo {
	precomputedInfo := PrecomputedInfo{}
	switch boardState.offsetToMove {
	case WHITE_OFFSET:
		precomputedInfo.otherOffset = BLACK_OFFSET
		precomputedInfo.ourOccupancy = boardState.bitboards.color[WHITE_OFFSET]
		precomputedInfo.otherOccupancy = boardState.bitboards.color[BLACK_OFFSET]
	case BLACK_OFFSET:
		precomputedInfo.otherOffset = WHITE_OFFSET
		precomputedInfo.ourOccupancy = boardState.bitboards.color[BLACK_OFFSET]
		precomputedInfo.otherOccupancy = boardState.bitboards.color[WHITE_OFFSET]
	}
	precomputedInfo.allOccupancy = precomputedInfo.ourOccupancy | precomputedInfo.otherOccupancy

	return &precomputedInfo
}

func GenerateMovesFromSquare(
	boardState *BoardState,
	sq byte,
	offset int,
	listing *MoveListing,
	precomputedInfo *PrecomputedInfo,
) {
	p := boardState.board[sq]
	if !isPawn(p) {
		generatePieceMoves(boardState, p, sq, offset, listing, precomputedInfo)
	} else {
		generatePawnMoves(boardState, p, sq, offset, listing, precomputedInfo)
	}
}

func GenerateMoves(boardState *BoardState) []Move {
	listing, _ := GenerateMoveListing(boardState, MoveSizeHint{})

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

func generatePieceMoves(
	boardState *BoardState,
	p byte,
	sq byte,
	offset int,
	listing *MoveListing,
	precomputedInfo *PrecomputedInfo,
) {
	// 8 possible directions to go (for the queen + king + knight)
	// 4 directions for bishop + rook
	// queen, bishop, and rook will continue along a "ray" until they find
	// a stopping point
	// Knight = 2, Bishop = 3, Rook = 4, Queen = 5, King = 6

	pieceType := p & 0x0F
	moveBitboards := boardState.moveBitboards

	var moves []Move

	switch pieceType {
	case KING_MASK:
		moves = moveBitboards.kingAttacks[sq].moves
	case KNIGHT_MASK:
		moves = moveBitboards.knightAttacks[sq].moves
	case BISHOP_MASK:
		magic := moveBitboards.bishopMagics[sq]
		key := hashKey(precomputedInfo.allOccupancy, magic)
		moves = moveBitboards.bishopAttacks[sq][key].moves
	case ROOK_MASK:
		magic := moveBitboards.rookMagics[sq]
		key := hashKey(precomputedInfo.allOccupancy, magic)
		moves = moveBitboards.rookAttacks[sq][key].moves
	case QUEEN_MASK:
		bishopKey := hashKey(precomputedInfo.allOccupancy, moveBitboards.bishopMagics[sq])
		moves = moveBitboards.bishopAttacks[sq][bishopKey].moves
		rookKey := hashKey(precomputedInfo.allOccupancy, moveBitboards.rookMagics[sq])
		moves = append(moves, moveBitboards.rookAttacks[sq][rookKey].moves...)
	}

	for i := range moves {
		move := moves[i]
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
		if boardState.offsetToMove == WHITE_OFFSET {
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

func generatePawnMoves(
	boardState *BoardState,
	p byte,
	sq byte,
	offset int,
	listing *MoveListing,
	precomputedInfo *PrecomputedInfo,
) {
	var isWhite bool
	if offset == WHITE_OFFSET {
		isWhite = true
	} else {
		isWhite = false
	}

	otherOccupancies := precomputedInfo.otherOccupancy
	if boardState.boardInfo.enPassantTargetSquare != 0 {
		otherOccupancies = SetBitboard(otherOccupancies, boardState.boardInfo.enPassantTargetSquare)
	}
	pawnAttacks := boardState.moveBitboards.pawnAttacks[offset][sq]
	captures := CreateCapturesFromBitboard(sq, pawnAttacks&otherOccupancies)

	for _, capture := range captures {
		var destRank = Rank(capture.to)
		if (offset == WHITE_OFFSET && destRank == RANK_8) || (offset == BLACK_OFFSET && destRank == RANK_1) {
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
	if offset == WHITE_OFFSET {
		sqOffset = 8
	} else {
		sqOffset = -8
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
