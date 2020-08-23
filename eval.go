package main

import (
	"fmt"
	"math/bits"
)

type EvalOptions struct {
	epdRegex string
}

const ENDGAME_MATERIAL_THRESHOLD = 1500

const QUEEN_EVAL_SCORE = 800
const PAWN_EVAL_SCORE = 100
const ROOK_EVAL_SCORE = 500
const KNIGHT_EVAL_SCORE = 300
const BISHOP_EVAL_SCORE = 300
const KING_IN_CENTER_EVAL_SCORE = 50
const KING_PAWN_COVER_EVAL_SCORE = 10
const KING_CANNOT_CASTLE_EVAL_SCORE = 70
const ENDGAME_KING_ON_EDGE_PENALTY = 30
const ENDGAME_KING_NEAR_EDGE_PENALTY = 10
const ENDGAME_QUEEN_BONUS_SCORE = 400
const KING_CASTLED_EVAL_SCORE = 20
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
	sideToMove        int
	material          int
	phase             int
	centerControl     int
	whiteMaterial     int
	blackMaterial     int
	development       int
	kingPosition      int
	whiteKingPosition int
	blackKingPosition int
	passedPawns       int
	hasMatingMaterial bool
}

var MATERIAL_SCORE = [7]int{
	0, PAWN_EVAL_SCORE, KNIGHT_EVAL_SCORE, BISHOP_EVAL_SCORE, ROOK_EVAL_SCORE, QUEEN_EVAL_SCORE, 0,
}

var pawnProtectionBoard = [2]uint64{
	0x000000000000FF00,
	0x00FF000000000000,
}

var pawnSixthRankBoard = [2]uint64{
	0x0000FF0000000000,
	0x0000000000FF0000,
}

var edges uint64 = 0xFF818181818181FF
var nextToEdges uint64 = 0x007E424242427E00

