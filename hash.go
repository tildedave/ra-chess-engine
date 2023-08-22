package main

import (
	"fmt"
	"math/rand"
)

var _ = fmt.Println

type HashInfo struct {
	// this is a sparse array - will be indexed by offset and piece type
	content                 [255][255]uint64
	enpassant               [255]uint64 // indexed by offset
	sideToMove              uint64
	whiteCanCastleKingside  uint64
	whiteCanCastleQueenside uint64
	blackCanCastleKingside  uint64
	blackCanCastleQueenside uint64
}

func CreateHashInfo(r *rand.Rand) HashInfo {
	var hashInfo HashInfo

	for i := byte(0); i < 8; i++ {
		for j := byte(0); j < 8; j++ {
			sq := idx(j, i)
			for _, m := range []byte{WHITE_MASK, BLACK_MASK} {
				hashInfo.content[sq][m|PAWN_MASK] = r.Uint64()
				hashInfo.content[sq][m|KNIGHT_MASK] = r.Uint64()
				hashInfo.content[sq][m|BISHOP_MASK] = r.Uint64()
				hashInfo.content[sq][m|ROOK_MASK] = r.Uint64()
				hashInfo.content[sq][m|QUEEN_MASK] = r.Uint64()
				hashInfo.content[sq][m|KING_MASK] = r.Uint64()
			}
		}
	}
	for i := byte(0); i < 8; i++ {
		hashInfo.enpassant[idx(i, 3)] = r.Uint64()
		hashInfo.enpassant[idx(i, 6)] = r.Uint64()
	}
	// target square 0 is used for a 'clear' EP target
	hashInfo.enpassant[0] = r.Uint64()
	hashInfo.sideToMove = r.Uint64()
	hashInfo.whiteCanCastleKingside = r.Uint64()
	hashInfo.whiteCanCastleQueenside = r.Uint64()
	hashInfo.blackCanCastleKingside = r.Uint64()
	hashInfo.blackCanCastleQueenside = r.Uint64()

	return hashInfo
}

func (boardState *BoardState) CreateHashKey(info *HashInfo) uint64 {
	var key uint64

	if boardState.sideToMove == WHITE_OFFSET {
		key ^= info.sideToMove
	}
	for i := byte(0); i < 8; i++ {
		for j := byte(0); j < 8; j++ {
			sq := idx(j, i)
			key ^= info.content[sq][boardState.board[sq]]
		}
	}
	if boardState.boardInfo.whiteCanCastleKingside {
		key ^= info.whiteCanCastleKingside
	}
	if boardState.boardInfo.whiteCanCastleQueenside {
		key ^= info.whiteCanCastleQueenside
	}
	if boardState.boardInfo.blackCanCastleKingside {
		key ^= info.blackCanCastleKingside
	}
	if boardState.boardInfo.blackCanCastleQueenside {
		key ^= info.blackCanCastleQueenside
	}
	if boardState.boardInfo.enPassantTargetSquare != 0x00 {
		key ^= info.enpassant[boardState.boardInfo.enPassantTargetSquare]
	}

	return key
}

func (boardState *BoardState) CreatePawnHashKey(info *HashInfo) uint64 {
	var key uint64

	for i := byte(0); i < 8; i++ {
		for j := byte(0); j < 8; j++ {
			sq := idx(j, i)
			pieceAtSquare := boardState.board[sq]
			if pieceAtSquare&0x0F == PAWN_MASK {
				key ^= info.content[sq][boardState.board[sq]]
			}
		}
	}

	return key
}

// To be applied after a move has been made, to incrementally update the hash key.
// For use in search
func (boardState *BoardState) UpdateHashApplyMove(oldBoardInfo BoardInfo, move Move, isCapture bool) {
	key := boardState.hashKey
	pawnKey := boardState.pawnHashKey

	info := boardState.hashInfo

	if isCapture {
		if move.IsEnPassantCapture() {
			var pos uint8
			var capturePiece byte
			if boardState.sideToMove == BLACK_OFFSET {
				pos = move.To() - 8
				capturePiece = BLACK_MASK | PAWN_MASK
			} else {
				pos = move.To() + 8
				capturePiece = WHITE_MASK | PAWN_MASK
			}
			key ^= info.content[pos][capturePiece]
			pawnKey ^= info.content[pos][capturePiece]
		} else {
			capturePiece := boardState.captureStack.Peek()
			update := info.content[move.To()][capturePiece]
			key ^= update
			if capturePiece&0x0F == PAWN_MASK {
				pawnKey ^= update
			}
		}
	}

	var toPiece uint8
	var fromPiece uint8

	if move.IsPromotion() {
		colorMask := sideToMoveToColorMask(oppositeColorOffset(boardState.sideToMove))
		pieceMask := move.GetPromotionPiece()
		toPiece = colorMask | pieceMask
		fromPiece = colorMask | PAWN_MASK
	} else {
		toPiece = boardState.board[move.To()]
		fromPiece = toPiece
	}

	key ^= info.content[move.To()][toPiece]
	key ^= info.content[move.From()][fromPiece]
	key ^= info.sideToMove

	if fromPiece&0x0F == PAWN_MASK {
		pawnKey ^= info.content[move.From()][fromPiece]
	}
	if toPiece&0x0F == PAWN_MASK {
		pawnKey ^= info.content[move.To()][toPiece]
	}

	if boardState.sideToMove == BLACK_OFFSET {
		if move.IsKingsideCastle() {
			key ^= info.content[SQUARE_H1][WHITE_MASK|ROOK_MASK]
			key ^= info.content[SQUARE_F1][WHITE_MASK|ROOK_MASK]
		} else if move.IsQueensideCastle() {
			key ^= info.content[SQUARE_A1][WHITE_MASK|ROOK_MASK]
			key ^= info.content[SQUARE_D1][WHITE_MASK|ROOK_MASK]
		}
		if oldBoardInfo.whiteCanCastleKingside != boardState.boardInfo.whiteCanCastleKingside {
			key ^= info.whiteCanCastleKingside
		}
		if oldBoardInfo.whiteCanCastleQueenside != boardState.boardInfo.whiteCanCastleQueenside {
			key ^= info.whiteCanCastleQueenside
		}
	} else {
		if move.IsKingsideCastle() {
			key ^= info.content[SQUARE_H8][BLACK_MASK|ROOK_MASK]
			key ^= info.content[SQUARE_F8][BLACK_MASK|ROOK_MASK]
		} else if move.IsQueensideCastle() {
			key ^= info.content[SQUARE_A8][BLACK_MASK|ROOK_MASK]
			key ^= info.content[SQUARE_D8][BLACK_MASK|ROOK_MASK]
		}

		if oldBoardInfo.blackCanCastleKingside != boardState.boardInfo.blackCanCastleKingside {
			key ^= info.blackCanCastleKingside
		}
		if oldBoardInfo.blackCanCastleQueenside != boardState.boardInfo.blackCanCastleQueenside {
			key ^= info.blackCanCastleQueenside
		}
	}
	if oldBoardInfo.enPassantTargetSquare != boardState.boardInfo.enPassantTargetSquare {
		key ^= info.enpassant[oldBoardInfo.enPassantTargetSquare]
		key ^= info.enpassant[boardState.boardInfo.enPassantTargetSquare]
	}

	boardState.hashKey = key
	boardState.pawnHashKey = pawnKey
}

