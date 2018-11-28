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
		output := bufio.NewWriter(os.Stdout)

		go func() {
			for thinkingOutput := range thinkingChan {
				if thinkingOutput.ply > 0 {
					sendThinkingOutput(output, thinkingOutput)
				}
			}
		}()

		go thinkAndChooseMove(&boardState, options.thinkingtimeMs, ch, thinkingChan)
		move := <-ch
		fmt.Printf("[%s] expected=%s result=%s\n", name, bestMove, MoveToPrettyString(move, &boardState))
	}

	return true, nil
}
