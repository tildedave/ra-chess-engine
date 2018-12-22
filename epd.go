package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type EpdLine struct {
	fen       string
	bestMove  string
	avoidMove string
	name      string
}

// ParseAndFilterEpdFile parses a slice of EpdLine objects from a passed in filename.
// The slice is then filtered down to only only objects with a name that match the regex.
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

// ParseEpd file parses a slice of EpdLine objects from a passed in filename.
// An EPD is assumed to be a file of lines of the following format:
// <fen> bm <move>; id <name>
// <fen> am <move>; id <name>
// <fen> am <move1> <move2> <move3>; id <name>
//
// For example:
// r5r1/pQ5p/1qp2R2/2k1p3/4P3/2PP4/P1P3PP/6K1 w - - bm Rxc6; id "testWac126";
// 4r3/p1p1rpbk/b1n3p1/1N1p1q1p/3P1B1P/1PN2PP1/P5Q1/2RR2K1 w - - am Nxc7; id "arasan6.12";
//
// It's also legal to not specify a bm/am or an id, for example:
// 4rk2/2p1n1bQ/1p2Bpp1/1q2B1N1/p1b1PP2/6P1/P1P5/3r1RK1 w - -
//
// How the last case is handled is different for perft/tactics/eval.  Tactics will assume
// that the desired move is checkmate for determining correctness.
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
		var avoidMove string
		var name string

		if len(line) == 1 {
			// no answers, just run stuff
			fen = line[0]
			name = fmt.Sprintf("position-%d", (totalPositions + 1))
		} else {
			fenWithMove, nameWithID := line[0], line[1]
			arr := strings.Split(fenWithMove, "bm")
			if len(arr) == 2 {
				fen, bestMove = strings.Trim(arr[0], " "), strings.Trim(arr[1], " ")
				arr2 := strings.Split(nameWithID, "id ")
				name = strings.Trim(arr2[1], "\"")
			} else {
				arr = strings.Split(fenWithMove, "am")
				// TODO: handle no bm or no am here (I have a file for perft positions with ;D1 ;D2)
				fen, avoidMove = strings.Trim(arr[0], " "), strings.Trim(arr[1], " ")
				arr2 := strings.Split(nameWithID, "id ")
				name = strings.Trim(arr2[1], "\"")
			}
		}

		lines = append(lines, EpdLine{
			name:      name,
			fen:       fen,
			bestMove:  bestMove,
			avoidMove: avoidMove,
		})
	}

	return lines, nil
}

// FilterEpdLines returns an epd slice with only epds that have a name matching the given regex.
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
