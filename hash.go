package main

import (
	"fmt"
	"math/rand"
)

var _ = fmt.Println

type HashInfo struct {
	// this is a sparse array - will be indexed by offset and piece type
	content                 [255][255]uint64
	whiteToMove             uint64
	whiteCanCastleKingside  uint64
	whiteCanCastleQueenside uint64
	blackCanCastleKingside  uint64
	blackCanCastleQueenside uint64
	// do we need the en passant target square here or is it ok if they collide (?)
}

func CreateHashInfo(r *rand.Rand) HashInfo {
	var hashInfo HashInfo

	for i := byte(0); i < 8; i++ {
		for j := byte(0); j < 8; j++ {
			sq := RowAndColToSquare(i, j)
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
	hashInfo.whiteToMove = r.Uint64()
	hashInfo.whiteCanCastleKingside = r.Uint64()
	hashInfo.whiteCanCastleQueenside = r.Uint64()
	hashInfo.blackCanCastleKingside = r.Uint64()
	hashInfo.blackCanCastleQueenside = r.Uint64()

	return hashInfo
}

func (boardState *BoardState) CreateHashKey(info *HashInfo) uint64 {
	var key uint64 = 0

	if boardState.whiteToMove {
		key ^= info.whiteToMove
	}
	for i := byte(0); i < 8; i++ {
		for j := byte(0); j < 8; j++ {
			sq := RowAndColToSquare(i, j)
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

	return key
}

// To be applied after a move has been made, to incrementally update the hash key.
// For use in search
func (boardState *BoardState) UpdateHashApplyMove(key uint64, info *HashInfo, oldBoardInfo BoardInfo, move Move) uint64 {
	if move.IsCapture() {
		capturePiece := boardState.captureStack.Peek()
		key ^= info.content[move.to][capturePiece]
	}

	movePiece := boardState.board[move.to]
	key ^= info.content[move.from][movePiece]
	key ^= info.content[move.to][movePiece]

	// if move is a castle, move work
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

	return key
}

func (boardState *BoardState) UpdateHashUnapplyMove(key uint64, info *HashInfo, oldBoardInfo BoardInfo, move Move) uint64 {
	if move.IsCapture() {
		// we've already put back the piece since this is done after move is unapplied
		key ^= info.content[move.to][boardState.board[move.to]]
	}

	movePiece := boardState.board[move.from]
	key ^= info.content[move.to][movePiece]
	key ^= info.content[move.from][movePiece]

	// if move is a castle, move work
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

	return key
}
