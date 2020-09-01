package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var CHECKMATE_FLAG byte = 0x80
var DRAW_FLAG byte = 0x40
var CHECK_FLAG byte = 0x20
var THREEFOLD_REP_FLAG byte = 0x10

const MAX_DEPTH uint = 32

type SearchStats struct {
	leafnodes         uint64
	branchnodes       uint64
	qbranchnodes      uint64
	hashcutoffs       uint64
	killercutoffs     uint64
	killer2cutoffs    uint64
	cutoffs           uint64
	qcutoffs          uint64
	qcapturesfiltered uint64
	tthits            uint64
}

type SearchResult struct {
	move  Move
	value int
	flags byte
	time  time.Duration
	depth uint
	stats SearchStats
	pv    string
}

type ThinkingOutput struct {
	ply   uint
	score string
	time  int64
	nodes uint64
	pv    string
}

func (result *SearchResult) IsCheckmate() bool {
	return result.flags&CHECKMATE_FLAG == CHECKMATE_FLAG
}

func (result *SearchResult) IsDraw() bool {
	return result.flags&DRAW_FLAG == DRAW_FLAG
}

type SearchConfig struct {
	move          Move
	isDebug       bool
	debugMoves    string
	startingDepth uint
	startTime     time.Time
}

type ExternalSearchConfig struct {
	isDebug       bool
	debugMoves    string
	searchToDepth uint
}

type SearchMoveInfo struct {
	killerMoves    [MAX_DEPTH]Move
	killerMoves2   [MAX_DEPTH]Move
	firstPlyScores [MAX_MOVES]int
}

func Search(
	boardState *BoardState,
	depth uint,
	stats *SearchStats,
	moveInfo *SearchMoveInfo,
) SearchResult {
	return SearchWithConfig(boardState, depth, stats, moveInfo, ExternalSearchConfig{}, nil)
}

func SearchWithConfig(
	boardState *BoardState,
	depth uint,
	stats *SearchStats,
	moveInfo *SearchMoveInfo,
	config ExternalSearchConfig,
	thinkingChan chan ThinkingOutput,
) SearchResult {
	startTime := time.Now()

	alpha := -INFINITY
	beta := INFINITY
	searchConfig := SearchConfig{
		isDebug:       config.isDebug,
		debugMoves:    config.debugMoves,
		startingDepth: depth,
		startTime:     startTime,
	}
	moves := make([]Move, 64*256)
	scores := make([]int, len(moves))

	var moveStart [64]int
	score := searchAlphaBeta(boardState, stats, moveInfo,
		thinkingChan,
		int(depth),
		0,
		alpha,
		beta,
		searchConfig,
		moves[:],
		scores[:],
		moveStart[:],
	)

	result := SearchResult{}

	pv, isDraw := extractPV(boardState)
	if isDraw {
		result.flags = DRAW_FLAG
	}
	if boardState.sideToMove == BLACK_OFFSET {
		score = -score
	}
	absScore := score
	if score < 0 {
		absScore = -score
	}
	if absScore > CHECKMATE_SCORE-100 {
		result.flags = CHECKMATE_FLAG
	}
	result.value = score
	result.time = time.Now().Sub(startTime)
	// TODO - guard against null
	result.move = pv[0]
	result.pv, _ = MoveArrayToPrettyString(pv, boardState)
	result.stats = *stats
	result.depth = depth

	if shouldAbort {
		close(thinkingChan)
	}

	return result
}

