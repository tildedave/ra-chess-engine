package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	startingFen := flag.String("fen", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "Fen board")
	isPerft := flag.Bool("perft", true, "Perft mode")
	perftDepth := flag.Uint("perftdepth", 5, "Perft depth to search")
	perftChecks := flag.Bool("countchecks", false, "Perft: count check positions (slower)")
	perftSanityCheck := flag.Bool("sanitycheck", false, "Perft: sanity check board and moves (slower)")
	perftPrintMoves := flag.Bool("printmoves", false, "Perft: print all generates moves at final depth")

	flag.Parse()

	if *isPerft {
		var options PerftOptions
		options.checks = *perftChecks
		options.sanityCheck = *perftSanityCheck
		options.perftPrintMoves = *perftPrintMoves

		for i := uint(0); i <= *perftDepth; i++ {
			board, err := CreateBoardStateFromFENString(*startingFen)
			if err == nil {
				fmt.Println(Perft(&board, i, options))
			} else {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	} else {
		// TODO: winboard/console mode
	}
}
