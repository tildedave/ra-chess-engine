package main

import (
	"container/heap"
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

const MOVE_HASH_MOVE = 0
const MOVE_CAPTURES = 1
const MOVE_PROMOTIONS = 2
const MOVE_KILLERS = 3
const MOVE_CHECKS = 4
const MOVE_NORMAL = 5

const MOVE_SCORE_NORMAL = 100
const MOVE_SCORE_CHECKS = 200
const MOVE_SCORE_CAPTURES = 300
const MOVE_SCORE_PROMOTIONS = 400
const MOVE_SCORE_KILLER_MOVE = 500

type SearchStats struct {
	leafnodes         uint64
	branchnodes       uint64
	qbranchnodes      uint64
	hashcutoffs       uint64
	killercutoffs     uint64
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

type Variation struct {
	move     [MAX_MOVES]Move
	numMoves uint
}

func CopyVariation(line *Variation, move Move, restLine *Variation) {
	line.move[0] = move
	copy(line.move[1:], restLine.move[0:restLine.numMoves])
	line.numMoves = restLine.numMoves + 1
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
	// eventually add a second array here
	killerMoves [MAX_DEPTH]Move
}

func Search(boardState *BoardState, depth uint) SearchResult {
	return SearchWithConfig(boardState, depth, ExternalSearchConfig{}, nil)
}

func SearchWithConfig(
	boardState *BoardState,
	depth uint,
	config ExternalSearchConfig,
	thinkingChan chan ThinkingOutput,
) SearchResult {
	startTime := time.Now()

	stats := SearchStats{}
	variation := Variation{}
	moveInfo := SearchMoveInfo{}

	alpha := -INFINITY
	beta := INFINITY
	searchConfig := SearchConfig{
		isDebug:       config.isDebug,
		debugMoves:    config.debugMoves,
		startingDepth: depth,
		startTime:     startTime,
	}
	moves := make([]Move, 64*256)
	var moveStart [64]int
	score := searchAlphaBeta(boardState, &stats, &variation, &moveInfo,
		thinkingChan,
		int(depth),
		0,
		alpha,
		beta,
		searchConfig,
		&moves,
		&moveStart,
	)

	result := SearchResult{}
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
	result.move = variation.move[0]
	result.pv, _ = MoveArrayToPrettyString(variation.move[0:variation.numMoves], boardState)
	result.stats = stats
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
	variation *Variation,
	moveInfo *SearchMoveInfo,
	thinkingChan chan ThinkingOutput,
	depthLeft int,
	currentDepth uint,
	alpha int,
	beta int,
	searchConfig SearchConfig,
	moves *[]Move,
	moveStart *[64]int,
) int {
	var line Variation

	isDebug := searchConfig.isDebug

	var hashMove = make([]Move, 0)
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
						return beta
					}
				case TT_FAIL_LOW:
					if entry.score <= alpha {
						searchStats.tthits++
						return alpha
					}
				}
			}
		}
		move := entry.move
		if _, err := boardState.IsMoveLegal(move); err == nil {
			hashMove = append(hashMove, move)
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
		score := searchQuiescent(boardState, searchStats, &line, depthLeft, currentDepth, alpha, beta, searchConfig, moves, moveStart)
		if score > alpha {
			copy(variation.move[0:], line.move[0:line.numMoves])
			variation.numMoves = line.numMoves
		}
		return score
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

	start := (*moveStart)[currentDepth]
	var moveEnd int
	(*moveStart)[currentDepth+1] = start

	for i := 0; i <= 1; i++ {
		var moveOrdering []Move
		if i == 0 {
			moveOrdering = hashMove
		} else {
			// TODO - don't want to do this, we should index
			// Fix this by putting the hashMove onto the move list in this scenario
			moveOrdering = (*moves)[start:moveEnd]
		}

		for _, move := range moveOrdering {
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

			score := -searchAlphaBeta(boardState, searchStats, &line, moveInfo, thinkingChan,
				depthLeft-1, currentDepth+1,
				-beta, -currentAlpha, // swap alpha and beta
				searchConfig,
				moves,
				moveStart,
			)

			boardState.UnapplyMove(move)

			if debugMode {
				var moveLine Variation
				CopyVariation(&moveLine, move, &line)
				str := MoveArrayToXboardString(moveLine.move[0:moveLine.numMoves])
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
				moveInfo.killerMoves[currentDepth] = move
				StoreTranspositionTable(boardState, move, score, TT_FAIL_HIGH, depthLeft)
				searchStats.cutoffs++
				if i == 0 {
					searchStats.hashcutoffs++
				} else if i == 3 {
					searchStats.killercutoffs++
				}
				return score
			}

			if score > bestScore {
				bestScore = score
				bestMove = move
				if bestScore > alpha {
					currentAlpha = score
					CopyVariation(variation, move, &line)
					if currentDepth == 0 && thinkingChan != nil {
						sendToThinkingChannel(boardState, searchStats, variation, thinkingChan, searchConfig, bestScore, depthLeft)
					}
				}
			}
		}

		if i == 0 {
			// add the other moves now that we're done with hash move
			moveEnd = GenerateMoves(boardState, moves, start)
			(*moveStart)[currentDepth+1] = moveEnd
			sortMoves(boardState, moveInfo, currentDepth, moves, start, moveEnd)
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

func sortMoves(
	boardState *BoardState,
	moveInfo *SearchMoveInfo,
	currentDepth uint,
	moves *[]Move,
	start int,
	end int,
) {
	priorityQueue := make(MovePriorityQueue, 0, end-start)

	for i := start; i < end; i++ {
		move := (*moves)[i]
		var score int = MOVE_SCORE_NORMAL
		toPiece := boardState.PieceAtSquare(move.to)
		if move == moveInfo.killerMoves[currentDepth] {
			score = MOVE_SCORE_KILLER_MOVE
		} else if move.flags&PROMOTION_MASK|QUEEN_MASK == PROMOTION_MASK|QUEEN_MASK {
			// don't bother with scoring underpromotions higher
			// we probably could avoid the & here since a pawn capture will be ranked somewhat higher
			score = MOVE_SCORE_PROMOTIONS
		} else if toPiece != EMPTY_SQUARE {
			fromPiece := boardState.PieceAtSquare(move.from)
			priority := mvvPriority[fromPiece&0x0F][toPiece&0x0F]
			score = MOVE_SCORE_CAPTURES + priority
		} else {
			// TODO - check detection
		}
		item := Item{move: move, score: score}
		priorityQueue.Push(&item)
	}
	heap.Init(&priorityQueue)

	current := start
	for priorityQueue.Len() > 0 {
		item := heap.Pop(&priorityQueue).(*Item)
		(*moves)[current] = item.move
		current++
	}
}

func searchQuiescent(
	boardState *BoardState,
	searchStats *SearchStats,
	variation *Variation,
	// depthLeft will always be negative
	depthLeft int,
	currentDepth uint,
	alpha int,
	beta int,
	searchConfig SearchConfig,
	moves *[]Move,
	moveStart *[64]int,
) int {
	var line Variation
	var bestMove Move
	bestScore := -INFINITY + 1

	// Evaluate the board to see what the position is without making any quiescent moves.
	score := getLeafResult(boardState, searchStats)
	if score >= beta {
		StoreTranspositionTable(boardState, bestMove, bestScore, TT_FAIL_HIGH, depthLeft)
		return beta
	}
	if score >= alpha {
		bestScore = score
		alpha = score
	}
	if shouldAbort {
		return score
	}

	start := (*moveStart)[currentDepth]
	endAllMoves := GenerateQuiescentMoves(boardState, moves, start)
	endGoodMoves := boardState.FilterSEECaptures(moves, start, endAllMoves)
	(*moveStart)[currentDepth+1] = endGoodMoves

	searchStats.qcapturesfiltered += uint64(endAllMoves - endGoodMoves)
	searchStats.qbranchnodes++

	for i := start; i < endGoodMoves; i++ {
		move := (*moves)[i]
		ourOffset := boardState.sideToMove
		boardState.ApplyMove(move)
		if boardState.IsInCheck(ourOffset) {
			boardState.UnapplyMove(move)
			continue
		}

		score := -searchQuiescent(boardState, searchStats, &line, depthLeft-1, currentDepth+1, -beta, -alpha, searchConfig, moves, moveStart)
		boardState.UnapplyMove(move)

		if score >= beta {
			searchStats.cutoffs++
			searchStats.qcutoffs++
			StoreTranspositionTable(boardState, bestMove, score, TT_FAIL_HIGH, depthLeft)

			return beta
		}

		if score > bestScore {
			bestScore = score
			bestMove = move
			if score > alpha {
				alpha = score
				variation.move[0] = move
				copy(variation.move[1:], line.move[0:line.numMoves])
				variation.numMoves = line.numMoves + 1
			}
		}
	}

	var ttEntryType int
	if bestScore < alpha {
		// never raised alpha
		ttEntryType = TT_FAIL_LOW
	} else {
		// we raised alpha, so we have an exact match
		// TODO(2020) - I'm not clear if we should be storing exact entries for Q-search
		// Possibly doesn't matter since depthLeft is negative in this code flow
		ttEntryType = TT_EXACT
	}
	StoreTranspositionTable(boardState, bestMove, bestScore, ttEntryType, depthLeft)

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
		movesToCheckmate := CHECKMATE_SCORE - score + 1
		return fmt.Sprintf("Mate(%d)", movesToCheckmate)
	}

	if result.IsDraw() {
		return fmt.Sprintf("Draw")
	}

	return strconv.Itoa(result.value)
}

func (stats *SearchStats) String() string {
	return fmt.Sprintf("[nodes=%d, leafnodes=%d, branchnodes=%d, qbranchnodes=%d, tthits=%d, cutoffs=%d, hash cutoffs=%d, killer cutoffs=%d, qcutoffs=%d, qcapturesfiltered=%d]",
		stats.Nodes(),
		stats.leafnodes,
		stats.branchnodes,
		stats.qbranchnodes,
		stats.tthits,
		stats.cutoffs,
		stats.hashcutoffs,
		stats.killercutoffs,
		stats.qcutoffs,
		stats.qcapturesfiltered)
}

func (stats *SearchStats) Nodes() uint64 {
	return stats.branchnodes + stats.leafnodes + stats.qbranchnodes
}

// Used to determine if we should extend search
func (m Move) IsQuiescentPawnPush(boardState *BoardState) bool {
	movePiece := boardState.board[m.from]
	if movePiece&0x0F != PAWN_MASK {
		return false
	}

	rank := Rank(m.to)
	return (movePiece == WHITE_MASK|PAWN_MASK && rank == 7) || (rank == 2)
}

func sendToThinkingChannel(
	boardState *BoardState,
	searchStats *SearchStats,
	variation *Variation,
	thinkingChan chan ThinkingOutput,
	searchConfig SearchConfig,
	score int,
	depthLeft int,
) {
	pv, _ := MoveArrayToPrettyString(variation.move[0:variation.numMoves], boardState)
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