// searchAlphaBeta runs an alpha-beta search over the boardState
//
// - depth, when positive, is the number of levels until quiescent search begins.
//   when negative it is the number of levels deep in quiescent search.
// - currentDepth is the total depth of the search tree, including quiescent levels.
//  - alpha: minimum score that player to move can achieve
//  - black: maximum score that opponent can achieve
//
func searchAlphaBeta(
	boardState *BoardState,
	searchStats *SearchStats,
	moveInfo *SearchMoveInfo,
	thinkingChan chan ThinkingOutput,
	depthLeft int,
	currentDepth uint,
	alpha int,
	beta int,
	searchConfig SearchConfig,
	moves []Move,
	moveScores []int,
	moveStart []int,
) int {
	var hashMove Move
	var hasHashMove bool

	isDebug := searchConfig.isDebug

	if entry := ProbeTranspositionTable(boardState); entry != nil {
		if currentDepth > 0 {
			if entry.depth >= depthLeft {
				switch entry.entryType {
				case TT_EXACT:
					searchStats.tthits++
					return entry.score
				case TT_FAIL_HIGH:
					if entry.score >= beta {
						searchStats.tthits++
						return entry.score
					}
				case TT_FAIL_LOW:
					if entry.score <= alpha {
						searchStats.tthits++
						return entry.score
					}
				}
			}
		}
		move := entry.move
		if _, err := boardState.IsMoveLegal(move); err == nil {
			hashMove = move
			hasHashMove = true
		}
	}

	if shouldAbort || currentDepth >= MAX_DEPTH {
		score := getLeafResult(boardState, searchStats)
		StoreTranspositionTable(boardState, Move{}, score, TT_EXACT, depthLeft)

		return score
	}

	inCheck := boardState.IsInCheck(boardState.sideToMove)
	if inCheck {
		depthLeft++
	}

	if depthLeft == 0 {
		return searchQuiescent(boardState, searchStats, depthLeft, currentDepth, alpha, beta, searchConfig, moves, moveScores, moveStart)
	}

	if boardState.HasStateOccurred() {
		// TODO - contempt value
		return 0
	}

	searchStats.branchnodes++

	// We'll generate the other moves after we test the hash move
	// 0 = hash
	// 1 = captures
	// 2 = promotions
	// 3 = checks
	// 4 = moves

	hasLegalMove := false
	bestScore := -INFINITY + 1
	var bestMove Move
	currentAlpha := alpha

	start := moveStart[currentDepth]
	var moveEnd int
	if hasHashMove {
		moves[start] = hashMove
		moveStart[currentDepth+1] = start + 1
		moveEnd = start + 1
	} else {
		moveStart[currentDepth+1] = start
		moveEnd = start
	}

	for i := 0; i <= 1; i++ {
		for j := start; j < moveEnd; j++ {
			move := moves[j]
			if move.IsCastle() && !boardState.TestCastleLegality(move) {
				continue
			}

			debugMode := isDebug && (strings.Contains(MoveToPrettyString(move, boardState), searchConfig.debugMoves) ||
				searchConfig.debugMoves == "*")

			offset := boardState.sideToMove
			boardState.ApplyMove(move)
			if boardState.IsInCheck(offset) {
				boardState.UnapplyMove(move)
				continue
			}
			hasLegalMove = true
			searchConfig.move = move
			var nodesStarting uint64
			var hashKey uint64

			if debugMode {
				nodesStarting = searchStats.Nodes()
				hashKey = boardState.hashKey
			}
			searchConfig.isDebug = false

			D := 0
			if IsPawnNearPromotion(boardState, move) {
				D = 1
			}

			score := -searchAlphaBeta(boardState,
				searchStats,
				moveInfo,
				thinkingChan,
				depthLeft-1+D,
				currentDepth+1,
				-beta, -currentAlpha, // swap alpha and beta
				searchConfig,
				moves,
				moveScores,
				moveStart,
			)

			boardState.UnapplyMove(move)

			if currentDepth == 0 && i == 1 {
				moveInfo.firstPlyScores[j] = score
			}

			if debugMode {
				pv, _ := extractPV(boardState)
				str := MoveArrayToXboardString(pv)
				entry := boardState.transpositionTable[hashKey]
				var entryType string
				if entry != nil {
					entryType = EntryTypeToString(entry.entryType)
				}
				fmt.Printf("[%d; %s] value=%d (alpha=%d, beta=%d) nodes=%d pv=%s result=%s\n",
					depthLeft, MoveToString(move, boardState), score, currentAlpha, beta, searchStats.Nodes()-nodesStarting, str,
					entryType)
			}

			if score >= beta {
				lastKiller := moveInfo.killerMoves[currentDepth]
				lastKiller2 := moveInfo.killerMoves2[currentDepth]
				if move == lastKiller {
					// nothing, we good
				} else {
					moveInfo.killerMoves[currentDepth] = move
					moveInfo.killerMoves2[currentDepth] = lastKiller
				}
				StoreTranspositionTable(boardState, move, score, TT_FAIL_HIGH, depthLeft)
				searchStats.cutoffs++
				if i == 0 {
					searchStats.hashcutoffs++
				} else if move == lastKiller {
					searchStats.killercutoffs++
				} else if move == lastKiller2 {
					searchStats.killer2cutoffs++
				}
				return score
			}

			if score > bestScore {
				bestScore = score
				bestMove = move
				if bestScore > alpha {
					currentAlpha = score
					if currentDepth == 0 && thinkingChan != nil {
						sendToThinkingChannel(boardState, searchStats, thinkingChan, searchConfig, bestScore, depthLeft)
					}
				}
			}
		}

		if i == 0 {
			// add the other moves now that we're done with hash move
			moveEnd = GenerateMoves(boardState, moves, start)
			moveStart[currentDepth+1] = moveEnd
			if currentDepth > 0 {
				SortMoves(boardState, moveInfo, currentDepth, moves, moveScores, start, moveEnd)
			} else {
				SortMovesFirstPly(boardState, moveInfo, moves, start, moveEnd)
				if isDebug {
					fmt.Printf("[%d] Move ordering: %s\nScores: %v\n",
						depthLeft,
						MoveArrayToXboardString(moves[start:moveEnd]),
						moveInfo.firstPlyScores[start:moveEnd])
				}
			}
		}
	}

	// IF WE HAD NO LEGAL MOVES, GAME IS OVER

	if !hasLegalMove {
		score := getNoLegalMoveResult(boardState, currentDepth)
		StoreTranspositionTable(boardState, Move{}, score, TT_EXACT, depthLeft)

		return score
	}

	var ttEntryType int
	if currentAlpha == alpha {
		// never raised alpha
		ttEntryType = TT_FAIL_LOW
	} else {
		// we raised alpha, so we have an exact match
		ttEntryType = TT_EXACT
	}

	StoreTranspositionTable(boardState, bestMove, bestScore, ttEntryType, depthLeft)

	return bestScore
}

