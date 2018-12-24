package main

import (
	"bufio"
	"errors"
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
	startingFen := flag.String("fen", "", "Fen board")
	variation := flag.String("variation", "", "Variation to apply prior to perft/tactics/eval (pair with --fen)")
	epdFile := flag.String("epd", "", "Position file in EPD format")
	epdRegex := flag.String("epdregex", "", "Run only positions matching the given id")
	cpuProfile := flag.String("cpuprofile", "", "File to write CPU profile to")
	isPerft := flag.Bool("perft", false, "Perft mode")
	perftDepth := flag.Uint("perftdepth", 5, "Perft depth to search")
	perftChecks := flag.Bool("perftchecks", false, "Perft: count check positions (slower)")
	perftSanityCheck := flag.Bool("sanitycheck", false, "Perft: sanity check board and moves (slower)")
	perftJSONFile := flag.String("perftjson", "", "JSON specification")
	perftPrintMoves := flag.Bool("printmoves", false, "Perft: print all generates moves at final depth")
	perftDivide := flag.Bool("perftdivide", false, "Perft: print divide of all moves at top depth")
	isTactics := flag.Bool("tactics", false, "Tactics mode")
	tacticsThinkingTime := flag.Uint("tacticsthinkingtime", 1500, "Time to think per position (ms)")
	tacticsDebug := flag.String("tacticsdebug", "", "Output more information during tactics if the move matches the string")
	tacticsDepth := flag.Uint("tacticsdepth", 0, "Only run tactics search for the given depth")
	tacticsHashVariation := flag.String("tacticshashvariation", "", "Output transposition table information for given variation")
	isMagic := flag.Bool("magic", false, "Generate magic bitboard constants (write to rook-magics.json and bishop-magics.json)")
	isEval := flag.Bool("eval", false, "Run evaluation on the specified position or positions (no search)")

	flag.Parse()

	InitializeLogger()
	InitializeMoveBitboards()
	var success = true
	var err error

	var f *os.File
	if *cpuProfile != "" {
		f, err = os.Create(*cpuProfile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
	}

	if *isPerft || *perftJSONFile != "" {
		var options PerftOptions
		options.checks = *perftChecks
		options.sanityCheck = *perftSanityCheck
		options.perftPrintMoves = *perftPrintMoves
		options.divide = *perftDivide
		options.depth = *perftDepth

		start := time.Now()
		if *perftJSONFile != "" {
			success, err = RunPerftJson(*perftJSONFile, options)
		} else {
			if *startingFen == "" {
				*startingFen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
			}
			success, err = RunPerft(*startingFen, *variation, *perftDepth, options)
		}

		fmt.Printf("Total time: %s\n", time.Since(start))
	} else if *isTactics {
		var options TacticsOptions
		options.thinkingtimeMs = *tacticsThinkingTime
		options.epdRegex = *epdRegex
		options.tacticsDebug = *tacticsDebug
		options.tacticsDepth = *tacticsDepth
		options.tacticsHashVariation = *tacticsHashVariation

		if *epdFile != "" {
			success, err = RunTacticsFile(*epdFile, *variation, options)
		} else if *startingFen != "" {
			prettyMove, result, err := RunTacticsFen(*startingFen, *variation, options)
			if err == nil {
				fmt.Printf("move=%s result=%s\n", prettyMove, SearchResultToString(result))
			}
		} else {
			err = errors.New("Must specify either an EPD file or a fen argument")
		}
	} else if *isMagic {
		GenerateMagicBitboards()
	} else if *isEval {
		var options EvalOptions
		options.epdRegex = *epdRegex

		if *epdFile != "" {
			success, err = RunEvalFile(*epdFile, *variation, options)
		} else if *startingFen != "" {
			var eval BoardEval
			eval, err = RunEvalFen(*startingFen, *variation, options)
			fmt.Println(BoardEvalToString(eval))
		} else {
			err = errors.New("Must specify either an EPD file or a fen argument")
		}
	} else {
		// xboard mode
		scanner := bufio.NewScanner(os.Stdin)
		output := bufio.NewWriter(os.Stdout)

		success, err = RunXboard(scanner, output)
	}

	if f != nil {
		pprof.StopCPUProfile()
		f.Close()
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
