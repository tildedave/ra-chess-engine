package main

import (
	"errors"
	"fmt"
	"strconv"
)

var _ = fmt.Println

func (boardState *BoardState) ApplyMove(move Move) {
	boardState.boardInfoHistory[boardState.moveIndex] = boardState.boardInfo

	var capturedPiece byte
	if move.IsCapture() {
		capturedPiece = boardState.board[move.to]
		boardState.captureStack.Push(capturedPiece)
	}

	var p = boardState.board[move.from]
	boardState.board[move.from] = 0x00
	boardState.board[move.to] = p
	// unset

	var offset int
	var otherOffset int
	switch boardState.whiteToMove {
	case true:
		offset = WHITE_OFFSET
		otherOffset = BLACK_OFFSET
	case false:
		offset = BLACK_OFFSET
		otherOffset = WHITE_OFFSET
	}

	boardState.bitboards.piece[p&0x0F] = SetBitboard(
		UnsetBitboard(boardState.bitboards.piece[p&0x0F], legacySquareToBitboardSquare(move.from)),
		legacySquareToBitboardSquare(move.to))

	boardState.bitboards.color[offset] = SetBitboard(
		UnsetBitboard(boardState.bitboards.color[offset], legacySquareToBitboardSquare(move.from)),
		legacySquareToBitboardSquare(move.to))

	if move.IsCapture() {
		boardState.bitboards.color[otherOffset] = UnsetBitboard(boardState.bitboards.color[otherOffset], legacySquareToBitboardSquare(move.to))
		boardState.bitboards.piece[capturedPiece&0x0F] = UnsetBitboard(boardState.bitboards.piece[capturedPiece&0x0F], legacySquareToBitboardSquare(move.to))
	}

	var epTargetSquare = boardState.boardInfo.enPassantTargetSquare

	// TODO(perf) - less if statements/work when castling is over
	boardState.boardInfo.enPassantTargetSquare = 0

	if move.IsQueensideCastle() {
		// white
		if boardState.whiteToMove {
			boardState.board[SQUARE_A1] = EMPTY_SQUARE
			boardState.board[SQUARE_D1] = WHITE_MASK | ROOK_MASK

			boardState.bitboards.piece[BITBOARD_ROOK_OFFSET] = SetBitboard(
				UnsetBitboard(boardState.bitboards.piece[BITBOARD_ROOK_OFFSET], BB_SQUARE_A1),
				BB_SQUARE_D1)
			boardState.bitboards.color[WHITE_OFFSET] = SetBitboard(
				UnsetBitboard(boardState.bitboards.color[WHITE_OFFSET], BB_SQUARE_A1),
				BB_SQUARE_D1)

			boardState.boardInfo.whiteCanCastleKingside = false
			boardState.boardInfo.whiteCanCastleQueenside = false
			boardState.lookupInfo.whiteKingSquare = SQUARE_C1
			boardState.boardInfo.whiteHasCastled = true
		} else {
			boardState.board[SQUARE_A8] = EMPTY_SQUARE
			boardState.board[SQUARE_D8] = BLACK_MASK | ROOK_MASK

			boardState.bitboards.piece[BITBOARD_ROOK_OFFSET] = SetBitboard(
				UnsetBitboard(boardState.bitboards.piece[BITBOARD_ROOK_OFFSET], BB_SQUARE_A8),
				BB_SQUARE_D8)
			boardState.bitboards.color[BLACK_OFFSET] = SetBitboard(
				UnsetBitboard(boardState.bitboards.color[BLACK_OFFSET], BB_SQUARE_A8),
				BB_SQUARE_D8)

			boardState.boardInfo.blackCanCastleKingside = false
			boardState.boardInfo.blackCanCastleQueenside = false
			boardState.lookupInfo.blackKingSquare = SQUARE_C8
			boardState.boardInfo.blackHasCastled = true
		}
	} else if move.IsKingsideCastle() {
		if boardState.whiteToMove {
			boardState.board[SQUARE_H1] = EMPTY_SQUARE
			boardState.board[SQUARE_F1] = WHITE_MASK | ROOK_MASK

			boardState.bitboards.piece[BITBOARD_ROOK_OFFSET] = SetBitboard(
				UnsetBitboard(boardState.bitboards.piece[BITBOARD_ROOK_OFFSET], BB_SQUARE_H1),
				BB_SQUARE_F1)
			boardState.bitboards.color[WHITE_OFFSET] = SetBitboard(
				UnsetBitboard(boardState.bitboards.color[WHITE_OFFSET], BB_SQUARE_H1),
				BB_SQUARE_F1)

			boardState.boardInfo.whiteCanCastleKingside = false
			boardState.boardInfo.whiteCanCastleQueenside = false
			boardState.lookupInfo.whiteKingSquare = SQUARE_G1
			boardState.boardInfo.whiteHasCastled = true
		} else {
			boardState.board[SQUARE_H8] = EMPTY_SQUARE
			boardState.board[SQUARE_F8] = BLACK_MASK | ROOK_MASK

			boardState.bitboards.piece[BITBOARD_ROOK_OFFSET] = SetBitboard(
				UnsetBitboard(boardState.bitboards.piece[BITBOARD_ROOK_OFFSET], BB_SQUARE_H8),
				BB_SQUARE_F8)
			boardState.bitboards.color[BLACK_OFFSET] = SetBitboard(
				UnsetBitboard(boardState.bitboards.color[BLACK_OFFSET], BB_SQUARE_H8),
				BB_SQUARE_F8)

			boardState.boardInfo.blackCanCastleKingside = false
			boardState.boardInfo.blackCanCastleQueenside = false
			boardState.lookupInfo.blackKingSquare = SQUARE_G8
			boardState.boardInfo.blackHasCastled = true
		}
	} else {
		switch p & 0x0F {
		case KING_MASK:
			if boardState.whiteToMove {
				boardState.boardInfo.whiteCanCastleKingside = false
				boardState.boardInfo.whiteCanCastleQueenside = false
				boardState.lookupInfo.whiteKingSquare = move.to
			} else {
				boardState.boardInfo.blackCanCastleKingside = false
				boardState.boardInfo.blackCanCastleQueenside = false
				boardState.lookupInfo.blackKingSquare = move.to
			}
		case ROOK_MASK:
			if boardState.whiteToMove {
				if move.from == SQUARE_H1 {
					boardState.boardInfo.whiteCanCastleKingside = false
				} else if move.from == SQUARE_A1 {
					boardState.boardInfo.whiteCanCastleQueenside = false
				}
			} else {
				if move.from == SQUARE_H8 {
					boardState.boardInfo.blackCanCastleKingside = false
				} else if move.from == SQUARE_A8 {
					boardState.boardInfo.blackCanCastleQueenside = false
				}
			}
		case PAWN_MASK:
			if !move.IsCapture() {
				if move.to > move.from {
					if move.to-move.from > 10 {
						boardState.boardInfo.enPassantTargetSquare = move.from + 10
					}
				} else if move.from-move.to > 10 {
					boardState.boardInfo.enPassantTargetSquare = move.from - 10
				}
			} else if move.to == epTargetSquare {
				var pos uint8
				if boardState.whiteToMove {
					pos = move.to - 10
				} else {
					pos = move.to + 10
				}

				// captureStack is wrong in this case (it has a 0 on it) so we need to fix it
				// better to do this check now than above because EP is not common
				boardState.captureStack.Pop()
				boardState.captureStack.Push(boardState.board[pos])
				boardState.board[pos] = 0x00
			}

			if move.IsPromotion() {
				boardState.board[move.to] = move.GetPromotionPiece() | whiteToMoveToColorMask(boardState.whiteToMove)
			}
		}
	}

	boardState.whiteToMove = !boardState.whiteToMove
	oldBoardInfo := boardState.boardInfoHistory[boardState.moveIndex]
	boardState.moveIndex++

	if boardState.whiteToMove {
		boardState.fullmoveNumber++
	}

	boardState.hashKey = boardState.UpdateHashApplyMove(boardState.hashKey, oldBoardInfo, move)
}