func searchQuiescent(
	boardState *BoardState,
	searchStats *SearchStats,
	// depthLeft will always be negative
	depthLeft int,
	currentDepth uint,
	alpha int,
	beta int,
	searchConfig SearchConfig,
	moves []Move,
	moveScores []int,
	moveStart []int,
) int {
	// var bestMove Move
	bestScore := -INFINITY + 1

	// Evaluate the board to see what the position is without making any quiescent moves.
	score := getLeafResult(boardState, searchStats)
	if score >= beta {
		return beta
	}
	if score >= alpha {
		bestScore = score
		alpha = score
	}
	if shouldAbort {
		return score
	}

	start := moveStart[currentDepth]
	endAllMoves := GenerateQuiescentMoves(boardState, moves, moveScores, start)
	endGoodMoves := boardState.FilterSEECaptures(moves, start, endAllMoves)
	moveStart[currentDepth+1] = endGoodMoves

	searchStats.qcapturesfiltered += uint64(endAllMoves - endGoodMoves)
	searchStats.qbranchnodes++

	for i := start; i < endGoodMoves; i++ {
		move := moves[i]
		ourOffset := boardState.sideToMove
		boardState.ApplyMove(move)
		if boardState.IsInCheck(ourOffset) {
			boardState.UnapplyMove(move)
			continue
		}

		score := -searchQuiescent(boardState, searchStats, depthLeft-1, currentDepth+1, -beta, -alpha, searchConfig, moves, moveScores, moveStart)
		boardState.UnapplyMove(move)

		if score >= beta {
			searchStats.cutoffs++
			searchStats.qcutoffs++

			return beta
		}

		if score > bestScore {
			bestScore = score
			if score > alpha {
				alpha = score
			}
		}
	}

	return bestScore
}

func getLeafResult(boardState *BoardState, searchStats *SearchStats) int {
	// TODO(perf): use an incremental evaluation state passed in as an argument
	searchStats.leafnodes++

	e := Eval(boardState)
	if !e.hasMatingMaterial {
		return 0
	}

	return e.value()
}

func getNoLegalMoveResult(boardState *BoardState, currentDepth uint) int {
	if boardState.IsInCheck(boardState.sideToMove) {
		// moves to mate = currentDepth
		return -(CHECKMATE_SCORE - int(currentDepth) + 1)
	}

	// Stalemate
	return 0

}

