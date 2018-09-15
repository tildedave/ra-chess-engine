package main

import (
	"errors"
	"strconv"
	"strings"
)

var m map[byte]byte = make(map[byte]byte)

func isPieceBlack(p byte) bool {
	return p&BLACK_MASK == BLACK_MASK
}

func isPieceWhite(p byte) bool {
	return p&WHITE_MASK == WHITE_MASK
}

func isSentinel(p byte) bool {
	return p == SENTINEL_MASK
}

func isPawn(p byte) bool {
	return p&0x0F == PAWN_MASK
}

func isBishop(p byte) bool {
	return p&0x0F == BISHOP_MASK
}

func isKnight(p byte) bool {
	return p&0x0F == KNIGHT_MASK
}

func isRook(p byte) bool {
	return p&0x0F == ROOK_MASK
}

func isQueen(p byte) bool {
	return p&0x0F == QUEEN_MASK
}

func isKing(p byte) bool {
	return p&0x0F == KING_MASK
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

	for i := byte(0); i < 8; i++ {
		for j := byte(0); j < 8; j++ {
			var p = boardState.PieceAtSquare(RowAndColToSquare(i, j))
			s[(7-i)*9+j] = pieceToString(p)
		}
		s[(7-i)*9+8] = '\n'
	}
	return string(s[:9*8]) + "\n" + boardState.ToFENString()
}

