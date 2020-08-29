package main

import (
	"container/heap"
	"math/bits"
)

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
			moves := make([]Move, 12)
			end := CreateMovesFromBitboard(sq, kingSqAttacks.board, moves, 0, 0)
			kingSqAttacks.moves = moves[0:end]

			knightSqAttacks := SquareAttacks{}
			knightSqAttacks.board = getKnightMoveBitboard(sq)
			moves = make([]Move, 12)
			end = CreateMovesFromBitboard(sq, knightSqAttacks.board, moves, 0, 0)
			knightSqAttacks.moves = moves[0:end]

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

// GenerateMoveListing writes into the moves array beginning at the start index, returning the new end
// index.
func GenerateMoveListing(boardState *BoardState, moves []Move, start int, applyOrdering bool) int {
	precomputedInfo := generatePrecomputedInfo(boardState)
	occupancy := precomputedInfo.ourOccupancy
	originalStart := start
	for occupancy != 0 {
		sq := byte(bits.TrailingZeros64(occupancy))
		start = GenerateMovesFromSquare(boardState, sq, boardState.sideToMove, moves, start, precomputedInfo)

		occupancy ^= 1 << sq
	}

	if applyOrdering {
		orderCaptures(boardState, moves, originalStart, start)
	}

	return start
}

func GenerateQuiescentMoves(boardState *BoardState, moves []Move, start int) int {
	// for now only generate captures
	originalStart := start
	precomputedInfo := generatePrecomputedInfo(boardState)
	occupancy := precomputedInfo.ourOccupancy
	moveBitboards := boardState.moveBitboards

	for occupancy != 0 {
		sq := byte(bits.TrailingZeros64(occupancy))
		fromPiece := boardState.board[sq] & 0x0F
		var attackBoard uint64
		var flags byte = CAPTURE_MASK
		switch fromPiece {
		case PAWN_MASK:
			attackBoard = moveBitboards.pawnAttacks[boardState.sideToMove][sq]
			rank := Rank(sq)
			if (boardState.sideToMove == WHITE_OFFSET && rank == RANK_7) ||
				(boardState.sideToMove == BLACK_OFFSET && rank == RANK_2) {
				flags |= PROMOTION_MASK | QUEEN_MASK
			}
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

		start = CreateMovesFromBitboard(sq, attackBoard&precomputedInfo.otherOccupancy, moves, start, flags)

		occupancy ^= 1 << sq
	}

	orderCaptures(boardState, moves, originalStart, start)

	return start
}

func orderCaptures(boardState *BoardState, moves []Move, start int, end int) {
	captureQueue := make(MovePriorityQueue, 0, end-start)

	for _, capture := range moves[start:end] {
		fromPiece := boardState.PieceAtSquare(capture.from)
		toPiece := boardState.PieceAtSquare(capture.to)
		priority := mvvPriority[fromPiece&0x0F][toPiece&0x0F]
		item := Item{move: capture, score: priority}
		captureQueue.Push(&item)
	}
	heap.Init(&captureQueue)

	current := start
	for captureQueue.Len() > 0 {
		item := heap.Pop(&captureQueue).(*Item)
		moves[current] = item.move
		current++
	}
}

func CreateMovesFromBitboard(from byte, bitboard uint64, moves []Move, start int, flags byte) int {
	for bitboard != 0 {
		to := byte(bits.TrailingZeros64(bitboard))
		moves[start] = Move{from: from, to: to, flags: flags}
		start++
		bitboard ^= 1 << to
	}

	return start
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
	moves []Move,
	start int,
	precomputedInfo *PrecomputedInfo,
) int {
	p := boardState.board[sq]
	if !isPawn(p) {
		return generatePieceMoves(boardState, p, sq, offset, moves, start, precomputedInfo)
	}

	return generatePawnMoves(boardState, p, sq, offset, moves, start, precomputedInfo)
}

func GenerateMoves(boardState *BoardState, moves []Move, start int) int {
	return GenerateMoveListing(boardState, moves, start, false)
}

func generatePieceMoves(
	boardState *BoardState,
	p byte,
	sq byte,
	offset int,
	moves []Move,
	start int,
	precomputedInfo *PrecomputedInfo,
) int {
	// 8 possible directions to go (for the queen + king + knight)
	// 4 directions for bishop + rook
	// queen, bishop, and rook will continue along a "ray" until they find
	// a stopping point
	// Knight = 2, Bishop = 3, Rook = 4, Queen = 5, King = 6

	pieceType := p & 0x0F
	moveBitboards := boardState.moveBitboards

	var pieceMoves []Move

	switch pieceType {
	case KING_MASK:
		pieceMoves = moveBitboards.kingAttacks[sq].moves
	case KNIGHT_MASK:
		pieceMoves = moveBitboards.knightAttacks[sq].moves
	case BISHOP_MASK:
		magic := moveBitboards.bishopMagics[sq]
		key := hashKey(precomputedInfo.allOccupancy, magic)
		pieceMoves = moveBitboards.bishopAttacks[sq][key].moves
	case ROOK_MASK:
		magic := moveBitboards.rookMagics[sq]
		key := hashKey(precomputedInfo.allOccupancy, magic)
		pieceMoves = moveBitboards.rookAttacks[sq][key].moves
	case QUEEN_MASK:
		bishopKey := hashKey(precomputedInfo.allOccupancy, moveBitboards.bishopMagics[sq])
		pieceMoves = moveBitboards.bishopAttacks[sq][bishopKey].moves
		rookKey := hashKey(precomputedInfo.allOccupancy, moveBitboards.rookMagics[sq])
		pieceMoves = append(pieceMoves, moveBitboards.rookAttacks[sq][rookKey].moves...)
	}

	for _, move := range pieceMoves {
		oppositePiece := boardState.PieceAtSquare(move.to)
		if oppositePiece != EMPTY_SQUARE {
			if oppositePiece&0xF0 != p&0xF0 {
				move.flags |= CAPTURE_MASK
				moves[start] = move
				start++
			} else {
				// same color, just skip it
				// I think this is how we need to filter out the precomputed sliding moves
			}
		} else {
			moves[start] = move
			start++
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
				moves[start] = CreateKingsideCastle(sq, SQUARE_G1)
				start++
			}
			if boardState.boardInfo.whiteCanCastleQueenside &&
				boardState.board[SQUARE_D1] == EMPTY_SQUARE &&
				boardState.board[SQUARE_C1] == EMPTY_SQUARE &&
				boardState.board[SQUARE_B1] == EMPTY_SQUARE &&
				boardState.board[SQUARE_A1] == WHITE_MASK|ROOK_MASK {
				moves[start] = CreateQueensideCastle(sq, SQUARE_C1)
				start++
			}
		} else {
			if boardState.boardInfo.blackCanCastleKingside &&
				boardState.board[SQUARE_F8] == EMPTY_SQUARE &&
				boardState.board[SQUARE_G8] == EMPTY_SQUARE &&
				boardState.board[SQUARE_H8] == BLACK_MASK|ROOK_MASK {
				moves[start] = CreateKingsideCastle(sq, SQUARE_G8)
				start++
			}
			if boardState.boardInfo.blackCanCastleQueenside &&
				boardState.board[SQUARE_D8] == EMPTY_SQUARE &&
				boardState.board[SQUARE_C8] == EMPTY_SQUARE &&
				boardState.board[SQUARE_B8] == EMPTY_SQUARE &&
				boardState.board[SQUARE_A8] == BLACK_MASK|ROOK_MASK {
				moves[start] = CreateQueensideCastle(sq, SQUARE_C8)
				start++
			}
		}
	}

	return start
}