func (result *SearchResult) String() string {
	return fmt.Sprintf("%s (time=%s, value=%s, depth=%d, stats=%s, pv=%s)",
		MoveToXboardString(result.move),
		result.time.String(),
		SearchValueToString(*result),
		result.depth,
		result.stats.String(),
		result.pv)
}

func SearchValueToString(result SearchResult) string {
	if result.IsCheckmate() {
		score := result.value
		if score < 0 {
			score = -score
		}
		movesToCheckmate := (CHECKMATE_SCORE - score + 1) / 2
		return fmt.Sprintf("\033[1;34mMate(%d)\033[0m", movesToCheckmate)
	}

	if result.IsDraw() {
		return fmt.Sprintf("\033[1;33mDraw\033[0m")
	}

	return strconv.Itoa(result.value)
}

func (stats *SearchStats) String() string {
	return fmt.Sprintf("[nodes=%d, leafnodes=%d, branchnodes=%d, qbranchnodes=%d, tthits=%d, cutoffs=%d, hash cutoffs=%d, killer cutoffs={1: %d, 2: %d}, qcutoffs=%d, qcapturesfiltered=%d]",
		stats.Nodes(),
		stats.leafnodes,
		stats.branchnodes,
		stats.qbranchnodes,
		stats.tthits,
		stats.cutoffs,
		stats.hashcutoffs,
		stats.killercutoffs,
		stats.killer2cutoffs,
		stats.qcutoffs,
		stats.qcapturesfiltered)
}

func (stats *SearchStats) Nodes() uint64 {
	return stats.branchnodes + stats.leafnodes + stats.qbranchnodes
}

// Used to determine if we should extend search
func IsPawnNearPromotion(boardState *BoardState, m Move) bool {
	movePiece := boardState.board[m.to]
	if movePiece&0x0F != PAWN_MASK {
		return false
	}

	rank := Rank(m.to)
	return (movePiece == WHITE_MASK|PAWN_MASK && rank == 7) || (rank == 2)
}

func sendToThinkingChannel(
	boardState *BoardState,
	searchStats *SearchStats,
	thinkingChan chan ThinkingOutput,
	searchConfig SearchConfig,
	score int,
	depthLeft int,
) {
	pvMoves, _ := extractPV(boardState)
	pv, _ := MoveArrayToPrettyString(pvMoves, boardState)
	timeNanos := time.Now().Sub(searchConfig.startTime).Nanoseconds()
	var scoreString string
	absScore := score
	if score < 0 {
		absScore = -score
	}
	if absScore > CHECKMATE_SCORE-100 {
		var prefix string
		if score < 0 {
			prefix = "-"
		} else {
			prefix = ""
		}
		scoreString = fmt.Sprintf("%sMate%d", prefix, (CHECKMATE_SCORE-absScore+1)/2)
	} else {
		scoreString = strconv.Itoa(score)
	}

	thinkingChan <- ThinkingOutput{
		ply:   uint(depthLeft),
		score: scoreString,
		time:  int64(float32(timeNanos) * 1e-7),
		nodes: searchStats.Nodes(),
		pv:    pv,
	}
}

// extractPV will return the move list for a given position from the transposition table.
func extractPV(boardState *BoardState) ([]Move, bool) {
	isDraw := false
	pvMoves := make([]Move, 0)
	e := ProbeTranspositionTable(boardState)
	lastDepth := INFINITY
	for e != nil && e.depth <= lastDepth {
		move := e.move
		if _, err := boardState.IsMoveLegal(move); err != nil {
			break
		}
		pvMoves = append(pvMoves, move)
		boardState.ApplyMove(move)
		// Avoid repetitions in moves
		if boardState.HasStateOccurred() {
			break
		}
		if !Eval(boardState).hasMatingMaterial {
			isDraw = true
			break
		}
		lastDepth = e.depth
		e = ProbeTranspositionTable(boardState)
	}
	for j := len(pvMoves) - 1; j >= 0; j-- {
		move := pvMoves[j]
		boardState.UnapplyMove(move)
	}

	return pvMoves, isDraw
}
