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
	tacticsregex   string
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

		if options.tacticsregex != "" {
			res, err := regexp.MatchString(options.tacticsregex, name)
			if err != nil {
				return false, err
			}
			if !res {
				continue
			}
		}

		boardState, err := CreateBoardStateFromFENString(fen)

		if err != nil {
			return false, err
		}

		ch := make(chan Move)
		thinkingChan := make(chan ThinkingOutput)
		output := bufio.NewWriter(os.Stderr)

		go func() {
			for thinkingOutput := range thinkingChan {
				if thinkingOutput.ply > 0 {
					sendThinkingOutput(output, thinkingOutput)
				}
			}
		}()

		go thinkAndChooseMove(&boardState, options.thinkingtimeMs, ch, thinkingChan)
		move := <-ch

		prettyMove := MoveToPrettyString(move, &boardState)

		var res string
		totalPositions++
		if strings.HasPrefix(bestMove, prettyMove) {
			res = "\033[1;32mOK\033[0m"
			successPositions++
		} else {
			res = "\033[1;31mFAIL\033[0m"
		}
		fmt.Printf("[%s - %s] expected=%s result=%s\n", name, res, bestMove, prettyMove)
	}

	fmt.Printf("Complete.  %d/%d positions correct (%.2f%%)\n", successPositions, totalPositions,
		float64(successPositions)/float64(totalPositions))
	if totalPositions == successPositions {
		return true, nil
	} else {
		return false, nil
	}
}
