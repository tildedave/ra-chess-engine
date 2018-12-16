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
	whiteToMove             uint64
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
	hashInfo.whiteToMove = r.Uint64()
	hashInfo.whiteCanCastleKingside = r.Uint64()
	hashInfo.whiteCanCastleQueenside = r.Uint64()
	hashInfo.blackCanCastleKingside = r.Uint64()
	hashInfo.blackCanCastleQueenside = r.Uint64()

	return hashInfo
}

func (boardState *BoardState) CreateHashKey(info *HashInfo) uint64 {
	var key uint64

	if boardState.whiteToMove {
		key ^= info.whiteToMove
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

// To be applied after a move has been made, to incrementally update the hash key.
// For use in search
func (boardState *BoardState) UpdateHashApplyMove(key uint64, oldBoardInfo BoardInfo, move Move) uint64 {
	info := boardState.hashInfo

	if move.IsCapture() {
		capturePiece := boardState.captureStack.Peek()

		if move.IsEnPassantCapture() {
			var pos uint8
			if !boardState.whiteToMove {
				pos = move.to - 8
			} else {
				pos = move.to + 8
			}
			key ^= info.content[pos][capturePiece]
		} else {
			key ^= info.content[move.to][capturePiece]
		}
	}

	var toPiece uint8
	var fromPiece uint8

	if move.IsPromotion() {
		colorMask := whiteToMoveToColorMask(!boardState.whiteToMove)
		pieceMask := move.GetPromotionPiece()
		toPiece = colorMask | pieceMask
		fromPiece = colorMask | PAWN_MASK
	} else {
		toPiece = boardState.board[move.to]
		fromPiece = toPiece
	}

	key ^= info.content[move.to][toPiece]
	key ^= info.content[move.from][fromPiece]
	key ^= info.whiteToMove

	if !boardState.whiteToMove {
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

	return key
}

// UpdateHashUnapplyMove gives a new hash key as a result of unapplying a move.
// It should be called _after_ the move has been unprocessed, and the oldBoardInfo
// should be the board info prior to the move being unprocessed.
func (boardState *BoardState) UpdateHashUnapplyMove(key uint64, oldBoardInfo BoardInfo, move Move) uint64 {
	info := boardState.hashInfo
	if move.IsCapture() {
		// we've already put back the piece since this is done after move is unapplied
		if move.IsEnPassantCapture() {
			var pos uint8
			if boardState.whiteToMove {
				pos = move.to - 8
			} else {
				pos = move.to + 8
			}
			key ^= info.content[pos][boardState.board[pos]]
		} else {
			key ^= info.content[move.to][boardState.board[move.to]]
		}
	}

	var toPiece uint8
	var fromPiece uint8

	if move.IsPromotion() {
		colorMask := whiteToMoveToColorMask(boardState.whiteToMove)
		pieceMask := move.GetPromotionPiece()
		toPiece = colorMask | pieceMask
		fromPiece = colorMask | PAWN_MASK
	} else {
		toPiece = boardState.board[move.from]
		fromPiece = toPiece
	}

	key ^= info.content[move.to][toPiece]
	key ^= info.content[move.from][fromPiece]
	key ^= info.whiteToMove

	if boardState.whiteToMove {
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

	return key
}
