package main

import (
	"fmt"
	"strconv"
)

var m map[byte]byte = make(map[byte]byte)

func isPieceBlack(p byte) bool {
	return p&BLACK_MASK == BLACK_MASK
}

func isPieceWhite(p byte) bool {
	// WHITE_MASK isn't a mask, gotta see if unsetting the high bit is the same
	return p&WHITE_MASK == WHITE_MASK
}

func isSentinel(p byte) bool {
	return p&SENTINEL_MASK == SENTINEL_MASK
}

func isPawn(p byte) bool {
	return p&PAWN_MASK == PAWN_MASK
}

func isBishop(p byte) bool {
	return p&BISHOP_MASK == BISHOP_MASK
}

func isKnight(p byte) bool {
	return p&KNIGHT_MASK == KNIGHT_MASK
}

func isRook(p byte) bool {
	return p&ROOK_MASK == ROOK_MASK
}

func isQueen(p byte) bool {
	return p&QUEEN_MASK == QUEEN_MASK
}

func isKing(p byte) bool {
	return p&KING_MASK == KING_MASK
}

func isSquareEmpty(p byte) bool {
	return p == 0x00
}

func pieceToString(p byte) byte {
	if p == 0x00 {
		return '.'
	} else if p == PAWN_MASK|WHITE_MASK {
		return 'P'
	} else if p == KNIGHT_MASK|WHITE_MASK {
		return 'N'
	} else if p == BISHOP_MASK|WHITE_MASK {
		return 'B'
	} else if p == ROOK_MASK|WHITE_MASK {
		return 'R'
	} else if p == QUEEN_MASK|WHITE_MASK {
		return 'Q'
	} else if p == KING_MASK|WHITE_MASK {
		return 'K'
	} else if p == PAWN_MASK|BLACK_MASK {
		return 'p'
	} else if p == KNIGHT_MASK|BLACK_MASK {
		return 'n'
	} else if p == BISHOP_MASK|BLACK_MASK {
		return 'b'
	} else if p == ROOK_MASK|BLACK_MASK {
		return 'r'
	} else if p == QUEEN_MASK|BLACK_MASK {
		return 'q'
	} else if p == KING_MASK|BLACK_MASK {
		return 'k'
	}

	return '-'
}

func (boardState *BoardState) ToString() string {
	var s [9 * 8]byte

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			var p = boardState.PieceAtSquare(RowAndColToSquare(byte(i), byte(j)))
			s[(7-i)*9+j] = pieceToString(p)
		}
		s[(7-i)*9+8] = '\n'
	}
	return string(s[:9*8])
}

func (boardState *BoardState) ToFENString() string {
	var s string

	for i := 0; i < 8; i++ {
		var numEmpty = 0
		for j := 0; j < 8; j++ {
			p := boardState.PieceAtSquare(RowAndColToSquare(byte(7-i), byte(j)))
			if isSquareEmpty(p) {
				numEmpty++
			} else {
				if numEmpty > 0 {
					s += strconv.Itoa(numEmpty)
					numEmpty = 0
				}
				s += string(pieceToString(p))
				numEmpty = 0
			}
		}
		if numEmpty > 0 {
			s += strconv.Itoa(numEmpty)
			numEmpty = 0
		}
		if i < 7 {
			s += "/"
		} else {
			s += " "
		}
	}

	if boardState.whiteToMove {
		s += "w "
	} else {
		s += "b "
	}

	var hasCastleSquare = false
	if boardState.boardInfo.whiteCanCastleKingside {
		s += "K"
		hasCastleSquare = true
	}
	if boardState.boardInfo.whiteCanCastleQueenside {
		s += "Q"
		hasCastleSquare = true
	}
	if boardState.boardInfo.blackCanCastleKingside {
		s += "k"
		hasCastleSquare = true
	}
	if boardState.boardInfo.blackCanCastleQueenside {
		s += "q"
		hasCastleSquare = true
	}
	if !hasCastleSquare {
		s += "-"
	}
	s += " "
	if boardState.boardInfo.enPassantTargetSquare == 0 {
		s += "-"
	} else {
		s += SquareToAlgebraicString(boardState.boardInfo.enPassantTargetSquare)
	}
	s += " " + strconv.Itoa(boardState.halfmoveClock) + " " + strconv.Itoa(boardState.fullmoveNumber)

	return s
}

