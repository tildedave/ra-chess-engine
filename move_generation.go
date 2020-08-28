package main

import (
	"container/heap"
	"math/bits"
)

type MoveListing struct {
	moves      []Move
	captures   []Move
	promotions []Move
}

type PrecomputedInfo struct {
	side           int
	otherSide      int
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

// Lifted from Apep https://github.com/tildedave/apep-chess-engine/blob/master/src/movegen.cpp#L8
var mvvPriority = [7][7]int{
	{0, 0, 0, 0, 0, 0, 0},
	{0, 6, 12, 18, 24, 30, 0}, // PAWN CAPTURES PRIORITY: PAWN, KNIGHT, BISHOP, ROOK, QUEEN
	{0, 5, 11, 17, 23, 29, 0}, // KNIGHT CAPTURES PRIORITY: PAWN, KNIGHT, BISHOP, ROOK, QUEEN
	{0, 4, 10, 16, 22, 28, 0}, // BISHOP CAPTURES PRIORITY: PAWN, KNIGHT, BISHOP, ROOK, QUEEN
	{0, 3, 9, 15, 21, 27, 0},  // ROOK CAPTURES PRIORITY: PAWN, KNIGHT, BISHOP, ROOK, QUEEN
	{0, 2, 8, 14, 20, 26, 0},  // QUEEN CAPTURES PRIORITY: PAWN, KNIGHT, BISHOP, ROOK, QUEEN
	{0, 1, 7, 13, 19, 25, 0},  // KING CAPTURES PRIORITY: PAWN, KNIGHT, BISHOP, ROOK, QUEEN
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

// GenerateMoveListing generates a MoveListing structure and a MoveSizeHint structure.  If the
// applyOrdering flag is passed, the ordering will be sorted in an attempt to choose good moves
// first.
// MoveSizeHint is passed in and returned to avoid the cost of slice recomputation.
func GenerateMoveListing(boardState *BoardState, hint MoveSizeHint, applyOrdering bool) (MoveListing, MoveSizeHint) {
	listing := createMoveListing(hint)
	precomputedInfo := generatePrecomputedInfo(boardState)
	occupancy := precomputedInfo.ourOccupancy
	for occupancy != 0 {
		sq := byte(bits.TrailingZeros64(occupancy))
		GenerateMovesFromSquare(boardState, sq, boardState.sideToMove, &listing, precomputedInfo)

		occupancy ^= 1 << sq
	}

	hint = MoveSizeHint{
		numMoves:      len(listing.moves),
		numCaptures:   len(listing.captures),
		numPromotions: len(listing.promotions),
	}

	if applyOrdering {
		orderCaptures(boardState, &listing)
	}

	return listing, hint
}

func GenerateQuiescentMoveListing(boardState *BoardState, hint MoveSizeHint) (MoveListing, MoveSizeHint) {
	// for now only generate captures
	listing := createMoveListing(hint)
	precomputedInfo := generatePrecomputedInfo(boardState)
	occupancy := precomputedInfo.ourOccupancy
	moveBitboards := boardState.moveBitboards

	for occupancy != 0 {
		sq := byte(bits.TrailingZeros64(occupancy))
		fromPiece := boardState.board[sq] & 0x0F
		var attackBoard uint64
		switch fromPiece {
		case PAWN_MASK:
			attackBoard = moveBitboards.pawnAttacks[boardState.sideToMove][sq]
		case KNIGHT_MASK:
			attackBoard = moveBitboards.knightAttacks[sq].board
		case KING_MASK:
			attackBoard = moveBitboards.kingAttacks[sq].board
		case BISHOP_MASK:
			bishopKey := hashKey(precomputedInfo.allOccupancy, moveBitboards.bishopMagics[sq])
			attackBoard = moveBitboards.bishopAttacks[sq][bishopKey].board
		case ROOK_MASK:
			rookKey := hashKey(precomputedInfo.allOccupancy, moveBitboards.rookMagics[sq])
			attackBoard = moveBitboards.rookAttacks[sq][rookKey].board
		case QUEEN_MASK:
			rookKey := hashKey(precomputedInfo.allOccupancy, moveBitboards.rookMagics[sq])
			bishopKey := hashKey(precomputedInfo.allOccupancy, moveBitboards.bishopMagics[sq])
			attackBoard = (moveBitboards.rookAttacks[sq][rookKey].board | moveBitboards.bishopAttacks[sq][bishopKey].board)
		}

		generateCapturesFromBoard(&listing, sq, attackBoard&precomputedInfo.otherOccupancy)

		occupancy ^= 1 << sq
	}

	orderCaptures(boardState, &listing)

	return listing, MoveSizeHint{numCaptures: len(listing.captures)}
}

func orderCaptures(boardState *BoardState, listing *MoveListing) {
	captureQueue := make(MovePriorityQueue, 0, len(listing.captures))

	for _, capture := range listing.captures {
		fromPiece := boardState.PieceAtSquare(capture.from)
		toPiece := boardState.PieceAtSquare(capture.to)
		priority := mvvPriority[fromPiece&0x0F][toPiece&0x0F]
		item := Item{move: capture, score: priority}
		captureQueue.Push(&item)
	}
	heap.Init(&captureQueue)

	listing.captures = make([]Move, 0, captureQueue.Len())
	for captureQueue.Len() > 0 {
		item := heap.Pop(&captureQueue).(*Item)
		listing.captures = append(listing.captures, item.move)
	}
}
func generateCapturesFromBoard(moveListing *MoveListing, from byte, attackBoard uint64) {
	for attackBoard != 0 {
		to := byte(bits.TrailingZeros64(attackBoard))
		moveListing.captures = append(moveListing.captures, CreateMove(from, to))
		attackBoard ^= 1 << to
	}
}

func generatePrecomputedInfo(boardState *BoardState) *PrecomputedInfo {
	precomputedInfo := PrecomputedInfo{side: boardState.sideToMove}
	switch boardState.sideToMove {
	case WHITE_OFFSET:
		precomputedInfo.otherSide = BLACK_OFFSET
		precomputedInfo.ourOccupancy = boardState.bitboards.color[WHITE_OFFSET]
		precomputedInfo.otherOccupancy = boardState.bitboards.color[BLACK_OFFSET]
	case BLACK_OFFSET:
		precomputedInfo.otherSide = WHITE_OFFSET
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

func GenerateMoves(boardState *BoardState, moves []Move, start int) int {
	listing, _ := GenerateMoveListing(boardState, MoveSizeHint{}, true)

	start += copy(moves[start:], listing.promotions)
	start += copy(moves[start:], listing.captures)
	start += copy(moves[start:], listing.moves)

	return start
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
		if boardState.sideToMove == WHITE_OFFSET {
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
	captures := CreateMovesFromBitboard(sq, pawnAttacks&otherOccupancies)

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
				capture.flags |= SPECIAL1_MASK | CAPTURE_MASK
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
