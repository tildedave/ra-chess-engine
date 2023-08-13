package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type TacticsOptions struct {
	thinkingtimeMs       uint
	epdRegex             string
	tacticsDebug         string
	tacticsHashVariation string
	tacticsDepth         uint
}

func RunTacticsFile(epdFile string, variation string, options TacticsOptions) (bool, error) {
	successPositions := 0
	totalPositions := 0

	lines, err := ParseAndFilterEpdFile(epdFile, options.epdRegex)
	if err != nil {
		return false, err
	}

	if variation != "" && len(lines) > 1 {
		return false, fmt.Errorf("Can only specify variation if regex filters to 1 positions, got %d", len(lines))
	}

	var totalStats SearchStats
	for _, line := range lines {
		prettyMove, result, err := RunTacticsFen(line.fen, variation, options)

		if err != nil {
			return false, err
		}

		totalStats.add(result.stats)

		var res string
		totalPositions++

		var moveToCheck string
		var wantMatch bool
		var desiredSummary string

		if line.bestMove != "" {
			moveToCheck = line.bestMove
			wantMatch = true
			desiredSummary = "expected"
		} else if line.avoidMove != "" {
			moveToCheck = line.avoidMove
			wantMatch = false
			desiredSummary = "avoid"
		}

		var success bool
		if moveToCheck != "" {
			var moveMatches bool
			if strings.Contains(moveToCheck, prettyMove) ||
				strings.Contains(moveToCheck, SquareToAlgebraicString(result.move.from)+SquareToAlgebraicString(result.move.to)) {
				moveMatches = true
			}
			success = moveMatches == wantMatch
		} else {
			// for now assume no move specified means checkmate is the desired result
			moveToCheck = "Mate"
			success = result.flags == CHECKMATE_FLAG
		}

		if prettyMove != "" && success {
			res = "\033[1;32mOK\033[0m"
			successPositions++
		} else {
			res = "\033[1;31mFAIL\033[0m"
		}
		fmt.Printf("[%s - %s] %s=%s move=%s result=%s\n",
			line.name, res, desiredSummary, moveToCheck, prettyMove, result.String())
	}

	fmt.Printf("Complete.  %d/%d positions correct (%.2f%%)\n", successPositions, totalPositions,
		100.0*float64(successPositions)/float64(totalPositions))
	fmt.Printf("Final stats %s", totalStats.String())
	if totalPositions == successPositions {
		return true, nil
	}

	return false, nil
}

func RunTacticsFen(fen string, variation string, options TacticsOptions) (string, SearchResult, error) {
	boardState, err := CreateBoardStateFromFENStringWithVariation(fen, variation)
	if err != nil {
		return "", SearchResult{}, err
	}

	ch := make(chan SearchResult)
	thinkingChan := make(chan ThinkingOutput)
	output := bufio.NewWriter(os.Stderr)

	output.Write([]byte(boardState.ToString()))
	output.WriteRune('\n')
	output.Flush()

	go func() {
		for thinkingOutput := range thinkingChan {
			if thinkingOutput.ply > 0 {
				sendThinkingOutput(output, thinkingOutput)
			}
		}
	}()

	config := ExternalSearchConfig{}
	config.isDebug = options.tacticsDebug != ""
	config.debugMoves = options.tacticsDebug
	config.searchToDepth = options.tacticsDepth
	thinkingTimeMs := options.thinkingtimeMs
	if options.tacticsDepth != 0 {
		thinkingTimeMs = INFINITY
	}

	// Question: Why are stats needed both in the thinkAndChooseMove and
	// returned in the SearchResult?
	stats := SearchStats{}
	go thinkAndChooseMove(&boardState, thinkingTimeMs, &stats, config, ch, thinkingChan)
	result := <-ch

	output.Flush()

	if (result.move == Move{}) {
		// no result was given in thinking time :(
		return "", result, nil
	}

	if options.tacticsHashVariation != "" {
		// "Wiggle room" to allow search to abort
		time.Sleep(time.Duration(200) * time.Millisecond)

		fmt.Println("---- Transposition Table information ----")
		moveList, err := VariationToMoveList(options.tacticsHashVariation, &boardState)
		if err != nil {
			return "", SearchResult{}, err
		}

		hasEntry, e := ProbeTranspositionTable(&boardState)
		fmt.Printf("%s - %s\n", boardState.ToString(), e.String())
		for i := 0; i < len(moveList) && hasEntry; i++ {
			boardState.ApplyMove(moveList[i])
			hasEntry, e = ProbeTranspositionTable(&boardState)
			fmt.Printf("%s - %s\n", boardState.ToString(), e.String())
		}

		for i := len(moveList) - 1; i >= 0; i-- {
			boardState.UnapplyMove(moveList[i])
		}

	}
	// .... can't use variation from transposition table while boardState is still being
	// searched, bad things

	return MoveToPrettyString(result.move, &boardState), result, nil
}
