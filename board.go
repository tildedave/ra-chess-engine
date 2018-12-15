package main

import (
	"errors"
	"fmt"
	"math/bits"
	"math/rand"
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

func whiteToMoveToColorMask(whiteToMove bool) uint8 {
	if whiteToMove {
		return WHITE_MASK
	}

	return BLACK_MASK
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

func ColumnToAlgebraicNotation(col uint8) string {
	switch col {
	case 1:
		return "a"
	case 2:
		return "b"
	case 3:
		return "c"
	case 4:
		return "d"
	case 5:
		return "e"
	case 6:
		return "f"
	case 7:
		return "g"
	case 8:
		return "h"
	default:
		return ""
	}
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
	return ColumnToAlgebraicNotation(col) + rowStr
}

type BoardInfo struct {
	whiteCanCastleKingside  bool
	whiteCanCastleQueenside bool
	blackCanCastleKingside  bool
	blackCanCastleQueenside bool
	blackHasCastled         bool
	whiteHasCastled         bool
	enPassantTargetSquare   uint8
}

type BoardLookupInfo struct {
	whiteKingSquare byte
	blackKingSquare byte
}

type BoardState struct {
	board         []byte
	bitboards     Bitboards
	moveBitboards *MoveBitboards
	whiteToMove   bool
	boardInfo     BoardInfo

	// number of moves since last capture or pawn advance
	halfmoveClock uint
	// starts at 1, incremented after Black moves
	fullmoveNumber uint

	hashKey uint32

	// Internal structures to allow unmaking moves
	captureStack     byteStack
	boardInfoHistory [MAX_MOVES]BoardInfo
	moveIndex        int // 0-based and increases after every move
	lookupInfo       BoardLookupInfo

	// Zobrist hash indices
	hashInfo *HashInfo
	// Transposition table
	transpositionTable map[uint32]*TranspositionEntry
}

func CopyBoardState(boardState *BoardState) BoardState {
	state := *boardState
	state.board = make([]byte, 120)
	copy(state.board, boardState.board)
	generateTranspositionTable(&state)
	return state
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
	if moveBitboards == nil {
		moveBoards := CreateMoveBitboards()
		moveBitboards = &moveBoards
	}
	b.moveBitboards = moveBitboards
	generateBoardLookupInfo(&b)
	generateZobrishHashInfo(&b)
	generateTranspositionTable(&b)

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
	r := rand.New(rand.NewSource(0))
	hashInfo := CreateHashInfo(r)
	boardState.hashInfo = &hashInfo
	// this factoring is dumb since we need to keep the hash
	// info around to progressively change the hash key
	boardState.hashKey = boardState.CreateHashKey(&hashInfo)
}

func CreateInitialBoardState() BoardState {
	boardState, err := CreateBoardStateFromFENString("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		panic(err)
	}

	return boardState
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

			if p != 0 {
				boardState.SetPieceAtSquare(sq, p)
			}
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

func PieceToColorOffset(p byte) int {
	switch p & 0xF0 {
	case WHITE_MASK:
		return WHITE_OFFSET
	case BLACK_MASK:
		return BLACK_OFFSET
	default:
		panic(fmt.Sprintf("Invalid piece: %x", p))
	}
}

// SetPieceAtSquare should only be used in non-performance critical places.
func (boardState *BoardState) SetPieceAtSquare(sq byte, p byte) {
	boardState.board[sq] = p

	bbSq := legacySquareToBitboardSquare(sq)
	if p != EMPTY_SQUARE {
		colorOffset := PieceToColorOffset(p)
		pieceOffset := p & 0x0F
		boardState.bitboards.color[colorOffset] = SetBitboard(boardState.bitboards.color[colorOffset], bbSq)
		boardState.bitboards.piece[pieceOffset] = SetBitboard(boardState.bitboards.piece[pieceOffset], bbSq)
	} else {
		for _, colorOffset := range []int{WHITE_OFFSET, BLACK_OFFSET} {
			boardState.bitboards.color[colorOffset] = UnsetBitboard(boardState.bitboards.color[colorOffset], bbSq)
		}
		for _, pieceOffset := range BITBOARD_PIECES {
			boardState.bitboards.piece[pieceOffset] = UnsetBitboard(boardState.bitboards.piece[pieceOffset], bbSq)
		}
	}
}

// CreateMovesFromBitboard transforms a bitboard and a square to a slice of moves.
func CreateMovesFromBitboard(sq byte, moveBoard uint64) []Move {
	moves := make([]Move, 0)

	for moveBoard != 0 {
		destSq := byte(bits.TrailingZeros64(moveBoard))
		moveBoard ^= 1 << destSq
		moves = append(moves, CreateMove(sq, destSq))
	}

	return moves
}

func sanityCheckBitboards(move Move, boardState *BoardState) {
	for row := byte(0); row < 8; row++ {
		for col := byte(0); col < 8; col++ {
			legacySq := RowAndColToSquare(row, col)
			sq := idx(col, row)
			piece := boardState.PieceAtSquare(legacySq)
			var isError = false
			var message string
			var bitboard uint64

			if piece == EMPTY_SQUARE {
				for _, colorOffset := range []int{WHITE_OFFSET, BLACK_OFFSET} {
					if IsBitboardSet(boardState.bitboards.color[colorOffset], sq) {
						isError = true
						message = "Empty square had occupancy set"
						bitboard = boardState.bitboards.color[colorOffset]
					}
				}
				for _, pieceOffset := range BITBOARD_PIECES {
					if IsBitboardSet(boardState.bitboards.piece[pieceOffset], sq) {
						isError = true
						message = "Empty square was set for piece offset"
						bitboard = boardState.bitboards.color[pieceOffset]
					}
				}
			} else {
				colorOffset := PieceToColorOffset(piece)
				var otherColorOffset = 1
				if colorOffset == 1 {
					otherColorOffset = 0
				}
				if !IsBitboardSet(boardState.bitboards.color[colorOffset], sq) {
					isError = true
					message = "Color occupancy bitboard was not set for piece on square"
					bitboard = boardState.bitboards.color[colorOffset]
				}
				if IsBitboardSet(boardState.bitboards.color[otherColorOffset], sq) {
					isError = true
					message = "Color occupancy bitboard was set for opposite color of piece on square"
					bitboard = boardState.bitboards.color[otherColorOffset]
				}
				for _, pieceOffset := range BITBOARD_PIECES {
					if pieceOffset == piece&0x0F {
						if !IsBitboardSet(boardState.bitboards.piece[pieceOffset], sq) {
							isError = true
							message = fmt.Sprintf("Piece occupancy bitboard was not set for piece on square (piece=%d)", pieceOffset)
							bitboard = boardState.bitboards.piece[pieceOffset]
						}
					} else if IsBitboardSet(boardState.bitboards.piece[pieceOffset], sq) {
						isError = true
						message = "Piece occupancy bitboard was set for wrong kind of piece"
						bitboard = boardState.bitboards.piece[pieceOffset]
					}
				}
			}
			if isError {
				fmt.Println(boardState.ToString())
				fmt.Println(BitboardToString(bitboard))
				fmt.Println(MoveToString(move))
				panic(message)
			}
		}
	}
}
