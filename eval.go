package main

import (
	"fmt"
	"math/bits"
)

type EvalOptions struct {
	epdRegex string
}

const ENDGAME_MATERIAL_THRESHOLD = 1200

const QUEEN_EVAL_SCORE = 800
const PAWN_EVAL_SCORE = 100
const ROOK_EVAL_SCORE = 500
const KNIGHT_EVAL_SCORE = 300
const BISHOP_EVAL_SCORE = 320
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
const PAWN_PASSED_ON_SIXTH_RANK_SCORE = 200
const ISOLATED_PAWN_SCORE = -20
const DOUBLED_PAWN_SCORE = -10

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
	whitePawnScore    int
	blackPawnScore    int
	pawnScore         int
	development       int
	kingPosition      int
	whiteKingPosition int
	blackKingPosition int
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

var passedPawnByRankScore = [8]int{
	0,
	0,   // RANK_1
	20,  // RANK_2
	30,  // RANK_3
	40,  // RANK_4
	60,  // RANK_5
	100, // RANK_6
	120, // RANK_7
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
	blackPawnMaterial := 0
	whitePawnMaterial := 0
	hasMatingMaterial := true

	// These are not actually full 64-bit bitboards; just a way to encode which pieces are
	// on the board.  (Example: detect king/pawn endgames)
	var whitePieceBitboard uint64
	var blackPieceBitboard uint64

	whiteOccupancy := boardState.bitboards.color[WHITE_OFFSET]
	blackOccupancy := boardState.bitboards.color[BLACK_OFFSET]

	for pieceMask := byte(1); pieceMask <= 6; pieceMask++ {
		pieceBoard := boardState.bitboards.piece[pieceMask]
		whitePieceBoard := whiteOccupancy & pieceBoard
		blackPieceBoard := blackOccupancy & pieceBoard

		whitePieceMaterial := bits.OnesCount64(whitePieceBoard) * MATERIAL_SCORE[pieceMask]
		blackPieceMaterial := bits.OnesCount64(blackPieceBoard) * MATERIAL_SCORE[pieceMask]

		whiteMaterial += whitePieceMaterial
		blackMaterial += blackPieceMaterial

		if whitePieceBoard != 0 {
			whitePieceBitboard = SetBitboard(whitePieceBitboard, pieceMask)
		}
		if blackPieceBoard != 0 {
			blackPieceBitboard = SetBitboard(blackPieceBitboard, pieceMask)
		}

		if pieceMask == PAWN_MASK {
			whitePawnMaterial = whitePieceMaterial
			blackPawnMaterial = blackPieceMaterial
		}
	}

	// This isn't correct, bitboards will make this easier
	if !IsBitboardSet(whitePieceBitboard, PAWN_MASK) &&
		!IsBitboardSet(blackPieceBitboard, PAWN_MASK) &&
		blackMaterial <= KNIGHT_EVAL_SCORE &&
		whiteMaterial <= KNIGHT_EVAL_SCORE {
		hasMatingMaterial = false
	}

	if blackMaterial-blackPawnMaterial < ENDGAME_MATERIAL_THRESHOLD &&
		whiteMaterial-whitePawnMaterial < ENDGAME_MATERIAL_THRESHOLD {
		boardPhase = PHASE_ENDGAME
	}

	whitePawnScore, blackPawnScore := evalPawnStructure(boardState, boardPhase)

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
		blackHasQueen := IsBitboardSet(blackPieceBitboard, QUEEN_MASK)
		whiteHasQueen := IsBitboardSet(whitePieceBitboard, QUEEN_MASK)

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
		whitePawnScore:    whitePawnScore,
		blackPawnScore:    blackPawnScore,
		kingPosition:      whiteKingPosition - blackKingPosition,
		pawnScore:         whitePawnScore - blackPawnScore,
		whiteKingPosition: whiteKingPosition,
		blackKingPosition: blackKingPosition,
		centerControl:     centerControl,
		hasMatingMaterial: hasMatingMaterial,
	}
}

func evalPawnStructure(boardState *BoardState, boardPhase int) (int, int) {
	pawnEntry := GetPawnTableEntry(boardState)
	whitePawnScore := 0
	blackPawnScore := 0

	whitePassers := pawnEntry.passedPawns[WHITE_OFFSET]
	blackPassers := pawnEntry.passedPawns[BLACK_OFFSET]

	for _, rank := range []byte{RANK_5, RANK_6, RANK_7} {
		rankWhitePawns := pawnEntry.pawnsPerRank[WHITE_OFFSET][rank]
		rankBlackPawns := pawnEntry.pawnsPerRank[BLACK_OFFSET][8-rank+1]
		whitePawnScore += passedPawnByRankScore[rank] * bits.OnesCount64(whitePassers&rankWhitePawns)
		blackPawnScore += passedPawnByRankScore[rank] * bits.OnesCount64(blackPassers&rankBlackPawns)
	}

	whitePawnScore += DOUBLED_PAWN_SCORE * bits.OnesCount64(pawnEntry.isolatedPawns[WHITE_OFFSET])
	blackPawnScore += DOUBLED_PAWN_SCORE * bits.OnesCount64(pawnEntry.isolatedPawns[BLACK_OFFSET])

	if boardPhase != PHASE_ENDGAME {
		whiteIsolatedPawnCount := bits.OnesCount64(pawnEntry.isolatedPawns[WHITE_OFFSET])
		blackIsolatedPawnCount := bits.OnesCount64(pawnEntry.isolatedPawns[BLACK_OFFSET])
		whitePawnScore += ISOLATED_PAWN_SCORE * whiteIsolatedPawnCount
		blackPawnScore += ISOLATED_PAWN_SCORE * blackIsolatedPawnCount
	}

	return whitePawnScore, blackPawnScore
}

func (eval BoardEval) value() int {
	score := eval.material + eval.kingPosition + eval.centerControl + eval.pawnScore
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