func generatePawnMoves(
	boardState *BoardState,
	p byte,
	sq byte,
	offset int,
	moves []Move,
	start int,
	precomputedInfo *PrecomputedInfo,
) int {
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
	captureEnd := CreateMovesFromBitboard(sq, pawnAttacks&otherOccupancies, moves, start, CAPTURE_MASK)

	for i, capture := range moves[start:captureEnd] {
		var destRank = Rank(capture.to)
		if destRank == RANK_8 || destRank == RANK_1 {
			// promotion time
			var flags byte = PROMOTION_MASK | CAPTURE_MASK
			capture.flags = flags | QUEEN_MASK
			moves[i+start] = capture

			// Now add the rest to the end
			capture.flags = flags | ROOK_MASK
			moves[captureEnd] = capture
			captureEnd++
			capture.flags = flags | BISHOP_MASK
			moves[captureEnd] = capture
			captureEnd++
			capture.flags = flags | KNIGHT_MASK
			moves[captureEnd] = capture
			captureEnd++
		} else {
			if capture.to == boardState.boardInfo.enPassantTargetSquare {
				capture.flags |= SPECIAL1_MASK | CAPTURE_MASK
				moves[i+start] = capture
			}
		}
	}
	start = captureEnd

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
			moves[start] = CreatePromotion(sq, dest, QUEEN_MASK)
			start++
			moves[start] = CreatePromotion(sq, dest, BISHOP_MASK)
			start++
			moves[start] = CreatePromotion(sq, dest, KNIGHT_MASK)
			start++
			moves[start] = CreatePromotion(sq, dest, ROOK_MASK)
			start++
		} else {
			// empty square
			moves[start] = CreateMove(sq, dest)
			start++

			if (isWhite && sourceRank == RANK_2) ||
				(!isWhite && sourceRank == RANK_7) {
				// home row for white so we can move one more
				dest = uint8(int8(dest) + sqOffset)

				if boardState.board[dest] == EMPTY_SQUARE {
					moves[start] = CreateMove(sq, dest)
					start++
				}
			}
		}
	}

	return start
}
