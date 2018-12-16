package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"
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
	perftChecks := flag.Bool("perftchecks", false, "Perft: count check positions (slower)")
	perftSanityCheck := flag.Bool("sanitycheck", false, "Perft: sanity check board and moves (slower)")
	perftJSONFile := flag.String("perftjson", "", "JSON specification")
	perftPrintMoves := flag.Bool("printmoves", false, "Perft: print all generates moves at final depth")
	perftDivide := flag.Bool("perftdivide", false, "Perft: print divide of all moves at top depth")
	perftCpuProfile := flag.String("perftcpuprofile", "", "Perft: file to write CPU profile to")
	isTactics := flag.Bool("tactics", false, "Tactics mode")
	tacticsEpdFile := flag.String("tacticsepd", "", "Tactics file in EPD format")
	tacticsFen := flag.String("tacticsfen", "", "Tactics position in FEN format")
	tacticsThinkingTime := flag.Uint("tacticsthinkingtime", 500, "Time to think per position (ms)")
	tacticsRegex := flag.String("tacticsregex", "", "Run only tactics matching the given id")
	tacticsDebug := flag.Bool("tacticsdebug", false, "Output more information during tactics")
	isMagic := flag.Bool("magic", false, "Generate magic bitboard constants (write to rook-magics.json and bishop-magics.json)")

	flag.Parse()

	InitializeLogger()
	var success = true
	var err error

	if *isPerft || *perftJSONFile != "" {
		var options PerftOptions
		options.checks = *perftChecks
		options.sanityCheck = *perftSanityCheck
		options.perftPrintMoves = *perftPrintMoves
		options.divide = *perftDivide
		options.depth = *perftDepth

		var f *os.File
		if *perftCpuProfile != "" {
			f, err = os.Create(*perftCpuProfile)
			if err != nil {
				log.Fatal("could not create CPU profile: ", err)
			}
			if err := pprof.StartCPUProfile(f); err != nil {
				log.Fatal("could not start CPU profile: ", err)
			}
		}

		start := time.Now()
		if *perftJSONFile != "" {
			success, err = RunPerftJson(*perftJSONFile, options)
		} else {
			success, err = RunPerft(*startingFen, *perftDepth, options)
		}

		if f != nil {
			pprof.StopCPUProfile()
			f.Close()
		}
		fmt.Printf("Total time: %s\n", time.Since(start))

	} else if *isTactics || *tacticsEpdFile != "" || *tacticsFen != "" {
		var options TacticsOptions
		options.thinkingtimeMs = *tacticsThinkingTime
		options.tacticsRegex = *tacticsRegex
		options.tacticsDebug = *tacticsDebug

		if *tacticsEpdFile != "" {
			success, err = RunTacticsFile(*tacticsEpdFile, options)
		} else if *tacticsFen != "" {
			prettyMove, err := RunTacticsFen(*tacticsFen, options)
			if err != nil {
				fmt.Printf("Move: %s\n", prettyMove)
			}
		}
	} else if *isMagic {
		GenerateMagicBitboards()
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
