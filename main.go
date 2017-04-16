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

	flag.Parse()
	if *isPerft {
		for i := uint(0); i <= *perftDepth; i++ {
			board, err := CreateBoardStateFromFENString(*startingFen)
			if err == nil {
				fmt.Println(Perft(&board, i))
			} else {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	} else {
		// TODO: winboard/console mode
	}
}
