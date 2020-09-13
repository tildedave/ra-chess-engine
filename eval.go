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

const KING_IN_CENTER_EVAL_SCORE = -30
const KING_PAWN_COVER_EVAL_SCORE = 10
const KING_OPEN_FILE_EVAL_SCORE = -30
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
const TEMPO_SCORE = 10

var BISHOP_SAME_COLOR_PAWN_SCORE = [9]int{
	15, 10, 5, 0, -5, -20, -30, -40, -50,
}

const KNIGHT_SUPPORTED_BY_PAWN_SCORE = 15
const KNIGHT_ON_ENEMY_HOLE_SCORE = 40
const KNIGHT_ATTACK_ENEMY_HOLE_SCORE = 15
const ROOK_OPEN_FILE_SCORE = 45
const ROOK_HALF_OPEN_FILE_SCORE = 30
const ROOK_DOUBLED_SCORE = 15

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

var KING_PAWN_COVER_SCORE = [9]int{
	-40, // no pawn cover
	-30, // 1 pawn
	-10, // 2 pawns
	0,   // 3 pawns
	0,   // 4 and on
	0, 0, 0, 0,
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

var ROOK_RANK_SCORE = [8]int{
	0, // Rank 1
	0,
	0,
	0,
	0,
	15, // Rank 6
	35, // Rank 7
	0,  // Rank 8
}

var MOBILITY_SCORE = [2][8][]int{
	// MIDDLEGAME
	[8][]int{
		[]int{}, // no piece
		[]int{}, // pawn - not used
		[]int{
			-16, -12, -8, -4, 0, 4, 8, 12, 16,
		}, // knight - up to 8 moves
		[]int{
			-21, // "A blockaded bishop is of little value" - Lisa Simpson
			-18, -15, -12, -9, -6, -3, 0, 3, 6, 9, 12, 15, 18,
		}, // bishop - up to 13 moves
		[]int{
			-14, -12, -10, -8, -6, -4, -2, 0, 2, 4, 6, 8, 10, 12, 14,
		}, // rook - up to 14 moves
		[]int{
			-28, -26, -24, -22, -20, -18, -16, -14, -12, -10, -8, -6, -4, -2, 0,
			2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30,
		}, // queen - up to 27 moves
		[]int{
			0, 0, 0, 0, 0, 0, 0, 0, 0,
		}, // king
	},
	// ENDGAME
	[8][]int{
		[]int{}, // no piece
		[]int{}, // pawn - not used
		[]int{
			-16, -12, -8, -4, 0, 4, 8, 12, 16,
		}, // knight
		[]int{
			-21, // "A blockaded bishop is of little value" - Lisa Simpson
			-18, -15, -12, -9, -6, -3, 0, 3, 6, 9, 12, 15, 18,
		}, // bishop
		[]int{
			-28, -24, -20, -16, -12, -8, -4, 0, 4, 8, 12, 16, 20, 24, 28,
		}, // rook
		[]int{
			-42, -30, -25, -21, -16, -8, -4, 0, 4, 8, 12, 16, 20, 24,
			28, 32, 36, 40, 44, 48, 52, 56, 56, 56, 56, 56, 56, 56, 56,
		}, // queen
		[]int{
			-10, -6, -2, 2, 6, 12, 16, 20, 24,
		}, // king
	},
}

// From https://www.chessprogramming.org/CPW-Engine_eval
// SafetyTable
var ATTACK_SCORE = [100]int{
	0, 0, 1, 2, 3, 5, 7, 9, 12, 15,
	18, 22, 26, 30, 35, 39, 44, 50, 56, 62,
	68, 75, 82, 85, 89, 97, 105, 113, 122, 131,
	140, 150, 169, 180, 191, 202, 213, 225, 237, 248,
	260, 272, 283, 295, 307, 319, 330, 342, 354, 366,
	377, 389, 401, 412, 424, 436, 448, 459, 471, 483,
	494, 500, 500, 500, 500, 500, 500, 500, 500, 500,
	500, 500, 500, 500, 500, 500, 500, 500, 500, 500,
	500, 500, 500, 500, 500, 500, 500, 500, 500, 500,
	500, 500, 500, 500, 500, 500, 500, 500, 500, 500,
}

var KNIGHT_PAWN_ADJUST = [9]int{-20, -16, -12, -8, -4, 0, 4, 8, 12}
var ROOK_PAWN_ADJUST = [9]int{15, 12, 9, 6, 3, 0, -3, -6, -9}

const (
	PHASE_OPENING    = iota
	PHASE_MIDDLEGAME = iota
	PHASE_ENDGAME    = iota
)

type BoardEval struct {
	gamePhase         int
	sideToMove        int
	phase             int
	centerControl     int
	material          [2][2]int
	nonPawnMaterial   [2][2]int
	mobility          [2][2]int
	pieceScore        [2][7]int
	pieceCount        [2][7]int
	kingAttackScore   int
	developmentScore  int
	whiteDevelopment  int
	blackDevelopment  int
	hasMatingMaterial bool
}

type EvalInfo struct {
	numKingAttacks   [2]int
	numKingAttackers [2]int
	kingAttackWeight [2]int
}

type EvalBitboards struct {
	allOccupancies uint64
	kingSquares    [2]uint64
	kingPositions  [2]byte
	blockedPawns   [2]uint64
}

var MATERIAL_SCORE = [2][7]int{
	// Middlegame
	[7]int{
		0,
		100,  // PAWN_EVAL_SCORE,
		320,  // KNIGHT_EVAL_SCORE
		330,  // BISHOP_EVAL_SCORE
		500,  // ROOK_EVAL_SCORE
		1050, // QUEEN_EVAL_SCORE,
		0,    // king
	},
	// Endgame
	[7]int{
		0,
		180,  // Pawns: more valuable in endgame
		320,  // Knights: less good in endgame (relative to other pieces)
		400,  // Bishops
		760,  // Rooks: needed to mate
		1200, // Queens are good
	},
}

var pawnProtectionBoard = [2]uint64{
	0x0000000000C3FF00,
	0x00FFC30000000000,
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
	// var attackBoards [2][7]uint64
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

	eval := BoardEval{
		sideToMove: boardState.sideToMove,
	}

	for pieceMask := PIECE_MASK_MIN; pieceMask <= PIECE_MASK_MAX; pieceMask++ {
		pieceBoard := boardState.bitboards.piece[pieceMask]
		whitePieceBoard := whiteOccupancy & pieceBoard
		blackPieceBoard := blackOccupancy & pieceBoard
		pieceBoards[WHITE_OFFSET][pieceMask] = whitePieceBoard
		pieceBoards[BLACK_OFFSET][pieceMask] = blackPieceBoard

		if whitePieceBoard != 0 {
			whitePieceBitboard = SetBitboard(whitePieceBitboard, pieceMask)
		}
		if blackPieceBoard != 0 {
			blackPieceBitboard = SetBitboard(blackPieceBitboard, pieceMask)
		}

		for whitePieceBoard != 0 {
			sq := byte(bits.TrailingZeros64(whitePieceBoard))
			evalPiece(&eval, boardState, &evalBitboards, pawnEntry, &evalInfo, sq, pieceMask, WHITE_OFFSET)
			whitePieceBoard ^= 1 << sq
		}
		for blackPieceBoard != 0 {
			sq := byte(bits.TrailingZeros64(blackPieceBoard))
			evalPiece(&eval, boardState, &evalBitboards, pawnEntry, &evalInfo, sq, pieceMask, BLACK_OFFSET)
			blackPieceBoard ^= 1 << sq
		}
	}

	// No pawns, rooks, queens.
	// 1 bishop/1 knight alone cannot mate
	// 2 bishops can mate
	// 1 knight + 1 bishop can mate
	// 2 knights cannot mate
	checkForInsufficientMatingMaterial(&eval, whitePieceBitboard, blackPieceBitboard)

	// TODO - endgame calculation

	evalPawnStructure(&eval, boardState, pawnEntry, boardPhase)

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
			kingAttackBoard := SetBitboardMultiple(
				boardState.moveBitboards.kingAttacks[whiteKingSq].board,
				SQUARE_A3,
				SQUARE_G3,
				SQUARE_H3)
			pawns := bits.OnesCount64(
				kingAttackBoard &
					pawnProtectionBoard[WHITE_OFFSET] &
					boardState.bitboards.piece[PAWN_MASK] &
					boardState.bitboards.color[WHITE_OFFSET])
			whiteKingPosition += KING_PAWN_COVER_SCORE[pawns]
		}
		if blackKingSq == SQUARE_G8 || blackKingSq == SQUARE_C8 || blackKingSq == SQUARE_B8 {
			kingAttackBoard := SetBitboardMultiple(
				boardState.moveBitboards.kingAttacks[blackKingPosition].board,
				SQUARE_A6,
				SQUARE_G6,
				SQUARE_H6)
			pawns := bits.OnesCount64(kingAttackBoard &
				pawnProtectionBoard[BLACK_OFFSET] &
				boardState.bitboards.piece[PAWN_MASK] &
				boardState.bitboards.color[BLACK_OFFSET])
			blackKingPosition += KING_PAWN_COVER_SCORE[pawns]
		}

		if IsBitboardSet(pawnEntry.openFileBoard, whiteKingSq) {
			whiteKingPosition += KING_OPEN_FILE_EVAL_SCORE
		}
		if IsBitboardSet(pawnEntry.openFileBoard, blackKingSq) {
			blackKingPosition += KING_OPEN_FILE_EVAL_SCORE
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
		// blackHasQueen := IsBitboardSet(blackPieceBitboard, QUEEN_MASK)
		// whiteHasQueen := IsBitboardSet(whitePieceBitboard, QUEEN_MASK)

		// if whiteHasQueen && !blackHasQueen {
		// 	whiteMaterial += ENDGAME_QUEEN_BONUS_SCORE
		// } else if blackHasQueen && !whiteHasQueen {
		// 	blackMaterial += ENDGAME_QUEEN_BONUS_SCORE
		// }
	}

	eval.pieceScore[WHITE_OFFSET][KING_MASK] = whiteKingPosition
	eval.pieceScore[BLACK_OFFSET][KING_MASK] = blackKingPosition

	var kingAttackScore int
	for side := WHITE_OFFSET; side <= BLACK_OFFSET; side++ {
		if eval.pieceCount[side][QUEEN_MASK] > 0 && evalInfo.numKingAttackers[side] > 1 {
			if side == WHITE_OFFSET {
				kingAttackScore += ATTACK_SCORE[evalInfo.kingAttackWeight[side]]
			} else {
				kingAttackScore -= ATTACK_SCORE[evalInfo.kingAttackWeight[side]]
			}
		}
	}

	eval.kingAttackScore = kingAttackScore
	eval.developmentScore = whiteDevelopment - blackDevelopment
	eval.whiteDevelopment = whiteDevelopment
	eval.blackDevelopment = blackDevelopment
	eval.centerControl = centerControl
	eval.hasMatingMaterial = hasMatingMaterial
	eval.phase = boardPhase

	return eval
}

