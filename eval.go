package main

import (
	"fmt"
	"math/bits"
	"strings"
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
const KING_IN_CENTER_EVAL_SCORE = -30
const KING_PAWN_COVER_EVAL_SCORE = 10
const ENDGAME_KING_ON_EDGE_SCORE = -30
const ENDGAME_KING_NEAR_EDGE_SCORE = -10
const ENDGAME_QUEEN_BONUS_SCORE = 400
const PAWN_IN_CENTER_EVAL_SCORE = 40
const PIECE_IN_CENTER_EVAL_SCORE = 25
const PIECE_ATTACKS_CENTER_EVAL_SCORE = 15
const PAWN_ON_SEVENTH_RANK_SCORE = 300
const PAWN_PASSED_ON_SIXTH_RANK_SCORE = 200
const ISOLATED_PAWN_SCORE = -10
const DOUBLED_PAWN_SCORE = -10
const LACK_OF_DEVELOPMENT_SCORE = -15
const LACK_OF_DEVELOPMENT_SCORE_QUEEN = -5
const IMBALANCED_PIECE_SCORE = 100
const ROOK_SUPPORT_PASSED_PAWN_SCORE = 50
const BISHOP_SAME_COLOR_PAWN_SCORE = -10
const KNIGHT_SUPPORTED_BY_PAWN_SCORE = 15

var PIECE_BEHIND_BLOCKED_PAWN_SCORE = [7]int{
	0,
	0,
	-35,
	-15,
	-25,
	-25,
	0,
}

var PIECE_ATTACK_ENEMY_KING_SCORE = [7]int{
	0,  // no piece
	0,  // pawn
	30, // bishop
	30, // knight
	50, // rook
	50, // queen
	0,  // king
}

var KNIGHT_RANK_SCORE = [8]int{
	-15, // Rank 1
	-5,  // Rank 2
	0,   // Rank 3
	5,   // Rank 4
	25,  // Rank 5
	35,  // Rank 6
	20,  // Rank 7
	5,   // Final rank 8
}

var KNIGHT_COLUMN_SCORE = [8]int{
	-25, // A-column
	-10, // B-column
	0,   // C-column
	10,  // D-column
	10,  // E-column
	0,   // F-column
	-10, // G-column
	-25, // H-column
}

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
	pieceScore        [2][7]int
	pieceCount        [2][7]int
	developmentScore  int
	whiteDevelopment  int
	blackDevelopment  int
	hasMatingMaterial bool
}

type EvalInfo struct {
	numKingAttacks   [2]int
	kingAttackWeight [2]int
}

type EvalBitboards struct {
	allOccupancies uint64
	kingSquares    [2]uint64
	kingPositions  [2]byte
	blockedPawns   [2]uint64
}

var MATERIAL_SCORE = [7]int{
	0, PAWN_EVAL_SCORE, KNIGHT_EVAL_SCORE, BISHOP_EVAL_SCORE, ROOK_EVAL_SCORE, QUEEN_EVAL_SCORE, 0,
}

var pawnProtectionBoard = [2]uint64{
	0x000000000000FF00,
	0x00FF000000000000,
}

