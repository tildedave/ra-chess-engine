package main

import (
	"fmt"
)

var _ = fmt.Println

const ENDGAME_MATERIAL_THRESHOLD = 1500

const QUEEN_EVAL_SCORE = 800
const PAWN_EVAL_SCORE = 100
const ROOK_EVAL_SCORE = 500
const KNIGHT_EVAL_SCORE = 300
const BISHOP_EVAL_SCORE = 300
const KING_IN_CENTER_EVAL_SCORE = 50
const KING_PAWN_COVER_EVAL_SCORE = 20
const KING_CANNOT_CASTLE_EVAL_SCORE = 70
const KING_CASTLED_EVAL_SCORE = 30
const PAWN_IN_CENTER_EVAL_SCORE = 40
const PIECE_IN_CENTER_EVAL_SCORE = 25
const PIECE_ATTACKS_CENTER_EVAL_SCORE = 15

const (
	PHASE_OPENING    = iota
	PHASE_MIDDLEGAME = iota
	PHASE_ENDGAME    = iota
)

type BoardEval struct {
	material      int
	phase         int
	centerControl int
	development   int
	kingPosition  int
	passedPawns   int
}

var materialScore = [7]int{
	0, PAWN_EVAL_SCORE, KNIGHT_EVAL_SCORE, BISHOP_EVAL_SCORE, ROOK_EVAL_SCORE, QUEEN_EVAL_SCORE, 0,
}

func Eval(boardState *BoardState) BoardEval {
	material := 0
	boardPhase := PHASE_OPENING
	if boardState.fullmoveNumber > 8 {
		boardPhase = PHASE_MIDDLEGAME
	}

	blackMaterial := 0
	whiteMaterial := 0

	for i := byte(0); i < 8; i++ {
		for j := byte(0); j < 8; j++ {
			p := boardState.PieceAtSquare(RowAndColToSquare(i, j))
			if p != 0x00 && p != SENTINEL_MASK {
				pieceMask := p & 0x0F
				score := materialScore[pieceMask]
				isBlack := p&0xF0 == BLACK_MASK

				if pieceMask != PAWN_MASK {
					if isBlack {
						blackMaterial += score
					} else {
						whiteMaterial += score
					}
				}
				if isBlack {
					material -= score
				} else {
					material += score
				}
			}
		}
	}

	if blackMaterial < ENDGAME_MATERIAL_THRESHOLD && whiteMaterial < ENDGAME_MATERIAL_THRESHOLD {
		boardPhase = PHASE_ENDGAME
	}

	// TODO - endgame: determine passed pawns and prioritize them

	kingPosition := 0
	centerControl := 0

	if boardPhase != PHASE_ENDGAME {
		// prioritize center control

		// D4, E4, D5, E5

		for _, sq := range [4]byte{SQUARE_D4, SQUARE_E4, SQUARE_D5, SQUARE_E5} {
			if boardState.IsSquareUnderAttack(sq, WHITE_MASK) {
				centerControl += PIECE_ATTACKS_CENTER_EVAL_SCORE
			} else if boardState.IsSquareUnderAttack(sq, BLACK_MASK) {
				centerControl -= PIECE_ATTACKS_CENTER_EVAL_SCORE
			}

			p := boardState.PieceAtSquare(sq)
			if p != 0x00 {
				isBlack := isPieceBlack(p)
				pieceScore := 0
				if isPawn(p) {
					pieceScore = PAWN_IN_CENTER_EVAL_SCORE
				} else {
					pieceScore = PIECE_IN_CENTER_EVAL_SCORE
				}
				if isBlack {
					centerControl -= pieceScore
				} else {
					centerControl += pieceScore
				}
			}
		}

		// prioritize king position
		blackKingSq := boardState.lookupInfo.blackKingSquare
		whiteKingSq := boardState.lookupInfo.whiteKingSquare

		if blackKingSq > SQUARE_C8 && blackKingSq < SQUARE_G8 {
			kingPosition += KING_IN_CENTER_EVAL_SCORE
		}
		if whiteKingSq > SQUARE_C1 && whiteKingSq < SQUARE_G1 {
			kingPosition -= KING_IN_CENTER_EVAL_SCORE
		}
		if !boardState.boardInfo.whiteHasCastled && !boardState.boardInfo.whiteCanCastleKingside && !boardState.boardInfo.whiteCanCastleQueenside {
			kingPosition -= KING_CANNOT_CASTLE_EVAL_SCORE
		}
		if !boardState.boardInfo.blackHasCastled && !boardState.boardInfo.blackCanCastleKingside && !boardState.boardInfo.blackCanCastleQueenside {
			kingPosition += KING_CANNOT_CASTLE_EVAL_SCORE
		}
		if boardState.boardInfo.whiteHasCastled {
			kingPosition += KING_CASTLED_EVAL_SCORE
		}
		if boardState.boardInfo.blackHasCastled {
			kingPosition -= KING_CASTLED_EVAL_SCORE
		}
		if whiteKingSq == SQUARE_G1 || whiteKingSq == SQUARE_C1 || whiteKingSq == SQUARE_B1 {
			pawns := 0
			for i := byte(9); i < 12; i++ {
				if boardState.PieceAtSquare(whiteKingSq+i) == PAWN_MASK|WHITE_MASK {
					pawns++
				}
			}
			kingPosition += pawns * KING_PAWN_COVER_EVAL_SCORE
		}
		if blackKingSq == SQUARE_G8 || whiteKingSq == SQUARE_C8 || whiteKingSq == SQUARE_B8 {
			pawns := 0
			for i := byte(9); i < 12; i++ {
				if boardState.PieceAtSquare(blackKingSq-i) == PAWN_MASK|WHITE_MASK {
					pawns++
				}
			}
			kingPosition -= pawns * KING_PAWN_COVER_EVAL_SCORE
		}
	} else {
		// prioritize king position
	}

	return BoardEval{
		phase:         boardPhase,
		material:      material,
		kingPosition:  kingPosition,
		centerControl: centerControl,
	}
}

func (eval BoardEval) value() int {
	return eval.material + eval.kingPosition + eval.centerControl
}

// TODO: incrementally update evaluation as a result of a move
