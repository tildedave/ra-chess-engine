package main

import (
	"flag"
	"fmt"
	"os"
)

var _ = fmt.Println

func main() {
	startingFen := flag.String("fen", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "Fen board")
	isPerft := flag.Bool("perft", false, "Perft mode")
	perftDepth := flag.Uint("perftdepth", 5, "Perft depth to search")
	perftChecks := flag.Bool("countchecks", false, "Perft: count check positions (slower)")
	perftSanityCheck := flag.Bool("sanitycheck", false, "Perft: sanity check board and moves (slower)")
	perftJsonFile := flag.String("perftjson", "", "JSON specification")
	perftPrintMoves := flag.Bool("printmoves", false, "Perft: print all generates moves at final depth")

	flag.Parse()

	var success bool = true
	var err error = nil

	if *isPerft || perftJsonFile != nil {
		var options PerftOptions
		options.checks = *perftChecks
		options.sanityCheck = *perftSanityCheck
		options.perftPrintMoves = *perftPrintMoves

		if *perftJsonFile != "" {
			success, err = RunPerftJson(*perftJsonFile, options)
		} else {
			success, err = RunPerft(*startingFen, *perftDepth, options)
		}
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