func (boardState *BoardState) PieceAtSquare(sq uint8) byte {
	// row is 0 - 7, col is 0 - 8
	// 10x12 board
	return boardState.board[sq]
}

func RowAndColToSquare(row uint8, col uint8) uint8 {
	return 20 + row*10 + 1 + col
}

func SquareToAlgebraicString(sq uint8) string {
	var row = sq / 10
	var col = sq % 10

	if row < 2 || row > 9 {
		return "??"
	}
	if col == 0 || col == 9 {
		return "??"
	}

	// No need to offset this as board is 1-indexed
	var rowStr = strconv.Itoa(int(row - 1))
	switch col {
	case 1:
		return "a" + rowStr
	case 2:
		return "b" + rowStr
	case 3:
		return "c" + rowStr
	case 4:
		return "d" + rowStr
	case 5:
		return "e" + rowStr
	case 6:
		return "f" + rowStr
	case 7:
		return "g" + rowStr
	case 8:
		return "h" + rowStr
	default:
	}
	return "-"

}

type BoardInfo struct {
	whiteCanCastleKingside  bool
	whiteCanCastleQueenside bool
	blackCanCastleKingside  bool
	blackCanCastleQueenside bool
	enPassantTargetSquare   uint8
}

type BoardState struct {
	board       []byte
	whiteToMove bool
	boardInfo   BoardInfo

	// number of moves since last capture or pawn advance
	halfmoveClock int
	// starts at 1, incremented after Black moves
	fullmoveNumber int

	// Internal structures to allow unmaking moves
	captureStack     byteStack
	boardInfoHistory [MAX_MOVES]BoardInfo
	moveIndex        int // 0-based and increases after every move
}

func CreateEmptyBoardState() BoardState {
	var b BoardState
	b.board = []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF,
		0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF,
		0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF,
		0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF,
		0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF,
		0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF,
		0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF,
		0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	}

	b.whiteToMove = true
	b.boardInfo = BoardInfo{
		whiteCanCastleKingside:  false,
		whiteCanCastleQueenside: false,
		blackCanCastleKingside:  false,
		blackCanCastleQueenside: false,
		enPassantTargetSquare:   0,
	}
	b.halfmoveClock = 0
	b.fullmoveNumber = 1

	return b
}

func CreateInitialBoardState() BoardState {
	var b BoardState

	// https://chessprogramming.wikispaces.com/10x12+Board
	b.board = []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0x04, 0x02, 0x03, 0x05, 0x06, 0x03, 0x02, 0x04, 0xFF,
		0xFF, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0xFF,
		0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF,
		0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF,
		0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF,
		0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF,
		0xFF, 0x81, 0x81, 0x81, 0x81, 0x81, 0x81, 0x81, 0x81, 0xFF,
		0xFF, 0x84, 0x82, 0x83, 0x85, 0x86, 0x83, 0x82, 0x84, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	}

	b.whiteToMove = true
	b.boardInfo = BoardInfo{
		whiteCanCastleKingside:  true,
		whiteCanCastleQueenside: true,
		blackCanCastleKingside:  true,
		blackCanCastleQueenside: true,
		enPassantTargetSquare:   0,
	}
	b.halfmoveClock = 0
	b.fullmoveNumber = 1

	return b
}

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
		} else {
			boardState.boardInfo.blackCanCastleKingside = false
			boardState.boardInfo.blackCanCastleQueenside = false
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
	}
}

func main() {
	fmt.Println("Hello, world")
}
