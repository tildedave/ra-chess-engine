package main

import (
	"bufio"
	"fmt"
	"regexp"
)

var _ = fmt.Println

// https://www.gnu.org/software/xboard/engine-intf.html

type XboardState struct {
	boardState *BoardState
	forceMode  bool
	randomMode bool

	// TODO: time control per side
	// TODO: recent sd limit
}

const (
	ACTION_NOTHING = iota
	ACTION_QUIT    = iota
	ACTION_MOVE    = iota
)

func RunXboard(scanner *bufio.Scanner, output *bufio.Writer) (bool, error) {
	var state XboardState
	var action int = ACTION_NOTHING

ReadLoop:
	for scanner.Scan() {
		action, state = ProcessXboardCommand(scanner.Text(), state)
		switch action {
		case ACTION_NOTHING:

		case ACTION_QUIT:
			break ReadLoop

		case ACTION_MOVE:
			// search the position and make a move
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}

	return true, nil
}

var protoverRegexp = regexp.MustCompile("protover \\d")
var variantRegexp = regexp.MustCompile("variant \\w+")
var moveRegexp = regexp.MustCompile("^([abcdefgh][1-8]){2}(nbqr)?$")

func ProcessXboardCommand(command string, state XboardState) (int, XboardState) {
	var action int = ACTION_NOTHING
	switch {

	case command == "xboard":
		// This command will be sent once immediately after your engine process is started. You can use it
		// to put your engine into "xboard mode" if that is needed. If your engine prints a prompt to ask
		// for user input, you must turn off the prompt and output a newline when the "xboard" command comes in.

		fallthrough

	case protoverRegexp.MatchString(command):
		// TODO
		fallthrough

	case command == "accepted":

	case command == "rejected":
		// These commands may be sent to your engine in reply to the "feature" command; see its documentation below.

	case command == "new":
		// Reset the board to the standard chess starting position. Set White on move. Leave force mode and
		// set the engine to play Black. Associate the engine's clock with Black and the opponent's clock
		// with White. Reset clocks and time controls to the start of a new game. Use wall clock for time
		// measurement. Stop clocks. Do not ponder on this move, even if pondering is on. Remove any search
		// depth limit previously set by the sd command.

		boardState := CreateInitialBoardState()
		state.boardState = &boardState
		state.forceMode = false
		state.randomMode = false

	case variantRegexp.MatchString(command):
		// we don't support any of these :)
		fallthrough

	case command == "quit":
		// The chess engine should immediately exit. This command is used when xboard is itself exiting,
		// and also between games if the -xreuse command line option is given (or -xreuse2 for the second
		// engine). See also Signals above.

		action = ACTION_QUIT

	case command == "random":
		state.randomMode = true

	case command == "force":
		// Set the engine to play neither color ("force mode"). Stop clocks. The engine should check that
		// moves received in force mode are legal and made in the proper turn, but should not think,
		// ponder, or make moves of its own.

		state.forceMode = true

	case command == "go":
		// Leave force mode and set the engine to play the color that is on move. Associate the engine's
		// clock with the color that is on move, the opponent's clock with the color that is not on move.
		// Start the engine's clock. Start thinking and eventually make a move.

		state.forceMode = false
		action = ACTION_MOVE

	case command == "playother":
		// (This command is new in protocol version 2. It is not sent unless you enable it with the feature
		// command.) Leave force mode and set the engine to play the color that is not on move. Associate
		// the opponent's clock with the color that is on move, the engine's clock with the color that is
		// not on move. Start the opponent's clock. If pondering is enabled, the engine should begin
		// pondering. If the engine later receives a move, it should start thinking and eventually reply.

	case moveRegexp.MatchString(command):
		move, err := ParseXboardMove(command, state.boardState)
		if err != nil {
			// TODO: yell at xboard
		}

		isLegal, err := state.boardState.IsMoveLegal(move)
		if !isLegal {
			// TODO: yell at xboard
		}

		state.boardState.ApplyMove(move)
		if !state.forceMode {
			action = ACTION_MOVE
		}
	}

	return action, state
}

// Parse a move command given from xboard: https://www.gnu.org/software/xboard/engine-intf.html#4
// Doesn't test move legality (e.g. would put the moving player in check, castle is legal, etc)

func ParseXboardMove(command string, boardState *BoardState) (Move, error) {
	var move Move

	from := command[0:2]
	to := command[2:4]
	var promotion byte
	if len(command) == 5 {
		promotion = command[4]
	}

	from_sq, from_err := ParseAlgebraicSquare(from)
	to_sq, to_err := ParseAlgebraicSquare(to)

	if from_err != nil {
		return move, from_err
	}

	if to_err != nil {
		return move, to_err
	}

	isCapture := false
	isKingsideCastle := false
	isQueensideCastle := false

	fromPiece := boardState.PieceAtSquare(from_sq)
	destPiece := boardState.PieceAtSquare(to_sq)

	if destPiece != 0 || boardState.boardInfo.enPassantTargetSquare == to_sq {
		isCapture = true
	}

	if boardState.whiteToMove && fromPiece == KING_MASK|WHITE_MASK && from_sq == SQUARE_E1 {
		if to_sq == SQUARE_G1 {
			isKingsideCastle = true
		} else if to_sq == SQUARE_C1 {
			isQueensideCastle = true
		}
	} else if !boardState.whiteToMove && fromPiece == KING_MASK|BLACK_MASK && from_sq == SQUARE_E8 {
		if to_sq == SQUARE_G8 {
			isKingsideCastle = true
		} else if to_sq == SQUARE_C8 {
			isQueensideCastle = true
		}
	}

	if isKingsideCastle {
		move = CreateKingsideCastle(from_sq, to_sq)
	} else if isQueensideCastle {
		move = CreateQueensideCastle(from_sq, to_sq)
	} else if promotion != 0 {
		var pieceMask byte
		switch promotion {
		case 'b':
			pieceMask = BISHOP_MASK
		case 'q':
			pieceMask = QUEEN_MASK
		case 'r':
			pieceMask = ROOK_MASK
		case 'n':
			pieceMask = KNIGHT_MASK
		}
		if isCapture {
			move = CreatePromotionCapture(from_sq, to_sq, pieceMask)
		} else {
			move = CreatePromotion(from_sq, to_sq, pieceMask)
		}
	} else if isCapture {
		move = CreateCapture(from_sq, to_sq)
	} else {
		move = CreateMove(from_sq, to_sq)
	}

	return move, nil
}
