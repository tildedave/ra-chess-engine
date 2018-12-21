package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type EpdLine struct {
	fen      string
	bestMove string
	name     string
}

func ParseAndFilterEpdFile(epdFile string, regex string) ([]EpdLine, error) {
	lines, err := ParseEpdFile(epdFile)
	if err != nil {
		return lines, err
	}

	if regex != "" {
		lines = FilterEpdLines(lines, regex)
	}

	return lines, nil
}

func ParseEpdFile(epdFile string) ([]EpdLine, error) {
	lines := make([]EpdLine, 0)
	file, err := os.Open(epdFile)

	if err != nil {
		return lines, err
	}
	scanner := bufio.NewScanner(file)
	defer file.Close()

	totalPositions := 0

	for scanner.Scan() {
		totalPositions++

		line := strings.Split(scanner.Text(), ";")
		var fen string
		var bestMove string
		var name string

		if len(line) == 1 {
			// no answers
			fen = line[0]
			name = fmt.Sprintf("position-%d", (totalPositions + 1))
		} else {
			fenWithMove, nameWithID := line[0], line[1]
			arr := strings.Split(fenWithMove, "bm")
			fen, bestMove = strings.Trim(arr[0], " "), strings.Trim(arr[1], " ")
			arr2 := strings.Split(nameWithID, "id")
			name = strings.Trim(arr2[1], " \"")
		}

		lines = append(lines, EpdLine{name: name, fen: fen, bestMove: bestMove})
	}

	return lines, nil
}

func FilterEpdLines(lines []EpdLine, regex string) []EpdLine {
	matchingLines := make([]EpdLine, 0)
	for _, line := range lines {
		res, err := regexp.MatchString(regex, line.name)
		if err != nil {
			// shouldn't be possible (?)
			panic(err)
		}

		if res {
			matchingLines = append(matchingLines, line)
		}
	}

	return matchingLines
}
