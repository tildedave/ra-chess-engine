package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
)

var _ = fmt.Println

func InitializeLogger() {
	file, err := os.Create("/tmp/ra-chess-engine.log")
	if err != nil {
		panic(err)
	}
	logger = log.New(file, "", log.LstdFlags|log.Lshortfile)
	logger.Println("Starting up!")
}

func main() {
	startingFen := flag.String("fen", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "Fen board")
	isPerft := flag.Bool("perft", false, "Perft mode")
	perftDepth := flag.Uint("perftdepth", 5, "Perft depth to search")
	perftChecks := flag.Bool("countchecks", false, "Perft: count check positions (slower)")
	perftSanityCheck := flag.Bool("sanitycheck", false, "Perft: sanity check board and moves (slower)")
	perftJSONFile := flag.String("perftjson", "", "JSON specification")
	perftPrintMoves := flag.Bool("printmoves", false, "Perft: print all generates moves at final depth")
	isTactics := flag.Bool("tactics", false, "Tactics mode")
	tacticsEpdFile := flag.String("tacticsepd", "", "Tactics file in EPD format")
	tacticsThinkingTime := flag.Uint("tacticsthinkingtime", 500, "Time to think per position (ms)")
	tacticsRegex := flag.String("tacticsregex", "", "Run only tactics matching the given id")

	flag.Parse()

	InitializeLogger()
	var success = true
	var err error

	if *isPerft || *perftJSONFile != "" {
		var options PerftOptions
		options.checks = *perftChecks
		options.sanityCheck = *perftSanityCheck
		options.perftPrintMoves = *perftPrintMoves

		if *perftJSONFile != "" {
			success, err = RunPerftJson(*perftJSONFile, options)
		} else {
			success, err = RunPerft(*startingFen, *perftDepth, options)
		}
	} else if *isTactics || *tacticsEpdFile != "" {
		var options TacticsOptions
		options.thinkingtimeMs = *tacticsThinkingTime
		options.tacticsregex = *tacticsRegex

		success, err = RunTacticsFile(*tacticsEpdFile, options)
	} else {
		// xboard mode
		scanner := bufio.NewScanner(os.Stdin)
		output := bufio.NewWriter(os.Stdout)

		success, err = RunXboard(scanner, output)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else if success {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