func (boardState *BoardState) IsMoveLegal(move Move) (bool, error) {
	fromPiece := boardState.board[move.from]
	toPiece := boardState.board[move.to]
	var pieceMask byte
	var captureMask byte

	if boardState.whiteToMove {
		pieceMask = WHITE_MASK
		captureMask = BLACK_MASK
	} else {
		pieceMask = BLACK_MASK
		captureMask = WHITE_MASK
	}

	if fromPiece == EMPTY_SQUARE {
		return false, errors.New("From square was empty")
	} else if fromPiece == SENTINEL_MASK {
		return false, errors.New("From square was sentinel")
	} else if fromPiece&pieceMask != pieceMask {
		return false, errors.New("From square was not occupied by expected piece: " + strconv.Itoa(int(fromPiece)))
	}

	if !move.IsCapture() {
		if toPiece != EMPTY_SQUARE {
			return false, errors.New("To square was occupied for non-capture")
		}
	} else {
		if toPiece == EMPTY_SQUARE {
			if !(isPawn(fromPiece) && move.to == boardState.boardInfo.enPassantTargetSquare) {
				return false, errors.New("To square was empty for capture")
			}
		} else if toPiece == SENTINEL_MASK {
			return false, errors.New("To square was a sentinel")
		} else if toPiece&captureMask != captureMask {
			return false, errors.New("To square was not occupied by piece of correct color: " + strconv.Itoa(int(toPiece)))
		}
	}

	// TODO: ensure side to move not in check
	// TODO: ensure castling isn't moving through check

	return true, nil
}

