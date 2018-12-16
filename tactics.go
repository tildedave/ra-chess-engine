package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type TacticsOptions struct {
	thinkingtimeMs uint
	tacticsRegex   string
	tacticsDebug   bool
}

func RunTacticsFile(epdFile string, options TacticsOptions) (bool, error) {
	file, err := os.Open(epdFile)
	if err != nil {
		return false, err
	}
	scanner := bufio.NewScanner(file)
	defer file.Close()

	successPositions := 0
	totalPositions := 0

	for scanner.Scan() {
		line := strings.Split(scanner.Text(), ";")
		fenWithMove, nameWithID := line[0], line[1]
		arr := strings.Split(fenWithMove, "bm")
		fen, bestMove := strings.Trim(arr[0], " "), strings.Trim(arr[1], " ")
		arr2 := strings.Split(nameWithID, "id")
		name := strings.Trim(arr2[1], " \"")

		if options.tacticsRegex != "" {
			res, err := regexp.MatchString(options.tacticsRegex, name)
			if err != nil {
				return false, err
			}
			if !res {
				continue
			}
		}

		prettyMove, result, err := RunTacticsFen(fen, options)
		if err != nil {
			return false, err
		}

		var res string
		totalPositions++
		if strings.Contains(bestMove, prettyMove) {
			res = "\033[1;32mOK\033[0m"
			successPositions++
		} else {
			res = "\033[1;31mFAIL\033[0m"
		}
		fmt.Printf("[%s - %s] expected=%s result=%s (score=%d, nodes=%d, cutoffs=%d, depth=%d)\n",
			name, res, bestMove, prettyMove, result.value, result.nodes, result.hashCutoffs, result.depth)
	}

	fmt.Printf("Complete.  %d/%d positions correct (%.2f%%)\n", successPositions, totalPositions,
		100.0*float64(successPositions)/float64(totalPositions))
	if totalPositions == successPositions {
		return true, nil
	}

	return false, nil
}

func RunTacticsFen(fen string, options TacticsOptions) (string, SearchResult, error) {
	boardState, err := CreateBoardStateFromFENString(fen)

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
	config.isDebug = options.tacticsDebug

	go thinkAndChooseMove(&boardState, options.thinkingtimeMs, config, ch, thinkingChan)
	result := <-ch

	return MoveToPrettyString(result.move, &boardState), result, nil
}
