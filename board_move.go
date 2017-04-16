package main

func (boardState *BoardState) ApplyMove(move Move) {
	boardState.boardInfoHistory[boardState.moveIndex] = boardState.boardInfo

	if move.IsCapture() {
		boardState.captureStack.Push(boardState.board[move.to])
	}

	var p = boardState.board[move.from]
	boardState.board[move.from] = 0x00
	boardState.board[move.to] = p

	// TODO(perf) - less if statements/work when castling is over
	if move.IsQueensideCastle() {
		// white
		if boardState.whiteToMove {
			boardState.board[21] = 0x00
			boardState.board[24] = WHITE_MASK | ROOK_MASK
			boardState.boardInfo.whiteCanCastleKingside = false
			boardState.boardInfo.whiteCanCastleQueenside = false
		} else {
			boardState.board[91] = 0x00
			boardState.board[94] = BLACK_MASK | ROOK_MASK
			boardState.boardInfo.blackCanCastleKingside = false
			boardState.boardInfo.blackCanCastleQueenside = false
		}
	} else if move.IsKingsideCastle() {
		if boardState.whiteToMove {
			boardState.board[28] = 0x00
			boardState.board[26] = WHITE_MASK | ROOK_MASK
			boardState.boardInfo.whiteCanCastleKingside = false
			boardState.boardInfo.whiteCanCastleQueenside = false
		} else {
			boardState.board[98] = 0x00
			boardState.board[96] = BLACK_MASK | ROOK_MASK
			boardState.boardInfo.blackCanCastleKingside = false
			boardState.boardInfo.blackCanCastleQueenside = false
		}
	} else if p&KING_MASK == KING_MASK {
		if boardState.whiteToMove {
			boardState.boardInfo.whiteCanCastleKingside = false
			boardState.boardInfo.whiteCanCastleQueenside = false
			boardState.lookupInfo.whiteKingSquare = move.to
		} else {
			boardState.boardInfo.blackCanCastleKingside = false
			boardState.boardInfo.blackCanCastleQueenside = false
			boardState.lookupInfo.blackKingSquare = move.from
		}
	} else if p&ROOK_MASK == ROOK_MASK {
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
	} else if p&PAWN_MASK == PAWN_MASK && !move.IsCapture() {
		if move.to > move.from {
			if move.to-move.from > 10 {
				boardState.boardInfo.enPassantTargetSquare = move.from + 10
			}
		} else if move.from-move.to > 10 {
			boardState.boardInfo.enPassantTargetSquare = move.from - 10
		}
	} else {
		boardState.boardInfo.enPassantTargetSquare = 0
	}

	boardState.whiteToMove = !boardState.whiteToMove
	boardState.moveIndex++

	if boardState.whiteToMove {
		boardState.fullmoveNumber += 1
	}
}

func (boardState *BoardState) UnapplyMove(move Move) {
	boardState.whiteToMove = !boardState.whiteToMove
	if !boardState.whiteToMove {
		boardState.fullmoveNumber -= 1
	}
	boardState.moveIndex--
	boardState.boardInfo = boardState.boardInfoHistory[boardState.moveIndex]

	var p = boardState.board[move.to]

	if move.IsCapture() {
		var capturedPiece = boardState.captureStack.Pop()
		boardState.board[move.to] = capturedPiece
	} else {
		boardState.board[move.to] = 0x00
	}

	boardState.board[move.from] = p

	// TODO(perf) - less if statements for the average case
	// this could be not a big deal b/c of branch prediction
	if move.IsQueensideCastle() {
		// black was to move, so we're unmaking a white move
		if boardState.whiteToMove {
			boardState.board[24] = 0x00
			boardState.board[21] = WHITE_MASK | ROOK_MASK
			boardState.boardInfo.whiteCanCastleQueenside = true
		} else {
			boardState.board[94] = 0x00
			boardState.board[91] = BLACK_MASK | ROOK_MASK
			boardState.boardInfo.blackCanCastleQueenside = true
		}
	} else if move.IsKingsideCastle() {
		if boardState.whiteToMove {
			boardState.board[26] = 0x00
			boardState.board[28] = WHITE_MASK | ROOK_MASK
			boardState.boardInfo.whiteCanCastleKingside = true
		} else {
			boardState.board[96] = 0x00
			boardState.board[98] = BLACK_MASK | ROOK_MASK
			boardState.boardInfo.blackCanCastleKingside = true
		}
	} else if p&KING_MASK == KING_MASK {
		if boardState.whiteToMove {
			boardState.lookupInfo.whiteKingSquare = move.from
		} else {
			boardState.lookupInfo.blackKingSquare = move.from
		}
	}
}
