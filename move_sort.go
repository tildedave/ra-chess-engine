package main

import (
	"sort"
)

// from https://golang.org/pkg/container/heap/

const MOVE_SCORE_NORMAL = 100
const MOVE_SCORE_CHECKS = 200
const MOVE_SCORE_CAPTURES = 300
const MOVE_SCORE_PROMOTIONS = 400
const MOVE_SCORE_KILLER_MOVE = 500
const MOVE_SCORE_KILLER_MOVE2 = 450

type MoveSort struct {
	moves      []Move
	moveScores []int16
}

func (s MoveSort) Len() int {
	return len(s.moves)
}

func (s MoveSort) Less(i, j int) bool {
	return s.moveScores[i] > s.moveScores[j]
}

func (s MoveSort) Swap(i, j int) {
	s.moves[i], s.moves[j] = s.moves[j], s.moves[i]
	s.moveScores[i], s.moveScores[j] = s.moveScores[j], s.moveScores[i]
}

var moveSort MoveSort

func SortMoves(
	boardState *BoardState,
	moveInfo *SearchMoveInfo,
	currentDepth uint,
	moves []Move,
	moveScores []int16,
	start int,
	end int,
) {
	checkDetectionInfo := makeCheckDetectionInfo(boardState)

	for i := start; i < end; i++ {
		move := moves[i]
		var score int16 = MOVE_SCORE_NORMAL
		toPiece := boardState.PieceAtSquare(move.to)
		if move == moveInfo.killerMoves[currentDepth] {
			score = MOVE_SCORE_KILLER_MOVE
		} else if move == moveInfo.killerMoves2[currentDepth] {
			score = MOVE_SCORE_KILLER_MOVE2
		} else if move.flags&(PROMOTION_MASK|QUEEN_MASK) == PROMOTION_MASK|QUEEN_MASK {
			// don't bother with scoring underpromotions higher
			// we probably could avoid the & here since a pawn capture will be ranked
			// somewhat higher because of MVV priority
			score = MOVE_SCORE_PROMOTIONS
		} else if toPiece != EMPTY_SQUARE {
			fromPiece := boardState.PieceAtSquare(move.from)
			priority := mvvPriority[fromPiece&0x0F][toPiece&0x0F]
			score = MOVE_SCORE_CAPTURES + priority
		} else if boardState.IsMoveCheck(move, &checkDetectionInfo) {
			// TODO - check detection
		}
		moveScores[i] = score
	}

	moveSort.moves = moves[start:end]
	moveSort.moveScores = moveScores[start:end]
	sort.Sort(&moveSort)
}

func SortQuiescentMoves(boardState *BoardState, moves []Move, moveScores []int16, start int, end int) {
	moveSort.moves = moves[start:end]
	moveSort.moveScores = moveScores[start:end]

	for i := start; i < end; i++ {
		capture := moves[i]
		fromPiece := boardState.PieceAtSquare(capture.from)
		toPiece := boardState.PieceAtSquare(capture.to)
		priority := mvvPriority[fromPiece&0x0F][toPiece&0x0F]
		moveScores[i] = priority
	}
	sort.Sort(&moveSort)
}

func SortMovesFirstPly(
	boardState *BoardState,
	moveInfo *SearchMoveInfo,
	moves []Move,
	start int,
	end int,
) {
	moveSort.moves = moves[start:end]
	moveSort.moveScores = moveInfo.firstPlyScores[start:end]
	sort.Sort(&moveSort)
}
