package main

import (
	"errors"
	"strconv"
)

var errMoveUninitialized error = errors.New("Uninitialized move")

func (boardState *BoardState) ApplyMove(move Move) {
	boardState.boardInfoHistory[boardState.moveIndex] = boardState.boardInfo

	capturedPiece := boardState.board[move.to]
	isCapture := capturedPiece != EMPTY_SQUARE
	boardState.wasCapture[boardState.moveIndex+1] = isCapture
	if isCapture {
		boardState.captureStack.Push(capturedPiece)
	}

	var p = boardState.board[move.from]
	var movePiece = p & 0x0F
	boardState.board[move.from] = 0x00
	boardState.board[move.to] = p
	// unset

	var offset = boardState.sideToMove
	var otherOffset = oppositeColorOffset(boardState.sideToMove)

	if capturedPiece != EMPTY_SQUARE {
		boardState.bitboards.color[otherOffset] = FlipBitboard(boardState.bitboards.color[otherOffset], move.to)
		boardState.bitboards.piece[capturedPiece&0x0F] = FlipBitboard(boardState.bitboards.piece[capturedPiece&0x0F], move.to)
	}

	boardState.bitboards.piece[movePiece] = FlipBitboard2(boardState.bitboards.piece[movePiece], move.from, move.to)
	boardState.bitboards.color[offset] = FlipBitboard2(boardState.bitboards.color[offset], move.from, move.to)

	// TODO(perf) - less if statements/work when castling is over
	boardState.boardInfo.enPassantTargetSquare = 0

	if move.IsQueensideCastle() {
		// white
		if boardState.sideToMove == WHITE_OFFSET {
			boardState.board[SQUARE_A1] = EMPTY_SQUARE
			boardState.board[SQUARE_D1] = WHITE_MASK | ROOK_MASK

			boardState.bitboards.piece[BITBOARD_ROOK_OFFSET] = FlipBitboard2(
				boardState.bitboards.piece[BITBOARD_ROOK_OFFSET],
				SQUARE_A1,
				SQUARE_D1)
			boardState.bitboards.color[WHITE_OFFSET] = FlipBitboard2(
				boardState.bitboards.color[WHITE_OFFSET],
				SQUARE_A1,
				SQUARE_D1)

			boardState.boardInfo.whiteCanCastleKingside = false
			boardState.boardInfo.whiteCanCastleQueenside = false
			boardState.boardInfo.whiteHasCastled = true
		} else {
			boardState.board[SQUARE_A8] = EMPTY_SQUARE
			boardState.board[SQUARE_D8] = BLACK_MASK | ROOK_MASK

			boardState.bitboards.piece[BITBOARD_ROOK_OFFSET] = FlipBitboard2(
				boardState.bitboards.piece[BITBOARD_ROOK_OFFSET],
				SQUARE_A8,
				SQUARE_D8)
			boardState.bitboards.color[BLACK_OFFSET] = FlipBitboard2(
				boardState.bitboards.color[BLACK_OFFSET],
				SQUARE_A8,
				SQUARE_D8)

			boardState.boardInfo.blackCanCastleKingside = false
			boardState.boardInfo.blackCanCastleQueenside = false
			boardState.boardInfo.blackHasCastled = true
		}
	} else if move.IsKingsideCastle() {
		if boardState.sideToMove == WHITE_OFFSET {
			boardState.board[SQUARE_H1] = EMPTY_SQUARE
			boardState.board[SQUARE_F1] = WHITE_MASK | ROOK_MASK

			boardState.bitboards.piece[BITBOARD_ROOK_OFFSET] = FlipBitboard2(
				boardState.bitboards.piece[BITBOARD_ROOK_OFFSET],
				SQUARE_H1,
				SQUARE_F1)
			boardState.bitboards.color[WHITE_OFFSET] = FlipBitboard2(
				boardState.bitboards.color[WHITE_OFFSET],
				SQUARE_H1,
				SQUARE_F1)

			boardState.boardInfo.whiteCanCastleKingside = false
			boardState.boardInfo.whiteCanCastleQueenside = false
			boardState.boardInfo.whiteHasCastled = true
		} else {
			boardState.board[SQUARE_H8] = EMPTY_SQUARE
			boardState.board[SQUARE_F8] = BLACK_MASK | ROOK_MASK

			boardState.bitboards.piece[BITBOARD_ROOK_OFFSET] = FlipBitboard2(
				boardState.bitboards.piece[BITBOARD_ROOK_OFFSET],
				SQUARE_H8,
				SQUARE_F8)
			boardState.bitboards.color[BLACK_OFFSET] = FlipBitboard2(
				boardState.bitboards.color[BLACK_OFFSET],
				SQUARE_H8,
				SQUARE_F8)

			boardState.boardInfo.blackCanCastleKingside = false
			boardState.boardInfo.blackCanCastleQueenside = false
			boardState.boardInfo.blackHasCastled = true
		}
	} else {
		switch movePiece {
		case KING_MASK:
			if boardState.sideToMove == WHITE_OFFSET {
				boardState.boardInfo.whiteCanCastleKingside = false
				boardState.boardInfo.whiteCanCastleQueenside = false
			} else {
				boardState.boardInfo.blackCanCastleKingside = false
				boardState.boardInfo.blackCanCastleQueenside = false
			}
		case ROOK_MASK:
			if boardState.sideToMove == WHITE_OFFSET {
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
			if move.IsEnPassantCapture() {
				var pos uint8
				var otherOffset int
				if boardState.sideToMove == WHITE_OFFSET {
					pos = move.to - 8
					otherOffset = BLACK_OFFSET
				} else {
					pos = move.to + 8
					otherOffset = WHITE_OFFSET
				}

				isCapture = true
				// logic above didn't do anything since the move wasn't determined
				// as a capture

				boardState.board[pos] = 0x00
				boardState.bitboards.color[otherOffset] = FlipBitboard(
					boardState.bitboards.color[otherOffset],
					pos)
				boardState.bitboards.piece[BITBOARD_PAWN_OFFSET] = FlipBitboard(
					boardState.bitboards.piece[BITBOARD_PAWN_OFFSET],
					pos)
			} else if capturedPiece == EMPTY_SQUARE {
				if move.to > move.from {
					if move.to-move.from > 8 {
						boardState.boardInfo.enPassantTargetSquare = move.from + 8
					}
				} else if move.from-move.to > 8 {
					boardState.boardInfo.enPassantTargetSquare = move.from - 8
				}
			}

			if move.IsPromotion() {
				promotionPiece := move.GetPromotionPiece()
				boardState.board[move.to] = promotionPiece | sideToMoveToColorMask(boardState.sideToMove)

				boardState.bitboards.piece[BITBOARD_PAWN_OFFSET] = FlipBitboard(
					boardState.bitboards.piece[BITBOARD_PAWN_OFFSET],
					move.to)

				offset := promotionPiece & 0x0F
				boardState.bitboards.piece[offset] = SetBitboard(
					boardState.bitboards.piece[offset],
					move.to)
			}
		}
	}

	switch boardState.sideToMove {
	case WHITE_OFFSET:
		boardState.sideToMove = BLACK_OFFSET
	case BLACK_OFFSET:
		boardState.sideToMove = WHITE_OFFSET
	}

	oldBoardInfo := boardState.boardInfoHistory[boardState.moveIndex]
	boardState.moveIndex++

	if boardState.sideToMove == WHITE_OFFSET {
		boardState.fullmoveNumber++
	}

	boardState.UpdateHashApplyMove(oldBoardInfo, move, isCapture)
	boardState.repetitionInfo.occurredHashes[boardState.moveIndex] = boardState.hashKey
	boardState.repetitionInfo.pawnMoveOrCapture[boardState.moveIndex] = movePiece == PAWN_MASK || capturedPiece != 0
}

func (boardState *BoardState) ApplyNullMove() {
	boardState.boardInfoHistory[boardState.moveIndex] = boardState.boardInfo
	boardState.moveIndex++
	boardState.sideToMove = oppositeColorOffset(boardState.sideToMove)
	boardState.hashKey ^= boardState.hashInfo.sideToMove
}

func (boardState *BoardState) UnapplyNullMove() {
	boardState.moveIndex--
	boardState.boardInfo = boardState.boardInfoHistory[boardState.moveIndex]
	boardState.sideToMove = oppositeColorOffset(boardState.sideToMove)
	boardState.hashKey ^= boardState.hashInfo.sideToMove
}

func (boardState *BoardState) IsMoveLegal(move Move) (bool, error) {
	if move.from == move.to {
		return false, errMoveUninitialized
	}

	fromPiece := boardState.board[move.from]
	toPiece := boardState.board[move.to]
	var pieceMask byte
	var captureMask byte

	if boardState.sideToMove == WHITE_OFFSET {
		pieceMask = WHITE_MASK
		captureMask = BLACK_MASK
	} else {
		pieceMask = BLACK_MASK
		captureMask = WHITE_MASK
	}

	if fromPiece == EMPTY_SQUARE {
		return false, errors.New("From square was empty")
	} else if fromPiece&pieceMask != pieceMask {
		return false, errors.New("From square was not occupied by expected piece: " + strconv.Itoa(int(fromPiece)))
	}

	if toPiece != EMPTY_SQUARE {
		if toPiece&captureMask != captureMask {
			return false, errors.New("Attempted to capture a piece of the same color: " + strconv.Itoa(int(toPiece)))
		}

		if toPiece == KING_MASK|captureMask {
			return false, errors.New("Attempted to capture king")
		}
	}

	if fromPiece&0x0F == PAWN_MASK {
		if pieceMask == BLACK_MASK {
			if move.to <= SQUARE_H1 && move.flags&PROMOTION_MASK == 0 {
				return false, errors.New("Promoting without a promotion mask")
			}
		} else {
			if move.to >= SQUARE_A8 && move.flags&PROMOTION_MASK == 0 {
				return false, errors.New("Promoting without a promotion mask")
			}
		}
	}

	// TODO: ensure side to move not in check
	// TODO: ensure castling isn't moving through check

	return true, nil
}

func (boardState *BoardState) UnapplyMove(move Move) {
	boardState.sideToMove = oppositeColorOffset(boardState.sideToMove)
	if boardState.sideToMove == BLACK_OFFSET {
		boardState.fullmoveNumber--
	}
	oldBoardInfo := boardState.boardInfo
	boardState.moveIndex--
	boardState.boardInfo = boardState.boardInfoHistory[boardState.moveIndex]

	var p = boardState.board[move.to]
	var movePiece = p & 0x0F

	var capturedPiece byte
	isCapture := boardState.wasCapture[boardState.moveIndex+1]
	if isCapture {
		capturedPiece = boardState.captureStack.Pop()
		boardState.board[move.to] = capturedPiece
	} else {
		boardState.board[move.to] = 0x00
	}

	boardState.board[move.from] = p

	var offset = boardState.sideToMove
	var otherOffset int
	switch boardState.sideToMove {
	case WHITE_OFFSET:
		otherOffset = BLACK_OFFSET
	case BLACK_OFFSET:
		otherOffset = WHITE_OFFSET
	}

	boardState.bitboards.piece[movePiece] = FlipBitboard2(boardState.bitboards.piece[movePiece], move.from, move.to)
	boardState.bitboards.color[offset] = FlipBitboard2(boardState.bitboards.color[offset], move.from, move.to)

	if isCapture {
		boardState.bitboards.color[otherOffset] = SetBitboard(boardState.bitboards.color[otherOffset], move.to)
		boardState.bitboards.piece[capturedPiece&0x0F] = SetBitboard(boardState.bitboards.piece[capturedPiece&0x0F], move.to)
	}

	// TODO(perf) - just switch statement on the different conditions here, they are all mutually exclusive
	if move.IsQueensideCastle() {
		// black was to move, so we're unmaking a white move
		if boardState.sideToMove == WHITE_OFFSET {
			boardState.board[SQUARE_D1] = 0x00
			boardState.board[SQUARE_A1] = WHITE_MASK | ROOK_MASK

			boardState.bitboards.piece[BITBOARD_ROOK_OFFSET] = FlipBitboard2(
				boardState.bitboards.piece[BITBOARD_ROOK_OFFSET],
				SQUARE_D1,
				SQUARE_A1)
			boardState.bitboards.color[WHITE_OFFSET] = FlipBitboard2(
				boardState.bitboards.color[WHITE_OFFSET],
				SQUARE_D1,
				SQUARE_A1)

			boardState.boardInfo.whiteCanCastleQueenside = true
			boardState.boardInfo.whiteHasCastled = false
		} else {
			boardState.board[SQUARE_D8] = 0x00
			boardState.board[SQUARE_A8] = BLACK_MASK | ROOK_MASK

			boardState.bitboards.piece[BITBOARD_ROOK_OFFSET] = FlipBitboard2(
				boardState.bitboards.piece[BITBOARD_ROOK_OFFSET],
				SQUARE_D8,
				SQUARE_A8)
			boardState.bitboards.color[BLACK_OFFSET] = FlipBitboard2(
				boardState.bitboards.color[BLACK_OFFSET],
				SQUARE_D8,
				SQUARE_A8)

			boardState.boardInfo.blackCanCastleQueenside = true
			boardState.boardInfo.blackHasCastled = false
		}
	} else if move.IsKingsideCastle() {
		if boardState.sideToMove == WHITE_OFFSET {
			boardState.board[SQUARE_F1] = 0x00
			boardState.board[SQUARE_H1] = WHITE_MASK | ROOK_MASK

			boardState.bitboards.piece[BITBOARD_ROOK_OFFSET] = FlipBitboard2(boardState.bitboards.piece[BITBOARD_ROOK_OFFSET],
				SQUARE_F1,
				SQUARE_H1)
			boardState.bitboards.color[WHITE_OFFSET] = FlipBitboard2(boardState.bitboards.color[WHITE_OFFSET],
				SQUARE_F1,
				SQUARE_H1)

			boardState.boardInfo.whiteCanCastleKingside = true
			boardState.boardInfo.whiteHasCastled = false
		} else {
			boardState.board[SQUARE_F8] = 0x00
			boardState.board[SQUARE_H8] = BLACK_MASK | ROOK_MASK

			boardState.bitboards.piece[BITBOARD_ROOK_OFFSET] = FlipBitboard2(
				boardState.bitboards.piece[BITBOARD_ROOK_OFFSET],
				SQUARE_F8,
				SQUARE_H8)
			boardState.bitboards.color[BLACK_OFFSET] = FlipBitboard2(
				boardState.bitboards.color[BLACK_OFFSET],
				SQUARE_F8,
				SQUARE_H8)

			boardState.boardInfo.blackCanCastleKingside = true
			boardState.boardInfo.blackHasCastled = false
		}
	}

	switch movePiece {
	case PAWN_MASK:
		if move.IsEnPassantCapture() {
			var pos uint8
			var otherOffset int
			var otherMask byte
			var offset int
			if boardState.sideToMove == WHITE_OFFSET {
				pos = move.to - 8
				otherOffset = BLACK_OFFSET
				otherMask = BLACK_MASK
				offset = WHITE_OFFSET
			} else {
				pos = move.to + 8
				offset = BLACK_OFFSET
				otherMask = WHITE_MASK
				otherOffset = WHITE_OFFSET
			}

			// we don't use the captureStack for EP capture
			isCapture = true
			boardState.board[pos] = otherMask | PAWN_MASK
			boardState.board[move.to] = 0x00
			boardState.boardInfo.enPassantTargetSquare = move.to

			boardState.bitboards.color[offset] = SetBitboard(
				UnsetBitboard(boardState.bitboards.color[offset], move.to),
				move.from)
			boardState.bitboards.color[otherOffset] = SetBitboard(
				boardState.bitboards.color[otherOffset],
				pos)
			boardState.bitboards.piece[BITBOARD_PAWN_OFFSET] = SetBitboard(
				UnsetBitboard(boardState.bitboards.piece[BITBOARD_PAWN_OFFSET], move.to),
				pos)
		}
	}

	if move.IsPromotion() {
		var mask byte
		if boardState.sideToMove == WHITE_OFFSET {
			mask = WHITE_MASK
		} else {
			mask = BLACK_MASK
		}
		boardState.board[move.from] = mask | PAWN_MASK

		boardState.bitboards.piece[BITBOARD_PAWN_OFFSET] = SetBitboard(
			boardState.bitboards.piece[BITBOARD_PAWN_OFFSET],
			move.from)

		offset := move.GetPromotionPiece() & 0x0F
		boardState.bitboards.piece[offset] = FlipBitboard(
			boardState.bitboards.piece[offset],
			move.from)
	}

	boardState.UpdateHashUnapplyMove(oldBoardInfo, move, isCapture)
}