func (boardState *BoardState) ToFENString() string {
	var s string

	for i := byte(0); i < 8; i++ {
		var numEmpty = 0
		for j := byte(0); j < 8; j++ {
			p := boardState.PieceAtSquare(RowAndColToSquare(7-i, j))
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
	s += " " + strconv.FormatUint(uint64(boardState.halfmoveClock), 10)
	s += " " + strconv.FormatUint(uint64(boardState.fullmoveNumber), 10)

	return s
}

func (boardState *BoardState) PieceAtSquare(sq uint8) byte {
	return boardState.board[sq]
}

func RowAndColToSquare(row uint8, col uint8) uint8 {
	return 20 + row*10 + 1 + col
}

func Rank(sq uint8) uint8 {
	return (sq-1)/10 - 1
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

type BoardLookupInfo struct {
	whiteKingSquare byte
	blackKingSquare byte
}

type BoardState struct {
	board       []byte
	whiteToMove bool
	boardInfo   BoardInfo

	// number of moves since last capture or pawn advance
	halfmoveClock uint
	// starts at 1, incremented after Black moves
	fullmoveNumber uint

	// Internal structures to allow unmaking moves
	captureStack     byteStack
	boardInfoHistory [MAX_MOVES]BoardInfo
	moveIndex        int // 0-based and increases after every move
	lookupInfo       BoardLookupInfo
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
	generateBoardLookupInfo(&b)
	// generateZobristHashInfo(&b)

	return b
}

func generateBoardLookupInfo(boardState *BoardState) {
	for i := byte(0); i < 8; i++ {
		for j := byte(0); j < 8; j++ {
			sq := RowAndColToSquare(i, j)
			p := boardState.board[sq]
			if p == BLACK_MASK|KING_MASK {
				boardState.lookupInfo.blackKingSquare = sq
			} else if p == WHITE_MASK|KING_MASK {
				boardState.lookupInfo.whiteKingSquare = sq
			}
		}
	}
}

func generateZobrishHashInfo(boardState *BoardState) {

}

func CreateInitialBoardState() BoardState {
	var b BoardState

	// https://chessprogramming.wikispaces.com/10x12+Board
	b.board = []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0x44, 0x42, 0x43, 0x45, 0x46, 0x43, 0x42, 0x44, 0xFF,
		0xFF, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0xFF,
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
	generateBoardLookupInfo(&b)

	return b
}

func CreateBoardStateFromFENString(s string) (BoardState, error) {
	var boardState BoardState = CreateEmptyBoardState()

	// first split string into board part and non-board part
	var splits = strings.SplitN(s, " ", 6)
	boardStr := splits[0]
	row := byte(7)

	for _, rowStr := range strings.Split(boardStr, "/") {
		col := byte(0)
		for _, pStr := range strings.Split(rowStr, "") {
			sq := RowAndColToSquare(row, col)
			var p byte
			switch pStr {
			case "P":
				p = WHITE_MASK | PAWN_MASK
				col++
			case "N":
				p = WHITE_MASK | KNIGHT_MASK
				col++
			case "B":
				p = WHITE_MASK | BISHOP_MASK
				col++
			case "R":
				p = WHITE_MASK | ROOK_MASK
				col++
			case "K":
				p = WHITE_MASK | KING_MASK
				col++
			case "Q":
				p = WHITE_MASK | QUEEN_MASK
				col++
			case "p":
				p = BLACK_MASK | PAWN_MASK
				col++
			case "n":
				p = BLACK_MASK | KNIGHT_MASK
				col++
			case "b":
				p = BLACK_MASK | BISHOP_MASK
				col++
			case "r":
				p = BLACK_MASK | ROOK_MASK
				col++
			case "k":
				p = BLACK_MASK | KING_MASK
				col++
			case "q":
				p = BLACK_MASK | QUEEN_MASK
				col++
			default:
				num, err := strconv.ParseUint(pStr, 10, 8)
				if err != nil {
					return boardState, errors.New("Found unknown character parsing FEN: " + pStr)
				}
				if num > 8 {
					return boardState, errors.New("Invalid FEN offset: " + pStr)
				}
				col = col + byte(num)
			}

			boardState.board[sq] = p
		}
		row = row - 1
	}

	switch splits[1] {
	case "w":
		boardState.whiteToMove = true
	case "b":
		boardState.whiteToMove = false
	default:
		return boardState, errors.New("Invalid side-to-move specification: " + splits[1])
	}

	if splits[2] != "-" {
		for _, castleStr := range strings.Split(splits[2], "") {
			switch castleStr {
			case "K":
				boardState.boardInfo.whiteCanCastleKingside = true
			case "Q":
				boardState.boardInfo.whiteCanCastleQueenside = true
			case "k":
				boardState.boardInfo.blackCanCastleKingside = true
			case "q":
				boardState.boardInfo.blackCanCastleQueenside = true
			}
		}
	}

	// en passant target square parsing
	if splits[3] != "-" {
		sq, err := ParseAlgebraicSquare(splits[3])

		if err != nil {
			return boardState, err
		}

		boardState.boardInfo.enPassantTargetSquare = sq
	}

	if len(splits) > 4 {
		halfmoveClock, err := strconv.ParseUint(splits[4], 10, 8)
		if err != nil {
			return boardState, errors.New("Error parsing halfmove clock count: " + splits[4])
		}

		fullmoveNumber, err := strconv.ParseUint(splits[5], 10, 8)
		if err != nil {
			return boardState, errors.New("Error parsing fullmove number count: " + splits[4])
		}

		boardState.halfmoveClock = uint(halfmoveClock)
		boardState.fullmoveNumber = uint(fullmoveNumber)
	}

	generateBoardLookupInfo(&boardState)

	return boardState, nil
}

func ParseAlgebraicSquare(sq string) (uint8, error) {
	var col byte
	var row byte
	for index, runeValue := range sq {
		if index == 0 {
			switch runeValue {
			case 'a':
				col = byte(0)
			case 'b':
				col = byte(1)
			case 'c':
				col = byte(2)
			case 'd':
				col = byte(3)
			case 'e':
				col = byte(4)
			case 'f':
				col = byte(5)
			case 'g':
				col = byte(6)
			case 'h':
				col = byte(7)
			default:
				return 0, errors.New("Column out of range: " + sq)
			}
		} else if index == 1 {
			rowUint, err := strconv.ParseUint(string(runeValue), 10, 8)
			if err != nil {
				return 0, errors.New("Row out of range: " + sq)
			}
			// Must subtract 1 because algebraic notation is 1-based
			row = byte(rowUint - 1)
		} else {
			return 0, errors.New("Algebraic square was not two characters: " + sq)
		}
	}

	return RowAndColToSquare(row, col), nil

}

func (boardState *BoardState) SetPieceAtSquare(sq byte, p byte) {
	boardState.board[sq] = p
}