func evalPawnStructure(
	eval *BoardEval,
	boardState *BoardState,
	pawnEntry *PawnTableEntry,
	boardPhase int,
) {
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

	eval.pieceScore[WHITE_OFFSET][PAWN_MASK] = whitePawnScore
	eval.pieceScore[BLACK_OFFSET][PAWN_MASK] = blackPawnScore
}

func evalPiece(
	eval *BoardEval,
	boardState *BoardState,
	evalBitboards *EvalBitboards,
	pawnEntry *PawnTableEntry,
	evalInfo *EvalInfo,
	sq byte,
	pieceMask byte,
	side int,
) {
	moveBitboards := boardState.moveBitboards
	allOccupancies := evalBitboards.allOccupancies
	otherSide := BLACK_OFFSET
	if side == BLACK_OFFSET {
		otherSide = WHITE_OFFSET
	}

	var score int = 0

	eval.pieceCount[side][pieceMask]++
	for j := 0; j <= 1; j++ {
		eval.material[j][side] += MATERIAL_SCORE[j][pieceMask]
		if pieceMask != PAWN_MASK {
			eval.nonPawnMaterial[j][side] += MATERIAL_SCORE[j][pieceMask]
		}
	}
	var attackBoard uint64

	switch pieceMask {
	case BISHOP_MASK:
		eval.gamePhase++
		bishopKey := hashKey(allOccupancies, moveBitboards.bishopMagics[sq])
		attackBoard = moveBitboards.bishopAttacks[sq][bishopKey].board
		// Blocked by pawns that won't move
		// Attacking pieces
		// Attacking king
		numKingAttacks := bits.OnesCount64(attackBoard & evalBitboards.kingSquares[otherSide])
		if numKingAttacks != 0 {
			evalInfo.numKingAttackers[side]++
			evalInfo.numKingAttacks[side] += numKingAttacks
			evalInfo.kingAttackWeight[side] += 2 * numKingAttacks
		}

		numBlockedPawns := bits.OnesCount64(attackBoard & evalBitboards.blockedPawns[side])
		score += int(numBlockedPawns) * PIECE_BEHIND_BLOCKED_PAWN_SCORE[BISHOP_MASK]

		if IsBitboardSet(lightSquareBoard, sq) {
			// light square bishop, de-incentivize pawns on light squares
			numSameColorPawns := bits.OnesCount64(lightSquareBoard & pawnEntry.pawns[side])
			score += BISHOP_SAME_COLOR_PAWN_SCORE[int(numSameColorPawns)]
		} else {
			// dark square bishop
			numSameColorPawns := bits.OnesCount64(darkSquareBoard & pawnEntry.pawns[side])
			score += BISHOP_SAME_COLOR_PAWN_SCORE[int(numSameColorPawns)]
		}
	case KNIGHT_MASK:
		eval.gamePhase++
		attackBoard = moveBitboards.knightAttacks[sq].board
		numKingAttacks := bits.OnesCount64(attackBoard & evalBitboards.kingSquares[otherSide])
		if numKingAttacks != 0 {
			evalInfo.numKingAttackers[side]++
			evalInfo.numKingAttacks[side] += numKingAttacks
			evalInfo.kingAttackWeight[side] += 2 * numKingAttacks
		}

		knightColumn := Column(sq)

		// no points for kamikaze into a pawn so you trade for a bishop
		if !IsBitboardSet(pawnEntry.attackBoard[otherSide], sq) {
			knightRank := Rank(sq)
			// Rank is between 1 and 8.  KNIGHT_RANK_SCORE is 0-indexed (0 -> 7).
			if side == WHITE_OFFSET {
				score += KNIGHT_RANK_SCORE[knightRank-1]
			} else {
				score += KNIGHT_RANK_SCORE[8-knightRank]
			}
		}
		score += KNIGHT_COLUMN_SCORE[knightColumn-1]
		if IsBitboardSet(pawnEntry.holesBoard[otherSide], sq) {
			score += KNIGHT_ON_ENEMY_HOLE_SCORE
		} else if pawnEntry.holesBoard[otherSide]&attackBoard != 0 {
			score += KNIGHT_ATTACK_ENEMY_HOLE_SCORE
		}

		score += KNIGHT_PAWN_ADJUST[bits.OnesCount64(pawnEntry.pawns[side])]
	case ROOK_MASK:
		eval.gamePhase += 2
		rookKey := hashKey(allOccupancies, moveBitboards.rookMagics[sq])
		attackBoard = moveBitboards.rookAttacks[sq][rookKey].board

		// Attacking near enemy king
		numKingAttacks := bits.OnesCount64(attackBoard & evalBitboards.kingSquares[otherSide])
		if numKingAttacks != 0 {
			evalInfo.numKingAttackers[side]++
			evalInfo.numKingAttacks[side] += numKingAttacks
			evalInfo.kingAttackWeight[side] += 3 * numKingAttacks
		}

		// Doubled rooks (will be taken into account for both rooks)
		sideRooks := boardState.bitboards.piece[ROOK_MASK] & boardState.bitboards.piece[side]
		if attackBoard&sideRooks != 0 {
			// don't bother with more complicated computations if you have > 2 rooks, this
			// will never happen at a point where it matters
			otherRookSq := byte(bits.TrailingZeros64(sideRooks))
			if Column(sq) == Column(otherRookSq) {
				score += ROOK_DOUBLED_SCORE
			}
		}

		// Blocked pawn
		numBlockedPawns := bits.OnesCount64(attackBoard & evalBitboards.blockedPawns[side])
		score += int(numBlockedPawns) * PIECE_BEHIND_BLOCKED_PAWN_SCORE[ROOK_MASK]

		// Bonus on open file
		if IsBitboardSet(pawnEntry.openFileBoard, sq) {
			score += ROOK_OPEN_FILE_SCORE
		} else if IsBitboardSet(pawnEntry.halfOpenFileBoard[otherSide], sq) {
			score += ROOK_HALF_OPEN_FILE_SCORE
		}

		// Bonus on 6th/7th rank
		rookRank := Rank(sq)
		if side == WHITE_OFFSET {
			score += ROOK_RANK_SCORE[rookRank-1]
		} else {
			score += ROOK_RANK_SCORE[8-rookRank]
		}

		score += ROOK_PAWN_ADJUST[bits.OnesCount64(pawnEntry.pawns[side])]

		// Passed pawn support
		passedPawnSupportBoard := attackBoard & pawnEntry.passedPawns[side]
		if passedPawnSupportBoard != 0 {
			rookCol := Column(sq)
			for passedPawnSupportBoard != 0 {
				pawnSq := bits.TrailingZeros64(passedPawnSupportBoard)
				pawnRank := Rank(uint8(pawnSq))
				var isBehind = rookRank < pawnRank
				if side == BLACK_OFFSET {
					isBehind = !isBehind
				}
				if Column(uint8(pawnSq)) == rookCol && isBehind {
					score += ROOK_SUPPORT_PASSED_PAWN_SCORE
				}
				passedPawnSupportBoard ^= 1 << pawnSq
			}
		}

	case QUEEN_MASK:
		eval.gamePhase += 4
		rookKey := hashKey(allOccupancies, moveBitboards.rookMagics[sq])
		bishopKey := hashKey(allOccupancies, moveBitboards.bishopMagics[sq])
		attackBoard = moveBitboards.rookAttacks[sq][rookKey].board | moveBitboards.bishopAttacks[sq][bishopKey].board
		numKingAttacks := bits.OnesCount64(attackBoard & evalBitboards.kingSquares[otherSide])
		if numKingAttacks != 0 {
			evalInfo.numKingAttackers[side]++
			evalInfo.numKingAttacks[side] += numKingAttacks
			evalInfo.kingAttackWeight[side] += 4 * numKingAttacks
		}

	case KING_MASK:
		attackBoard = moveBitboards.kingAttacks[sq].board
	}

	if attackBoard != 0 {
		// Don't count mobility as attacking your own piece
		mobilityBoard := attackBoard ^ (attackBoard & boardState.bitboards.color[side])
		for j := 0; j <= 1; j++ {
			numMoves := bits.OnesCount64(mobilityBoard)
			score := MOBILITY_SCORE[j][pieceMask][numMoves]
			eval.mobility[j][side] += score
		}
	}

	eval.pieceScore[side][pieceMask] += score
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
	middleGameMaterial := eval.material[0][WHITE_OFFSET] - eval.material[0][BLACK_OFFSET]
	endgameMaterial := eval.material[1][WHITE_OFFSET] - eval.material[1][BLACK_OFFSET]
	// We want to combine the middlegame material with the endgame material.
	// This uses a method from the https://www.chessprogramming.org/CPW-Engine_eval
	var score int
	midgameWeight := Min(24, eval.gamePhase)
	endgameWeight := 24 - midgameWeight

	score += (midgameWeight*middleGameMaterial + endgameWeight*endgameMaterial) / 24

	middleGameMobility := eval.mobility[0][WHITE_OFFSET] - eval.mobility[0][BLACK_OFFSET]
	endgameMobility := eval.mobility[1][WHITE_OFFSET] - eval.mobility[1][BLACK_OFFSET]

	score += (midgameWeight*middleGameMobility + endgameWeight*endgameMobility) / 24
	score += eval.centerControl
	score += totalPieceScore
	score += eval.developmentScore
	score += eval.kingAttackScore

	if eval.sideToMove == BLACK_OFFSET {
		return -score - TEMPO_SCORE
	}
	return score + TEMPO_SCORE
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
	tempoScore := TEMPO_SCORE
	if eval.sideToMove == BLACK_OFFSET {
		tempoScore = -tempoScore
	}

	return fmt.Sprintf("VALUE: %d\n\tphase=%s\n\tmiddlegame material=%d (white: %d, black: %d)\n\tendgame material=%d (white: %d, black: %d)\n\tmiddlegame mobility=%d (white: %d, black: %d)\n\tendgame mobility=%d (white: %d, black: %d)\n\tpieces=%d [%s]\n\tdevelopment=%d (white: %d, black: %d)\n\tcenterControl=%d\n\tattackScore=%d\n\ttempo=%d",
		eval.value(),
		phaseString,
		eval.material[0][WHITE_OFFSET]-eval.material[0][BLACK_OFFSET],
		eval.material[0][WHITE_OFFSET],
		eval.material[0][BLACK_OFFSET],
		eval.material[1][WHITE_OFFSET]-eval.material[1][BLACK_OFFSET],
		eval.material[1][WHITE_OFFSET],
		eval.material[1][BLACK_OFFSET],
		eval.mobility[0][WHITE_OFFSET]-eval.mobility[0][BLACK_OFFSET],
		eval.mobility[0][WHITE_OFFSET],
		eval.mobility[0][BLACK_OFFSET],
		eval.mobility[1][WHITE_OFFSET]-eval.mobility[1][BLACK_OFFSET],
		eval.mobility[1][WHITE_OFFSET],
		eval.mobility[1][BLACK_OFFSET],
		totalPieceScore,
		strings.Trim(pieces, " "),
		eval.developmentScore,
		eval.whiteDevelopment,
		eval.blackDevelopment,
		eval.centerControl,
		eval.kingAttackScore,
		tempoScore,
	)
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

func checkForInsufficientMatingMaterial(
	eval *BoardEval,
	whitePieceBoard uint64,
	blackPieceBoard uint64,
) {
	if !IsBitboardSet(whitePieceBoard, PAWN_MASK) &&
		!IsBitboardSet(blackPieceBoard, PAWN_MASK) &&
		!IsBitboardSet(whitePieceBoard, QUEEN_MASK) &&
		!IsBitboardSet(blackPieceBoard, QUEEN_MASK) &&
		!IsBitboardSet(whitePieceBoard, ROOK_MASK) &&
		!IsBitboardSet(blackPieceBoard, ROOK_MASK) {
		whiteBishops := eval.pieceCount[WHITE_OFFSET][BISHOP_MASK]
		blackBishops := eval.pieceCount[BLACK_OFFSET][BISHOP_MASK]
		whiteKnights := eval.pieceCount[WHITE_OFFSET][KNIGHT_MASK]
		blackKnights := eval.pieceCount[BLACK_OFFSET][KNIGHT_MASK]

		whiteCanMate := true
		blackCanMate := true
		if whiteBishops == 1 && whiteKnights == 0 {
			whiteCanMate = false
		} else if whiteKnights == 1 && whiteBishops == 0 {
			whiteCanMate = false
		} else if whiteKnights == 2 {
			whiteCanMate = false
		}

		if blackBishops == 1 && blackKnights == 0 {
			blackCanMate = false
		} else if blackKnights == 1 && blackBishops == 0 {
			blackCanMate = false
		} else if blackKnights == 2 {
			blackCanMate = false
		}

		eval.hasMatingMaterial = !whiteCanMate && !blackCanMate
	}
}
