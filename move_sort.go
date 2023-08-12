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
	moveScores []int
	startIndex int
	endIndex   int
}

func (pq MoveSort) Len() int {
	return pq.endIndex - pq.startIndex
}

func (pq MoveSort) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	idx := pq.startIndex
	return pq.moveScores[idx+i] > pq.moveScores[idx+j]
}

func (pq MoveSort) Swap(i, j int) {
	idx := pq.startIndex
	pq.moves[idx+i], pq.moves[idx+j] = pq.moves[idx+j], pq.moves[idx+i]
	pq.moveScores[idx+i], pq.moveScores[idx+j] = pq.moveScores[idx+j], pq.moveScores[idx+i]
}

var moveSort MoveSort

func SortMoves(
	boardState *BoardState,
	moveInfo *SearchMoveInfo,
	currentDepth uint,
	moves []Move,
	moveScores []int,
	start int,
	end int,
) {
	checkDetectionInfo := makeCheckDetectionInfo(boardState)

	for i := start; i < end; i++ {
		move := moves[i]
		var score int = MOVE_SCORE_NORMAL
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

	moveSort.startIndex = start
	moveSort.endIndex = end
	moveSort.moves = moves
	moveSort.moveScores = moveScores
	sort.Sort(&moveSort)
}

func SortQuiescentMoves(boardState *BoardState, moves []Move, moveScores []int, start int, end int) {
	moveSort.startIndex = start
	moveSort.endIndex = end
	moveSort.moves = moves
	moveSort.moveScores = moveScores

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
	moveSort.startIndex = start
	moveSort.endIndex = end
	moveSort.moves = moves
	moveSort.moveScores = moveInfo.firstPlyScores[start:end]
	sort.Sort(&moveSort)
}