// UpdateHashUnapplyMove gives a new hash key as a result of unapplying a move.
// It should be called _after_ the move has been unprocessed, and the oldBoardInfo
// should be the board info prior to the move being unprocessed.
func (boardState *BoardState) UpdateHashUnapplyMove(oldBoardInfo BoardInfo, move Move, isCapture bool) {
	key := boardState.hashKey
	pawnKey := boardState.pawnHashKey
	info := boardState.hashInfo
	if isCapture {
		// we've already put back the piece since this is done after move is unapplied
		if move.IsEnPassantCapture() {
			var pos uint8
			if boardState.sideToMove == WHITE_OFFSET {
				pos = move.To() - 8
			} else {
				pos = move.To() + 8
			}
			update := info.content[pos][boardState.board[pos]]
			key ^= update
			pawnKey ^= update
		} else {
			capturedPiece := boardState.board[move.To()]
			update := info.content[move.To()][capturedPiece]
			key ^= update
			if capturedPiece&0x0F == PAWN_MASK {
				pawnKey ^= update
			}
		}
	}

	var toPiece uint8
	var fromPiece uint8

	if move.IsPromotion() {
		colorMask := sideToMoveToColorMask(boardState.sideToMove)
		pieceMask := move.GetPromotionPiece()
		toPiece = colorMask | pieceMask
		fromPiece = colorMask | PAWN_MASK
	} else {
		toPiece = boardState.board[move.From()]
		fromPiece = toPiece
	}

	key ^= info.content[move.To()][toPiece]
	key ^= info.content[move.From()][fromPiece]
	key ^= info.sideToMove
	if toPiece&0x0F == PAWN_MASK {
		pawnKey ^= info.content[move.To()][toPiece]
	}
	if fromPiece&0x0F == PAWN_MASK {
		pawnKey ^= info.content[move.From()][fromPiece]
	}

	if boardState.sideToMove == WHITE_OFFSET {
		if move.IsKingsideCastle() {
			key ^= info.content[SQUARE_H1][WHITE_MASK|ROOK_MASK]
			key ^= info.content[SQUARE_F1][WHITE_MASK|ROOK_MASK]
		} else if move.IsQueensideCastle() {
			key ^= info.content[SQUARE_A1][WHITE_MASK|ROOK_MASK]
			key ^= info.content[SQUARE_D1][WHITE_MASK|ROOK_MASK]
		}
		if oldBoardInfo.whiteCanCastleKingside != boardState.boardInfo.whiteCanCastleKingside {
			key ^= info.whiteCanCastleKingside
		}
		if oldBoardInfo.whiteCanCastleQueenside != boardState.boardInfo.whiteCanCastleQueenside {
			key ^= info.whiteCanCastleQueenside
		}
	} else {
		if move.IsKingsideCastle() {
			key ^= info.content[SQUARE_H8][BLACK_MASK|ROOK_MASK]
			key ^= info.content[SQUARE_F8][BLACK_MASK|ROOK_MASK]
		} else if move.IsQueensideCastle() {
			key ^= info.content[SQUARE_A8][BLACK_MASK|ROOK_MASK]
			key ^= info.content[SQUARE_D8][BLACK_MASK|ROOK_MASK]
		}

		if oldBoardInfo.blackCanCastleKingside != boardState.boardInfo.blackCanCastleKingside {
			key ^= info.blackCanCastleKingside
		}
		if oldBoardInfo.blackCanCastleQueenside != boardState.boardInfo.blackCanCastleQueenside {
			key ^= info.blackCanCastleQueenside
		}
	}

	if oldBoardInfo.enPassantTargetSquare != boardState.boardInfo.enPassantTargetSquare {
		key ^= info.enpassant[oldBoardInfo.enPassantTargetSquare]
		key ^= info.enpassant[boardState.boardInfo.enPassantTargetSquare]
	}

	boardState.hashKey = key
	boardState.pawnHashKey = pawnKey
}
