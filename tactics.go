package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type TacticsOptions struct {
	thinkingtimeMs uint
	epdRegex       string
	tacticsDebug   string
	tacticsDepth   uint
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

	for _, line := range lines {
		prettyMove, result, err := RunTacticsFen(line.fen, variation, options)
		if err != nil {
			return false, err
		}

		var res string
		totalPositions++

		var moveToCheck string
		var wantMatch bool

		if line.bestMove != "" {
			moveToCheck = line.bestMove
			wantMatch = true
		} else if line.avoidMove != "" {
			moveToCheck = line.avoidMove
			wantMatch = false
		}

		var success bool
		if moveToCheck != "" {
			var moveMatches bool
			if strings.Contains(moveToCheck, prettyMove) ||
				strings.Contains(moveToCheck, MoveToString(result.move)) ||
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
		fmt.Printf("[%s - %s] expected=%s move=%s result=%s\n",
			line.name, res, moveToCheck, prettyMove, SearchResultToString(result))
	}

	fmt.Printf("Complete.  %d/%d positions correct (%.2f%%)\n", successPositions, totalPositions,
		100.0*float64(successPositions)/float64(totalPositions))
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
	config.onlySearchDepth = options.tacticsDepth

	go thinkAndChooseMove(&boardState, options.thinkingtimeMs, config, ch, thinkingChan)
	result := <-ch

	output.Flush()

	if (result.move == Move{}) {
		// no result was given in thinking time :(
		return "", result, nil
	}

	boardState.shouldAbort = true
	return MoveToPrettyString(result.move, &boardState), result, nil
}