var lightSquareBoard uint64 = 0x55AA55AA55AA55AA
var darkSquareBoard uint64 = 0xAA55AA55AA55AA55

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

	var pieceBoards [2][7]uint64
	var pieceCount [2][7]int
	// var attackBoards [2][7]uint64
	var pieceScore [2][7]int
	blackMaterial := 0
	whiteMaterial := 0
	blackPawnMaterial := 0
	whitePawnMaterial := 0
	hasMatingMaterial := true

	pawnEntry := GetPawnTableEntry(boardState)
	evalBitboards := createEvalBitboards(boardState, pawnEntry)
	evalInfo := EvalInfo{}

	// These are not actually full 64-bit bitboards; just a way to encode which pieces are
	// on the board.  (Example: detect king/pawn endgames)
	var whitePieceBitboard uint64
	var blackPieceBitboard uint64

	whiteOccupancy := boardState.bitboards.color[WHITE_OFFSET]
	blackOccupancy := boardState.bitboards.color[BLACK_OFFSET]
	allOccupancies := boardState.GetAllOccupanciesBitboard()

	for pieceMask := PIECE_MASK_MIN; pieceMask <= PIECE_MASK_MAX; pieceMask++ {
		pieceBoard := boardState.bitboards.piece[pieceMask]
		whitePieceBoard := whiteOccupancy & pieceBoard
		blackPieceBoard := blackOccupancy & pieceBoard
		pieceBoards[WHITE_OFFSET][pieceMask] = whitePieceBoard
		pieceBoards[BLACK_OFFSET][pieceMask] = blackPieceBoard

		pieceCount[WHITE_OFFSET][pieceMask] = bits.OnesCount64(whitePieceBoard)
		pieceCount[BLACK_OFFSET][pieceMask] = bits.OnesCount64(blackPieceBoard)

		whitePieceMaterial := pieceCount[WHITE_OFFSET][pieceMask] * MATERIAL_SCORE[pieceMask]
		blackPieceMaterial := pieceCount[BLACK_OFFSET][pieceMask] * MATERIAL_SCORE[pieceMask]

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
		} else {
			for whitePieceBoard != 0 {
				sq := byte(bits.TrailingZeros64(whitePieceBoard))
				pieceScore[WHITE_OFFSET][pieceMask] += evalPiece(boardState, &evalBitboards, pawnEntry, &evalInfo, sq, pieceMask, WHITE_OFFSET)
				whitePieceBoard ^= 1 << sq
			}
			for blackPieceBoard != 0 {
				sq := byte(bits.TrailingZeros64(blackPieceBoard))
				pieceScore[BLACK_OFFSET][pieceMask] += evalPiece(boardState, &evalBitboards, pawnEntry, &evalInfo, sq, pieceMask, BLACK_OFFSET)
				blackPieceBoard ^= 1 << sq
			}

			// TODO: bonus for bishop pair
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

	pieceScore[WHITE_OFFSET][PAWN_MASK],
		pieceScore[BLACK_OFFSET][PAWN_MASK] = evalPawnStructure(boardState, pawnEntry, boardPhase)

	// TODO - endgame: determine passed pawns and prioritize them

	blackKingPosition := 0
	whiteKingPosition := 0
	centerControl := 0

	kings := boardState.bitboards.piece[KING_MASK]
	blackKingSq := byte(bits.TrailingZeros64(boardState.bitboards.color[BLACK_OFFSET] & kings))
	whiteKingSq := byte(bits.TrailingZeros64(boardState.bitboards.color[WHITE_OFFSET] & kings))
	var whiteDevelopment int
	var blackDevelopment int

	if boardPhase != PHASE_ENDGAME {
		// prioritize center control

		// D4, E4, D5, E5
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

		// Penalize pieces on original squares
		whiteDevelopment, blackDevelopment = evalDevelopment(boardState, boardPhase, &pieceBoards)

		// penalize king position
		if blackKingSq > SQUARE_C8 && blackKingSq < SQUARE_G8 {
			blackKingPosition += KING_IN_CENTER_EVAL_SCORE
		}
		if whiteKingSq > SQUARE_C1 && whiteKingSq < SQUARE_G1 {
			whiteKingPosition += KING_IN_CENTER_EVAL_SCORE
		}

		if whiteKingSq == SQUARE_G1 || whiteKingSq == SQUARE_C1 || whiteKingSq == SQUARE_B1 {
			pawns := bits.OnesCount64(
				boardState.moveBitboards.kingAttacks[whiteKingSq].board &
					pawnProtectionBoard[WHITE_OFFSET] &
					boardState.bitboards.piece[PAWN_MASK] &
					boardState.bitboards.color[WHITE_OFFSET])
			whiteKingPosition += pawns * KING_PAWN_COVER_EVAL_SCORE
		}
		if blackKingSq == SQUARE_G8 || blackKingSq == SQUARE_C8 || blackKingSq == SQUARE_B8 {
			pawns := bits.OnesCount64(boardState.moveBitboards.kingAttacks[blackKingSq].board &
				pawnProtectionBoard[BLACK_OFFSET] &
				boardState.bitboards.piece[PAWN_MASK] &
				boardState.bitboards.color[BLACK_OFFSET])
			blackKingPosition += pawns * KING_PAWN_COVER_EVAL_SCORE
		}
	} else {
		// prioritize king position
		if IsBitboardSet(edges, whiteKingSq) {
			whiteKingPosition += ENDGAME_KING_ON_EDGE_SCORE
		} else if IsBitboardSet(nextToEdges, whiteKingSq) {
			whiteKingPosition += ENDGAME_KING_NEAR_EDGE_SCORE
		}

		if IsBitboardSet(edges, blackKingSq) {
			blackKingPosition += ENDGAME_KING_ON_EDGE_SCORE
		} else if IsBitboardSet(nextToEdges, blackKingSq) {
			blackKingPosition += ENDGAME_KING_NEAR_EDGE_SCORE
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

	pieceScore[WHITE_OFFSET][KING_MASK] = whiteKingPosition
	pieceScore[BLACK_OFFSET][KING_MASK] = blackKingPosition

	return BoardEval{
		sideToMove:        boardState.sideToMove,
		phase:             boardPhase,
		material:          whiteMaterial - blackMaterial,
		blackMaterial:     blackMaterial,
		whiteMaterial:     whiteMaterial,
		pieceScore:        pieceScore,
		pieceCount:        pieceCount,
		developmentScore:  whiteDevelopment - blackDevelopment,
		whiteDevelopment:  whiteDevelopment,
		blackDevelopment:  blackDevelopment,
		centerControl:     centerControl,
		hasMatingMaterial: hasMatingMaterial,
	}
}

func evalPawnStructure(boardState *BoardState, pawnEntry *PawnTableEntry, boardPhase int) (int, int) {
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

	whitePawnScore += DOUBLED_PAWN_SCORE * pawnEntry.doubledPawnCount[WHITE_OFFSET]
	blackPawnScore += DOUBLED_PAWN_SCORE * pawnEntry.doubledPawnCount[BLACK_OFFSET]

	if boardPhase != PHASE_ENDGAME {
		whitePawnScore += ISOLATED_PAWN_SCORE * pawnEntry.isolatedPawnCount[WHITE_OFFSET]
		blackPawnScore += ISOLATED_PAWN_SCORE * pawnEntry.isolatedPawnCount[BLACK_OFFSET]
	}

	return whitePawnScore, blackPawnScore
}

func evalPiece(
	boardState *BoardState,
	evalBitboards *EvalBitboards,
	pawnEntry *PawnTableEntry,
	evalInfo *EvalInfo,
	sq byte,
	pieceMask byte,
	side int,
) int {
	moveBitboards := boardState.moveBitboards
	allOccupancies := evalBitboards.allOccupancies
	otherSide := BLACK_OFFSET
	if side == BLACK_OFFSET {
		otherSide = WHITE_OFFSET
	}

	var score int = 0

	switch pieceMask {
	case BISHOP_MASK:
		bishopKey := hashKey(allOccupancies, moveBitboards.bishopMagics[sq])
		attackBoard := moveBitboards.bishopAttacks[sq][bishopKey].board
		// Blocked by pawns that won't move
		// Attacking pieces
		// Attacking king
		numKingAttacks := bits.OnesCount64(attackBoard & evalBitboards.kingSquares[otherSide])
		evalInfo.numKingAttacks[side] += numKingAttacks
		evalInfo.kingAttackWeight[side] += 2 * numKingAttacks

		numBlockedPawns := bits.OnesCount64(attackBoard & evalBitboards.blockedPawns[side])
		score += int(numBlockedPawns) * PIECE_BEHIND_BLOCKED_PAWN_SCORE[BISHOP_MASK]

		if IsBitboardSet(lightSquareBoard, sq) {
			// light square bishop, de-incentivize pawns on light squares
			numSameColorPawns := bits.OnesCount64(lightSquareBoard & pawnEntry.pawns[side])
			score += int(numSameColorPawns) * BISHOP_SAME_COLOR_PAWN_SCORE
		} else {
			// dark square bishop
			numSameColorPawns := bits.OnesCount64(darkSquareBoard & pawnEntry.pawns[side])
			score += int(numSameColorPawns) * BISHOP_SAME_COLOR_PAWN_SCORE
		}
	case KNIGHT_MASK:
		attackBoard := moveBitboards.knightAttacks[sq].board
		numKingAttacks := bits.OnesCount64(attackBoard & evalBitboards.kingSquares[otherSide])
		evalInfo.numKingAttacks[side] += numKingAttacks
		evalInfo.kingAttackWeight[side] += 2 * numKingAttacks

		knightColumn := Column(sq)
		knightRank := Rank(sq)
		// Rank is between 1 and 8.  KNIGHT_RANK_SCORE is 0-indexed (0 -> 7).
		if side == WHITE_OFFSET {
			score += KNIGHT_RANK_SCORE[knightRank-1]
		} else {
			score += KNIGHT_RANK_SCORE[8-knightRank]
		}
		score += KNIGHT_COLUMN_SCORE[knightColumn-1]

	case ROOK_MASK:
		rookKey := hashKey(allOccupancies, moveBitboards.rookMagics[sq])
		attackBoard := moveBitboards.rookAttacks[sq][rookKey].board

		// Attacking near enemy king
		numKingAttacks := bits.OnesCount64(attackBoard & evalBitboards.kingSquares[otherSide])
		evalInfo.numKingAttacks[side] += numKingAttacks
		evalInfo.kingAttackWeight[side] += 3 * numKingAttacks

		// Blocked pawn
		numBlockedPawns := bits.OnesCount64(attackBoard & evalBitboards.blockedPawns[side])
		score += int(numBlockedPawns) * PIECE_BEHIND_BLOCKED_PAWN_SCORE[ROOK_MASK]

		// Bonus on open file

		// Bonus on half-open file (only enemy pawns)

		// Bonus on 6th/7th rank

		// Passed pawn support
		passedPawnSupportBoard := attackBoard & pawnEntry.passedPawns[side]
		if passedPawnSupportBoard != 0 {
			rookCol := Column(sq)
			for passedPawnSupportBoard != 0 {
				pawnSq := bits.TrailingZeros64(passedPawnSupportBoard)
				if Column(uint8(pawnSq)) == rookCol {
					score += ROOK_SUPPORT_PASSED_PAWN_SCORE
				}
				passedPawnSupportBoard ^= 1 << pawnSq
			}
		}

	case QUEEN_MASK:
		rookKey := hashKey(allOccupancies, moveBitboards.rookMagics[sq])
		bishopKey := hashKey(allOccupancies, moveBitboards.bishopMagics[sq])
		attackBoard := (moveBitboards.rookAttacks[sq][rookKey].board | moveBitboards.bishopAttacks[sq][bishopKey].board)
		numKingAttacks := bits.OnesCount64(attackBoard & evalBitboards.kingSquares[otherSide])
		evalInfo.numKingAttacks[side] += numKingAttacks
		evalInfo.kingAttackWeight[side] += 4 * numKingAttacks

	case KING_MASK:
	}

	return score
}

func evalDevelopment(boardState *BoardState, boardPhase int, pieceBoards *[2][7]uint64) (int, int) {
	whiteDevelopment := 0
	blackDevelopment := 0

	if boardState.board[SQUARE_B1] == KNIGHT_MASK|WHITE_MASK {
		whiteDevelopment += LACK_OF_DEVELOPMENT_SCORE
	}
	if boardState.board[SQUARE_C1] == BISHOP_MASK|WHITE_MASK {
		whiteDevelopment += LACK_OF_DEVELOPMENT_SCORE
	}
	if boardState.board[SQUARE_F1] == BISHOP_MASK|WHITE_MASK {
		whiteDevelopment += LACK_OF_DEVELOPMENT_SCORE
	}
	if boardState.board[SQUARE_G1] == KNIGHT_MASK|WHITE_MASK {
		whiteDevelopment += LACK_OF_DEVELOPMENT_SCORE
	}

	// Now black
	if boardState.board[SQUARE_B8] == KNIGHT_MASK|BLACK_MASK {
		blackDevelopment += LACK_OF_DEVELOPMENT_SCORE
	}
	if boardState.board[SQUARE_C8] == BISHOP_MASK|BLACK_MASK {
		blackDevelopment += LACK_OF_DEVELOPMENT_SCORE
	}
	if boardState.board[SQUARE_F8] == BISHOP_MASK|BLACK_MASK {
		blackDevelopment += LACK_OF_DEVELOPMENT_SCORE
	}
	if boardState.board[SQUARE_G8] == KNIGHT_MASK|BLACK_MASK {
		blackDevelopment += LACK_OF_DEVELOPMENT_SCORE
	}

	if boardPhase == PHASE_MIDDLEGAME {
		if boardState.board[SQUARE_D1] == QUEEN_MASK|WHITE_MASK {
			whiteDevelopment += LACK_OF_DEVELOPMENT_SCORE_QUEEN
		}
		if boardState.board[SQUARE_D8] == QUEEN_MASK|BLACK_MASK {
			blackDevelopment += LACK_OF_DEVELOPMENT_SCORE_QUEEN
		}
	}

	return whiteDevelopment, blackDevelopment
}

func (eval BoardEval) value() int {
	totalPieceScore := 0
	for i := PIECE_MASK_MIN; i <= PIECE_MASK_MAX; i++ {
		totalPieceScore += eval.pieceScore[WHITE_OFFSET][i] - eval.pieceScore[BLACK_OFFSET][i]
	}
	score := eval.material + eval.centerControl + totalPieceScore + eval.developmentScore
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

	pieces := ""
	for j := PAWN_MASK; j <= KING_MASK; j++ {
		if eval.pieceCount[WHITE_OFFSET][j] == 0 && eval.pieceCount[BLACK_OFFSET][j] == 0 {
			continue
		}

		for mask := WHITE_OFFSET; mask <= BLACK_OFFSET; mask++ {
			if eval.pieceCount[mask][j] != 0 {
				pieceStr := ""
				for i := 0; i < eval.pieceCount[mask][j]; i++ {
					m := WHITE_MASK
					if mask == BLACK_OFFSET {
						m = BLACK_MASK
					}
					pieceStr += string(pieceToString(j | m))
				}
				pieces += fmt.Sprintf("%d (%s) ", eval.pieceScore[mask][j], pieceStr)
			}
		}
	}

	totalPieceScore := 0
	for i := PIECE_MASK_MIN; i <= PIECE_MASK_MAX; i++ {
		totalPieceScore += eval.pieceScore[WHITE_OFFSET][i] - eval.pieceScore[BLACK_OFFSET][i]
	}

	return fmt.Sprintf("VALUE: %d\n\tphase=%s\n\tmaterial=%d (white: %d, black: %d)\n\tpieces=%d [%s]\n\tdevelopment=%d (white: %d, black: %d)\n\tcenterControl=%d",
		eval.value(),
		phaseString,
		eval.material,
		eval.whiteMaterial,
		eval.blackMaterial,
		totalPieceScore,
		strings.Trim(pieces, " "),
		eval.developmentScore,
		eval.whiteDevelopment,
		eval.blackDevelopment,
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

func createEvalBitboards(boardState *BoardState, pawnEntry *PawnTableEntry) EvalBitboards {
	var whiteKingSq byte
	var blackKingSq byte
	kingBoard := boardState.bitboards.piece[KING_MASK]
	sq := byte(bits.TrailingZeros64(kingBoard))
	if boardState.board[sq] == WHITE_MASK|KING_MASK {
		whiteKingSq = sq
		kingBoard ^= 1 << sq
		blackKingSq = byte(bits.TrailingZeros64(kingBoard))
	} else {
		blackKingSq = sq
		kingBoard ^= 1 << sq
		whiteKingSq = byte(bits.TrailingZeros64(kingBoard))
	}

	allOccupancies := boardState.GetAllOccupanciesBitboard()

	// Find pawns that cannot advance

	return EvalBitboards{
		allOccupancies: allOccupancies,
		kingPositions:  [2]byte{whiteKingSq, blackKingSq},
		kingSquares: [2]uint64{
			moveBitboards.kingAttacks[whiteKingSq].board,
			moveBitboards.kingAttacks[blackKingSq].board,
		},
		blockedPawns: [2]uint64{
			(pawnEntry.pawns[WHITE_OFFSET] << 8 & allOccupancies) >> 8,
			(pawnEntry.pawns[BLACK_OFFSET] >> 8 & allOccupancies) << 8,
		},
	}
}