func Eval(boardState *BoardState) BoardEval {
	boardPhase := PHASE_OPENING
	if boardState.fullmoveNumber > 8 {
		boardPhase = PHASE_MIDDLEGAME
	}

	blackMaterial := 0
	whiteMaterial := 0
	blackHasPawns := false
	whiteHasPawns := false
	whiteHasQueen := false
	blackHasQueen := false
	hasMatingMaterial := true

	whiteOccupancy := boardState.bitboards.color[WHITE_OFFSET]
	blackOccupancy := boardState.bitboards.color[BLACK_OFFSET]
	pawnEntry := GetPawnTableEntry(boardState)

	for pieceMask := byte(1); pieceMask <= 6; pieceMask++ {
		pieceBoard := boardState.bitboards.piece[pieceMask]
		whitePieceBoard := whiteOccupancy & pieceBoard
		blackPieceBoard := blackOccupancy & pieceBoard
		whiteMaterial += bits.OnesCount64(whitePieceBoard) * MATERIAL_SCORE[pieceMask]
		blackMaterial += bits.OnesCount64(blackPieceBoard) * MATERIAL_SCORE[pieceMask]

		if pieceMask == PAWN_MASK {
			whitePawns := pieceBoard & whiteOccupancy
			blackPawns := pieceBoard & blackOccupancy

			// TODO: pawn checks beyond 7th rank checks need to verify that they're passed pawns
			// or we're going to get some weird behavior in the engine

			if whitePawns != 0 {
				whiteHasPawns = true
				whiteMaterial += bits.OnesCount64(pawnEntry.pawnsPerRank[WHITE_OFFSET][RANK_7]) * PAWN_ON_SEVENTH_RANK_SCORE
				// This is prioritizing connected pawns on the 6th rank
				var sixthRank = pawnEntry.pawnsPerRank[WHITE_OFFSET][RANK_6]
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
				blackMaterial += bits.OnesCount64(pawnEntry.pawnsPerRank[BLACK_OFFSET][RANK_2]) * PAWN_ON_SEVENTH_RANK_SCORE
				// This is prioritizing connected pawns on the 3rd rank
				var sixthRank = pawnEntry.pawnsPerRank[BLACK_OFFSET][RANK_3]
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
		} else if pieceMask == QUEEN_MASK {
			whiteHasQueen = whitePieceBoard != 0
			blackHasQueen = blackPieceBoard != 0
		}
	}

	// This isn't correct, bitboards will make this easier
	if !blackHasPawns && !whiteHasPawns && blackMaterial <= KNIGHT_EVAL_SCORE && whiteMaterial <= KNIGHT_EVAL_SCORE {
		hasMatingMaterial = false
	}

	// TODO: pawns probably shouldn't count for this
	if blackMaterial < ENDGAME_MATERIAL_THRESHOLD && whiteMaterial < ENDGAME_MATERIAL_THRESHOLD {
		boardPhase = PHASE_ENDGAME
	}

	// TODO - endgame: determine passed pawns and prioritize them

	blackKingPosition := 0
	whiteKingPosition := 0
	centerControl := 0

	kings := boardState.bitboards.piece[KING_MASK]
	blackKingSq := byte(bits.TrailingZeros64(boardState.bitboards.color[BLACK_OFFSET] & kings))
	whiteKingSq := byte(bits.TrailingZeros64(boardState.bitboards.color[WHITE_OFFSET] & kings))

	if boardPhase != PHASE_ENDGAME {
		// prioritize center control

		// D4, E4, D5, E5
		allOccupancies := boardState.GetAllOccupanciesBitboard()
		for _, sq := range [4]byte{SQUARE_D4, SQUARE_E4, SQUARE_D5, SQUARE_E5} {
			squareAttackBoard := boardState.GetSquareAttackersBoard(allOccupancies, sq)

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

		// penalize king position
		if blackKingSq > SQUARE_C8 && blackKingSq < SQUARE_G8 {
			blackKingPosition -= KING_IN_CENTER_EVAL_SCORE
		}
		if whiteKingSq > SQUARE_C1 && whiteKingSq < SQUARE_G1 {
			whiteKingPosition -= KING_IN_CENTER_EVAL_SCORE
		}
		if !boardState.boardInfo.whiteHasCastled && !boardState.boardInfo.whiteCanCastleKingside && !boardState.boardInfo.whiteCanCastleQueenside {
			whiteKingPosition -= KING_CANNOT_CASTLE_EVAL_SCORE
		}
		if !boardState.boardInfo.blackHasCastled && !boardState.boardInfo.blackCanCastleKingside && !boardState.boardInfo.blackCanCastleQueenside {
			blackKingPosition -= KING_CANNOT_CASTLE_EVAL_SCORE
		}
		if boardState.boardInfo.whiteHasCastled {
			whiteKingPosition += KING_CASTLED_EVAL_SCORE
		}
		if boardState.boardInfo.blackHasCastled {
			blackKingPosition += KING_CASTLED_EVAL_SCORE
		}
		if whiteKingSq == SQUARE_G1 || whiteKingSq == SQUARE_C1 || whiteKingSq == SQUARE_B1 {
			pawns := bits.OnesCount64(boardState.moveBitboards.kingAttacks[whiteKingSq].board &
				pawnProtectionBoard[WHITE_OFFSET])
			whiteKingPosition += pawns * KING_PAWN_COVER_EVAL_SCORE
		}
		if blackKingSq == SQUARE_G8 || blackKingSq == SQUARE_C8 || blackKingSq == SQUARE_B8 {
			pawns := bits.OnesCount64(boardState.moveBitboards.kingAttacks[blackKingSq].board &
				pawnProtectionBoard[BLACK_OFFSET])
			blackKingPosition += pawns * KING_PAWN_COVER_EVAL_SCORE
		}
	} else {
		// prioritize king position

		if IsBitboardSet(edges, whiteKingSq) {
			whiteKingPosition -= ENDGAME_KING_ON_EDGE_PENALTY
		} else if IsBitboardSet(nextToEdges, whiteKingSq) {
			whiteKingPosition -= ENDGAME_KING_NEAR_EDGE_PENALTY
		}

		if IsBitboardSet(edges, blackKingSq) {
			blackKingPosition -= ENDGAME_KING_ON_EDGE_PENALTY
		} else if IsBitboardSet(nextToEdges, blackKingSq) {
			blackKingPosition -= ENDGAME_KING_NEAR_EDGE_PENALTY
		}

		// if you have a queen and enemy doesn't that's a good thing
		if whiteHasQueen && !blackHasQueen {
			whiteMaterial += ENDGAME_QUEEN_BONUS_SCORE
		} else if blackHasQueen && !whiteHasQueen {
			blackMaterial += ENDGAME_QUEEN_BONUS_SCORE
		}
	}

	return BoardEval{
		sideToMove:        boardState.sideToMove,
		phase:             boardPhase,
		material:          whiteMaterial - blackMaterial,
		blackMaterial:     blackMaterial,
		whiteMaterial:     whiteMaterial,
		kingPosition:      whiteKingPosition - blackKingPosition,
		whiteKingPosition: whiteKingPosition,
		blackKingPosition: blackKingPosition,
		centerControl:     centerControl,
		hasMatingMaterial: hasMatingMaterial,
	}
}

func (eval BoardEval) value() int {
	score := eval.material + eval.kingPosition + eval.centerControl
	if eval.sideToMove == BLACK_OFFSET {
		return -score
	}
	return score
}

func BoardEvalToString(eval BoardEval) string {
	var phaseString string
	switch eval.phase {
	case PHASE_ENDGAME:
		phaseString = "endgame"
	case PHASE_MIDDLEGAME:
		phaseString = "middlegame"
	case PHASE_OPENING:
		phaseString = "opening"
	}

	return fmt.Sprintf("phase=%s, material=%d (white: %d, black: %d), kingPosition=%d, centerControl=%d",
		phaseString,
		eval.material,
		eval.whiteMaterial,
		eval.blackMaterial,
		eval.kingPosition,
		eval.centerControl)
}

func RunEvalFile(epdFile string, variation string, options EvalOptions) (bool, error) {
	lines, err := ParseAndFilterEpdFile(epdFile, options.epdRegex)
	if err != nil {
		return false, err
	}

	if variation != "" && len(lines) > 1 {
		return false, fmt.Errorf("Can only specify variation if regex filters to 1 positions, got %d", len(lines))
	}

	for _, line := range lines {
		boardEval, err := RunEvalFen(line.fen, variation, options)
		if err != nil {
			return false, err
		}

		fmt.Println(BoardEvalToString(boardEval))
	}

	return true, nil
}

func RunEvalFen(fen string, variation string, options EvalOptions) (BoardEval, error) {
	boardState, err := CreateBoardStateFromFENStringWithVariation(fen, variation)
	if err != nil {
		return BoardEval{}, nil
	}

	fmt.Println(boardState.ToString())

	return Eval(&boardState), nil
}

// TODO: incrementally update evaluation as a result of a move