func (boardState *BoardState) UnapplyMove(move Move) {
	boardState.whiteToMove = !boardState.whiteToMove
	if !boardState.whiteToMove {
		boardState.fullmoveNumber--
	}
	oldBoardInfo := boardState.boardInfo
	boardState.moveIndex--
	boardState.boardInfo = boardState.boardInfoHistory[boardState.moveIndex]

	var p = boardState.board[move.to]

	var capturedPiece byte
	if move.IsCapture() {
		capturedPiece = boardState.captureStack.Pop()
		boardState.board[move.to] = capturedPiece
	} else {
		boardState.board[move.to] = 0x00
	}

	boardState.board[move.from] = p

	var offset int
	var otherOffset int
	switch boardState.whiteToMove {
	case true:
		offset = WHITE_OFFSET
		otherOffset = BLACK_OFFSET
	case false:
		offset = BLACK_OFFSET
		otherOffset = WHITE_OFFSET
	}

	boardState.bitboards.piece[p&0x0F] = SetBitboard(boardState.bitboards.piece[p&0x0F], legacySquareToBitboardSquare(move.from))
	boardState.bitboards.color[offset] = SetBitboard(boardState.bitboards.color[offset], legacySquareToBitboardSquare(move.from))
	boardState.bitboards.piece[p&0x0F] = UnsetBitboard(boardState.bitboards.piece[p&0x0F], legacySquareToBitboardSquare(move.to))
	boardState.bitboards.color[offset] = UnsetBitboard(boardState.bitboards.color[offset], legacySquareToBitboardSquare(move.to))
	if move.IsCapture() {
		boardState.bitboards.color[otherOffset] = SetBitboard(boardState.bitboards.color[otherOffset], legacySquareToBitboardSquare(move.to))
		boardState.bitboards.piece[capturedPiece&0x0F] = SetBitboard(boardState.bitboards.piece[capturedPiece&0x0F], legacySquareToBitboardSquare(move.to))
	}

	// TODO(perf) - just switch statement on the different conditions here, they are all mutually exclusive
	if move.IsQueensideCastle() {
		// black was to move, so we're unmaking a white move
		if boardState.whiteToMove {
			boardState.board[SQUARE_D1] = 0x00
			boardState.board[SQUARE_A1] = WHITE_MASK | ROOK_MASK

			boardState.bitboards.piece[BITBOARD_ROOK_OFFSET] = SetBitboard(
				UnsetBitboard(boardState.bitboards.piece[BITBOARD_ROOK_OFFSET], BB_SQUARE_D1),
				BB_SQUARE_A1)
			boardState.bitboards.color[WHITE_OFFSET] = SetBitboard(
				UnsetBitboard(boardState.bitboards.color[WHITE_OFFSET], BB_SQUARE_D1),
				BB_SQUARE_A1)

			boardState.boardInfo.whiteCanCastleQueenside = true
			boardState.boardInfo.whiteHasCastled = false
		} else {
			boardState.board[SQUARE_D8] = 0x00
			boardState.board[SQUARE_A8] = BLACK_MASK | ROOK_MASK

			boardState.bitboards.piece[BITBOARD_ROOK_OFFSET] = SetBitboard(
				UnsetBitboard(boardState.bitboards.piece[BITBOARD_ROOK_OFFSET], BB_SQUARE_D8),
				BB_SQUARE_A8)
			boardState.bitboards.color[BLACK_OFFSET] = SetBitboard(
				UnsetBitboard(boardState.bitboards.color[BLACK_OFFSET], BB_SQUARE_D8),
				BB_SQUARE_A8)

			boardState.boardInfo.blackCanCastleQueenside = true
			boardState.boardInfo.blackHasCastled = false
		}
	} else if move.IsKingsideCastle() {
		if boardState.whiteToMove {
			boardState.board[SQUARE_F1] = 0x00
			boardState.board[SQUARE_H1] = WHITE_MASK | ROOK_MASK

			boardState.bitboards.piece[BITBOARD_ROOK_OFFSET] = SetBitboard(
				UnsetBitboard(boardState.bitboards.piece[BITBOARD_ROOK_OFFSET], BB_SQUARE_F1),
				BB_SQUARE_H1)
			boardState.bitboards.color[WHITE_OFFSET] = SetBitboard(
				UnsetBitboard(boardState.bitboards.color[WHITE_OFFSET], BB_SQUARE_F1),
				BB_SQUARE_H1)

			boardState.boardInfo.whiteCanCastleKingside = true
			boardState.boardInfo.whiteHasCastled = false
		} else {
			boardState.board[SQUARE_F8] = 0x00
			boardState.board[SQUARE_H8] = BLACK_MASK | ROOK_MASK

			boardState.bitboards.piece[BITBOARD_ROOK_OFFSET] = SetBitboard(
				UnsetBitboard(boardState.bitboards.piece[BITBOARD_ROOK_OFFSET], BB_SQUARE_F8),
				BB_SQUARE_H8)
			boardState.bitboards.color[BLACK_OFFSET] = SetBitboard(
				UnsetBitboard(boardState.bitboards.color[BLACK_OFFSET], BB_SQUARE_F8),
				BB_SQUARE_H8)

			boardState.boardInfo.blackCanCastleKingside = true
			boardState.boardInfo.blackHasCastled = false
		}
	}

	switch p & 0x0F {
	case KING_MASK:
		if boardState.whiteToMove {
			boardState.lookupInfo.whiteKingSquare = move.from
		} else {
			boardState.lookupInfo.blackKingSquare = move.from
		}
	case PAWN_MASK:
		if move.IsEnPassantCapture() {
			var pos uint8
			if boardState.whiteToMove {
				pos = move.to - 10
			} else {
				pos = move.to + 10
			}

			boardState.board[pos] = boardState.board[move.to]
			boardState.board[move.to] = 0x00
			boardState.boardInfo.enPassantTargetSquare = move.to
		}
	}

	if move.IsPromotion() {
		var mask byte
		if boardState.whiteToMove {
			mask = WHITE_MASK
		} else {
			mask = BLACK_MASK
		}
		boardState.board[move.from] = mask | PAWN_MASK
	}

	boardState.hashKey = boardState.UpdateHashUnapplyMove(boardState.hashKey, oldBoardInfo, move)
}
