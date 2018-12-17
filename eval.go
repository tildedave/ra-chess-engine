package main

import (
	"fmt"
	"math/bits"
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
const PAWN_ON_SEVENTH_RANK_SCORE = 300

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

var pawnProtectionBoard = [2]uint64{
	0x000000000000FF00,
	0x00FF000000000000,
}

var pawnSeventhRankBoard = [2]uint64{
	0x00FF000000000000,
	0x000000000000FF00,
}

var pawnSixthRankBoard = [2]uint64{
	0x0000FF0000000000,
	0x0000000000FF0000,
}

var passedPawnBoard = [2]uint64{
	0x000000FFFFFFFF00,
	0x00FFFFFFFF000000,
}

func Eval(boardState *BoardState) (BoardEval, bool) {
	material := 0
	boardPhase := PHASE_OPENING
	if boardState.fullmoveNumber > 8 {
		boardPhase = PHASE_MIDDLEGAME
	}

	blackMaterial := 0
	whiteMaterial := 0
	blackHasPawns := false
	whiteHasPawns := false
	hasMatingMaterial := true

	whiteOccupancy := boardState.bitboards.color[WHITE_OFFSET]
	blackOccupancy := boardState.bitboards.color[BLACK_OFFSET]

	for pieceMask := byte(1); pieceMask <= 6; pieceMask++ {
		pieceBoard := boardState.bitboards.piece[pieceMask]
		whiteMaterial += bits.OnesCount64(whiteOccupancy&pieceBoard) * materialScore[pieceMask]
		blackMaterial += bits.OnesCount64(blackOccupancy&pieceBoard) * materialScore[pieceMask]

		if pieceMask == PAWN_MASK {
			whitePawns := pieceBoard & whiteOccupancy
			blackPawns := pieceBoard & blackOccupancy

			// TODO: pawn checks beyond 7th rank checks need to verify that they're passed pawns
			// or we're going to get some weird behavior in the engine

			if whitePawns != 0 {
				whiteHasPawns = true
				whiteMaterial += bits.OnesCount64(whitePawns&pawnSeventhRankBoard[WHITE_OFFSET]) * PAWN_ON_SEVENTH_RANK_SCORE

				var sixthRank = whitePawns & pawnSixthRankBoard[WHITE_OFFSET]
				if sixthRank != 0 && (sixthRank == 0x00C0000000000000 ||
					sixthRank == 0x0060000000000000 ||
					sixthRank == 0x0030000000000000 ||
					sixthRank == 0x001A000000000000 ||
					sixthRank == 0x000C000000000000 ||
					sixthRank == 0x0006000000000000 ||
					sixthRank == 0x0003000000000000) {
					whiteMaterial += ROOK_EVAL_SCORE
				}
			}

			if blackPawns != 0 {
				blackHasPawns = true
				blackMaterial += bits.OnesCount64(blackPawns&pawnSeventhRankBoard[BLACK_OFFSET]) * PAWN_ON_SEVENTH_RANK_SCORE

				var sixthRank = blackPawns & pawnSixthRankBoard[BLACK_OFFSET]
				if sixthRank != 0 && (sixthRank == 0x0000000000C00000 ||
					sixthRank == 0x0000000000600000 ||
					sixthRank == 0x0000000000300000 ||
					sixthRank == 0x00000000001A0000 ||
					sixthRank == 0x00000000000C0000 ||
					sixthRank == 0x0000000000060000 ||
					sixthRank == 0x0000000000030000) {
					blackMaterial += ROOK_EVAL_SCORE
				}
			}
		}
	}

	material = whiteMaterial - blackMaterial

	// This isn't correct, bitboards will make this easier
	if !blackHasPawns && !whiteHasPawns && blackMaterial <= KNIGHT_EVAL_SCORE && whiteMaterial <= KNIGHT_EVAL_SCORE {
		hasMatingMaterial = false
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
			squareAttackBoard := boardState.GetSquareAttackersBoard(sq)

			centerControl += (PIECE_ATTACKS_CENTER_EVAL_SCORE * bits.OnesCount64(squareAttackBoard&whiteOccupancy))
			centerControl -= (PIECE_ATTACKS_CENTER_EVAL_SCORE * bits.OnesCount64(squareAttackBoard&blackOccupancy))

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
		kings := boardState.bitboards.piece[KING_MASK]
		blackKingSq := byte(bits.TrailingZeros64(boardState.bitboards.color[BLACK_OFFSET] & kings))
		whiteKingSq := byte(bits.TrailingZeros64(boardState.bitboards.color[WHITE_OFFSET] & kings))

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
			pawns := bits.OnesCount64(boardState.moveBitboards.kingAttacks[whiteKingSq].board &
				pawnProtectionBoard[WHITE_OFFSET])
			kingPosition += pawns * KING_PAWN_COVER_EVAL_SCORE
		}
		if blackKingSq == SQUARE_G8 || blackKingSq == SQUARE_C8 || blackKingSq == SQUARE_B8 {
			pawns := bits.OnesCount64(boardState.moveBitboards.kingAttacks[blackKingSq].board &
				pawnProtectionBoard[BLACK_OFFSET])
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
	}, hasMatingMaterial
}

func (eval BoardEval) value() int {
	return eval.material + eval.kingPosition + eval.centerControl
}

// TODO: incrementally update evaluation as a result of a move
