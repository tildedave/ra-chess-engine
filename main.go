package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

type PerftSpecification struct {
	Depth uint   `json:depth`
	Nodes uint   `json:nodes`
	Fen   string `json:fen`
}

func main() {
	startingFen := flag.String("fen", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "Fen board")
	isPerft := flag.Bool("perft", true, "Perft mode")
	perftDepth := flag.Uint("perftdepth", 5, "Perft depth to search")
	perftChecks := flag.Bool("countchecks", false, "Perft: count check positions (slower)")
	perftSanityCheck := flag.Bool("sanitycheck", false, "Perft: sanity check board and moves (slower)")
	perftJsonFile := flag.String("perftjson", "", "JSON specification")
	perftPrintMoves := flag.Bool("printmoves", false, "Perft: print all generates moves at final depth")

	flag.Parse()

	if *isPerft {
		var options PerftOptions
		options.checks = *perftChecks
		options.sanityCheck = *perftSanityCheck
		options.perftPrintMoves = *perftPrintMoves

		if *perftJsonFile != "" {
			b, err := ioutil.ReadFile(*perftJsonFile)
			if err != nil {
				panic(err)
			}

			var specs []PerftSpecification
			json.Unmarshal(b, &specs)

			for _, spec := range specs {
				board, err := CreateBoardStateFromFENString(spec.Fen)
				if err != nil {
					fmt.Println("Unable to parse FEN " + spec.Fen + ", continuing")
					fmt.Println(err)
					continue
				}
				perftResult := Perft(&board, spec.Depth, options)
				if perftResult.nodes != spec.Nodes {
					fmt.Printf("NOT OK: %s (depth=%d, expected nodes=%d, actual nodes=%d)\n", spec.Fen, spec.Depth, spec.Nodes, perftResult.nodes)
				} else {
					fmt.Printf("OK: %s (depth=%d, nodes=%d)\n", spec.Fen, spec.Depth, spec.Nodes)
				}
			}
			return
		}

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
