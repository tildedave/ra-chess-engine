package main

import (
	"fmt"
	"time"
)

var _ = fmt.Println

var CHECKMATE_FLAG byte = 0x80
var STALEMATE_FLAG byte = 0x40
var CHECK_FLAG byte = 0x20
var THREEFOLD_REP_FLAG byte = 0x10

type SearchResult struct {
	move        Move
	value       int
	flags       byte
	nodes       uint
	hashCutoffs uint
	cutoffs     uint
	time        int64
	depth       uint
	pv          string
}

type SearchConfig struct {
	alpha         int
	beta          int
	move          Move
	isDebug       bool
	startingDepth uint
}

type ExternalSearchConfig struct {
	isDebug bool
}

func Search(boardState *BoardState, depth uint) SearchResult {
	return SearchWithConfig(boardState, depth, ExternalSearchConfig{})
}

func SearchWithConfig(boardState *BoardState, depth uint, config ExternalSearchConfig) SearchResult {
	startTime := time.Now().UnixNano()

	result := searchAlphaBeta(boardState, depth, 0, SearchConfig{
		alpha:         -INFINITY,
		beta:          INFINITY,
		isDebug:       config.isDebug,
		startingDepth: depth,
	}, MoveSizeHint{})
	result.time = (time.Now().UnixNano() - startTime) / 10000000

	return result
}

func searchAlphaBeta(
	boardState *BoardState,
	depth uint,
	currentDepth uint,
	searchConfig SearchConfig,
	hint MoveSizeHint,
) SearchResult {
	if depth == 0 {
		return getTerminalResult(boardState, searchConfig)
	}

	isDebug := searchConfig.isDebug

	entry := ProbeTranspositionTable(boardState)
	var hashMove = make([]Move, 0)
	if entry != nil {
		if entry.depth >= depth {
			// I guess move might be bogus so we shouldn't just trust this
			return *entry.result
		}

		move := entry.result.move
		// Hash move might be bogus due to hash collision, so we need to validate
		// that it's a legit move (later)
		hashMove = append(hashMove, move)
	}

	var nodes uint
	var hashCutoffs uint
	var cutoffs uint

	listing, hint := GenerateMoveListing(boardState, hint)
	var bestResult *SearchResult

	for i, moveList := range [][]Move{hashMove, listing.promotions, listing.captures, listing.moves} {
		for _, move := range moveList {
			if i == 0 {
				// validate hash move is one of our generated moves
				// TODO: is this the best way?  we could just generate the moves from the square in
				// the hash move prior to generating the full move listing.  feels like this would be
				// effective at pruning the search tree and also save a bunch of time computing the full
				// move listing.
				// NOTE - Apep doesn't generate move list, it uses hash move first and then something similar
				// to IsMoveLegal to sanity check moves.

				isLegal := false
			CheckHashMoveLegality:
				for _, legalMoveListing := range [][]Move{listing.promotions, listing.captures, listing.moves} {
					for _, legalMove := range legalMoveListing {
						if move == legalMove {
							isLegal = true
							break CheckHashMoveLegality
						}
					}
				}

				if !isLegal {
					if _, err := boardState.IsMoveLegal(move); err == nil {
						fmt.Println(boardState.ToString())
						fmt.Println(MoveToPrettyString(move, boardState))
						panic("hash move was not legal but IsMoveLegal says it is")
					}
					break
				}
			}

			if move.IsCastle() && !boardState.TestCastleLegality(move) {
				continue
			}

			searchDepth := depth - 1

			boardState.ApplyMove(move)

			if !boardState.IsInCheck(oppositeColorOffset(boardState.offsetToMove)) {
				searchConfig.move = move
				searchConfig.isDebug = false

				if move.IsCapture() || boardState.IsInCheck(boardState.offsetToMove) || move.IsPromotion() {
					searchDepth++
				}

				result := searchAlphaBeta(boardState, searchDepth, currentDepth+1, searchConfig, hint)

				result.move = move
				result.depth++
				nodes += result.nodes
				hashCutoffs += result.hashCutoffs
				cutoffs += result.cutoffs

				if bestResult == nil {
					bestResult = &result
				} else {
					if (boardState.offsetToMove == BLACK_OFFSET && result.value > bestResult.value) || // white move, maximize score
						(boardState.offsetToMove == WHITE_OFFSET && result.value < bestResult.value) { // black move, minimize score
						bestResult = &result
					}
				}

				if isDebug {
					fmt.Printf("[%d; %s] value=%d, result=%s, pv=%s\n", depth,
						MoveToString(move), result.value, SearchResultToString(result), result.pv)
				}

				if boardState.offsetToMove == BLACK_OFFSET {
					searchConfig.alpha = Max(searchConfig.alpha, bestResult.value)
				} else {
					searchConfig.beta = Min(searchConfig.beta, bestResult.value)
				}
			}

			boardState.UnapplyMove(move)

			if searchConfig.alpha >= searchConfig.beta {
				if i == 0 {
					hashCutoffs++
				}
				cutoffs++
				break
			}
		}
	}

	if bestResult == nil {
		result := getNoLegalMoveResult(boardState, depth, searchConfig)
		bestResult = &result
	} else {
		bestResult.pv = MoveToPrettyString(bestResult.move, boardState) + " " + bestResult.pv
	}

	bestResult.nodes = nodes
	bestResult.hashCutoffs = hashCutoffs
	bestResult.cutoffs = cutoffs
	StoreTranspositionTable(boardState, bestResult, depth)

	return *bestResult
}

func getTerminalResult(boardState *BoardState, searchConfig SearchConfig) SearchResult {
	// TODO(perf): use an incremental evaluation state passed in as an argument

	e, hasMatingMaterial := Eval(boardState)
	if !hasMatingMaterial {
		return SearchResult{
			value: 0,
			flags: STALEMATE_FLAG,
			pv:    "1/2-1/2 (Insufficient mating material)",
		}
	}

	return SearchResult{
		value: e.value(),
		move:  searchConfig.move,
		nodes: 1,
		pv:    "",
	}
}

func getNoLegalMoveResult(boardState *BoardState, depth uint, searchConfig SearchConfig) SearchResult {
	if boardState.IsInCheck(boardState.offsetToMove) {
		// moves to mate = startingDepth - depth
		movesToMate := searchConfig.startingDepth - depth
		score := CHECKMATE_SCORE - int(movesToMate)
		if boardState.offsetToMove == WHITE_OFFSET {
			score = -score
		}

		return SearchResult{
			value: score,
			flags: CHECKMATE_FLAG,
			pv:    "#",
		}
	}

	// Stalemate
	return SearchResult{
		value: 0,
		flags: STALEMATE_FLAG,
		pv:    "1/2-1/2",
	}

}

func SearchResultToString(result SearchResult) string {
	return fmt.Sprintf("%s (value=%d, depth=%d, nodes=%d, cutoffs=%d, hash cutoffs=%d)",
		MoveToString(result.move), result.value, result.depth, result.nodes, result.cutoffs, result.hashCutoffs)
}

// Used to determine if we should extend search
func (m Move) IsQuiescentPawnPush(boardState *BoardState) bool {
	movePiece := boardState.board[m.from]
	if movePiece&0x0F != PAWN_MASK {
		return false
	}

	rank := Rank(m.to)
	return (movePiece == WHITE_MASK|PAWN_MASK && rank >= 6) || (rank <= 3)
}
