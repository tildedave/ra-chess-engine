package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

var logger *log.Logger = nil

// https://www.gnu.org/software/xboard/engine-intf.html

type XboardMove struct {
	move   Move
	result SearchResult // possibly null
}

type XboardState struct {
	boardState   *BoardState
	forceMode    bool
	post         bool
	randomMode   bool
	opponentName string
	moveHistory  []XboardMove
	err          error
	initialFEN   string

	// TODO: time control per side
	// TODO: recent sd limit
}

const (
	ACTION_NOTHING        = iota
	ACTION_QUIT           = iota
	ACTION_HALT           = iota
	ACTION_THINK          = iota
	ACTION_THINK_AND_MOVE = iota
	ACTION_MOVE_NOW       = iota
	ACTION_WAIT           = iota
	ACTION_GAME_OVER      = iota
	ACTION_ERROR          = iota
)

type ThinkingOutput struct {
	ply   uint
	score int
	time  int64
	nodes uint64
	pv    string
}

func RunXboard(scanner *bufio.Scanner, output *bufio.Writer) (bool, error) {
	var state XboardState
	var action int = ACTION_NOTHING
	sendPreamble(output)

	defer func() {
		if r := recover(); r != nil {
			logger.Println("Recovered in f", r)
		}
	}()

ReadLoop:
	for scanner.Scan() {
		command := scanner.Text()
		logger.Println("Received command: " + command)
		action, state = ProcessXboardCommand(scanner.Text(), state)
		output.WriteString(fmt.Sprintf("# action=%d\n", action))
		output.Flush()

		switch action {
		case ACTION_ERROR:
			// send error back to engine
			output.WriteString(state.err.Error() + "\n")
			output.Flush()
			state.err = nil

		case ACTION_NOTHING:
			// don't change anything we're doing now

		case ACTION_WAIT:
			// wait for opponent to move

		case ACTION_QUIT:
			break ReadLoop

		case ACTION_THINK:
			// TODO: use think code but never time out

		case ACTION_THINK_AND_MOVE:
			sendBoardAsComment(output, state.boardState)

			ch := make(chan SearchResult)
			thinkingChan := make(chan ThinkingOutput)

			go func() {
				for thinkingOutput := range thinkingChan {
					if state.post && thinkingOutput.ply > 0 {
						sendThinkingOutput(output, thinkingOutput)
					}
				}
			}()

			// TODO - smarter thinking
			go thinkAndChooseMove(state.boardState, 7400, ExternalSearchConfig{}, ch, thinkingChan)
			result := <-ch
			move := result.move

			sendStringMessage(output, fmt.Sprintf("move %s\n", MoveToXboardString(move)))
			sendStringMessage(output, fmt.Sprintf("# %s\n", result.String()))

			state.boardState.ApplyMove(move)
			xboardMove := XboardMove{move: move, result: result}
			state.moveHistory = append(state.moveHistory, xboardMove)

			sendBoardAsComment(output, state.boardState)

		case ACTION_GAME_OVER:
			sendGameAsComment(output, &state)
		}

		logger.Println("Waiting for commands...")
	}
	if err := scanner.Err(); err != nil {
		logger.Println(err)
	}

	return true, nil
}

func sendPreamble(output *bufio.Writer) {
	sendStringMessage(output, "feature myname=\"ra v0.0.1\" setboard=1 sigterm=0 sigint=0 done=1\n")
}

func sendStringMessage(output *bufio.Writer, str string) {
	logger.Print("-> " + str)
	output.WriteString(str)
	output.Flush()
}

func sendBoardAsComment(output *bufio.Writer, boardState *BoardState) {
	str := boardState.ToString()
	for _, line := range strings.Split(str, "\n") {
		sendStringMessage(output, "# "+line+"\n")
	}
}

func sendGameAsComment(output *bufio.Writer, state *XboardState) {
	boardState, err := CreateBoardStateFromFENString(state.initialFEN)
	if err != nil {
		panic("Initial FEN was invalid - should never happen")
	}

	var gameAsPgn string
	for i, xboardMove := range state.moveHistory {
		if i%2 == 0 {
			gameAsPgn += fmt.Sprintf("%d. ", (i/2)+1)
		}
		gameAsPgn += MoveToPrettyString(xboardMove.move, &boardState) + " "
		if xboardMove.result.depth > 0 {
			gameAsPgn += fmt.Sprintf("{%d/%d %s} ",
				xboardMove.result.value,
				xboardMove.result.depth,
				xboardMove.result.pv)
		}
		boardState.ApplyMove(xboardMove.move)
	}
	// TODO - include result string for maximum prettiness :)
	sendStringMessage(output, fmt.Sprintf("# %s\n", gameAsPgn))
}

func sendThinkingOutput(output *bufio.Writer, thinkingOutput ThinkingOutput) {
	sendStringMessage(output, fmt.Sprintf(
		"%d %d %d %d %s\n",
		thinkingOutput.ply,
		thinkingOutput.score,
		thinkingOutput.time,
		thinkingOutput.nodes,
		thinkingOutput.pv,
	))
}

func thinkAndChooseMove(
	boardState *BoardState,
	thinkingTimeMs uint,
	config ExternalSearchConfig,
	ch chan SearchResult,
	thinkingChan chan ThinkingOutput,
) {
	shouldAbort = false
	// NOTE - There seems to be a bug in the TT where the wrong move is being returned
	// For now clear out the TT before starting to think
	generateTranspositionTable(boardState)

	searchQuit := make(chan bool)
	resultCh := make(chan SearchResult)

	go func() {
		var i uint = 1
		// var res SearchResult

		for {
			select {
			case <-searchQuit:
				close(resultCh)
				close(searchQuit)
				return
			default:
				// TODO: having to copy the board state indicates a bug somewhere
				state := CopyBoardState(boardState)
				result := SearchWithConfig(&state, uint(i), config)
				resultCh <- result
				i = i + 1
			}
		}
	}()

	go func() {
		var bestResult SearchResult

		startTime := time.Now()
		// after no output for 50ms we check the value
		checkInterval := 50

	ThinkingLoop:
		for {
			select {
			case bestResult = <-resultCh:
				logger.Println("New result:")
				logger.Println(bestResult.String())
				thinkingChan <- ThinkingOutput{
					ply:   bestResult.depth,
					score: bestResult.value,
					time:  bestResult.time.Nanoseconds(),
					nodes: bestResult.stats.Nodes(),
					pv:    bestResult.pv,
				}

				if bestResult.flags == CHECKMATE_FLAG ||
					bestResult.flags == DRAW_FLAG ||
					bestResult.depth == MAX_DEPTH {
					logger.Println("Best result is terminal, time to stop thinking")
					break ThinkingLoop
				}

				if bestResult.move.from != bestResult.move.to && bestResult.depth == config.searchToDepth {
					logger.Printf("Only wanted to search depth %d, done", config.searchToDepth)
					break ThinkingLoop
				}

			case <-time.After(time.Duration(checkInterval) * time.Millisecond):
				if time.Since(startTime) > time.Duration(thinkingTimeMs)*time.Millisecond {
					logger.Println("Thinking time is up!")
					break ThinkingLoop
				}
			}
		}

		shouldAbort = true
		ch <- bestResult
		close(ch)
		close(thinkingChan)
		searchQuit <- true
	}()
}

var protoverRegexp = regexp.MustCompile("^protover \\d$")
var variantRegexp = regexp.MustCompile("^variant \\w+$")
var moveRegexp = regexp.MustCompile("^([abcdefgh][1-8]){2}([nbqr])?$")
var pingRegexp = regexp.MustCompile("^ping \\d$")
var resultRegexp = regexp.MustCompile("^result (1\\-0|0\\-1|1/2\\-1/2|\\*) {[^}]+}$")
var fenRegexp = regexp.MustCompile("^setboard (.*)$")
var nameRegexp = regexp.MustCompile("^name (.*)$")

func ProcessXboardCommand(command string, state XboardState) (int, XboardState) {
	var action = ACTION_NOTHING
	switch {

	case command == "xboard":
		// This command will be sent once immediately after your engine process is started. You can use it
		// to put your engine into "xboard mode" if that is needed. If your engine prints a prompt to ask
		// for user input, you must turn off the prompt and output a newline when the "xboard" command comes in.

	case protoverRegexp.MatchString(command):
		// TODO

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
		state.initialFEN = boardState.ToFENString()
		state.boardState = &boardState
		state.forceMode = false
		state.randomMode = false
		state.err = nil
		action = ACTION_HALT

	case variantRegexp.MatchString(command):
		// we don't support any variants :)

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
		action = ACTION_HALT

	case command == "go":
		if state.boardState == nil {
			action = ACTION_ERROR
			state.err = errors.New("Error (board was not initialized)")
			break
		}
		// Leave force mode and set the engine to play the color that is on move. Associate the engine's
		// clock with the color that is on move, the opponent's clock with the color that is not on move.
		// Start the engine's clock. Start thinking and eventually make a move.

		state.forceMode = false
		action = ACTION_THINK_AND_MOVE

	case command == "playother":
		// (This command is new in protocol version 2. It is not sent unless you enable it with the feature
		// command.) Leave force mode and set the engine to play the color that is not on move. Associate
		// the opponent's clock with the color that is on move, the engine's clock with the color that is
		// not on move. Start the opponent's clock. If pondering is enabled, the engine should begin
		// pondering. If the engine later receives a move, it should start thinking and eventually reply.

		state.forceMode = false
		action = ACTION_WAIT

	case moveRegexp.MatchString(command):
		// See below for the syntax of moves. If the move is illegal, print an error message; see the section
		// "Commands from the engine to xboard". If the move is legal and in turn, make it. If not in force
		// mode, stop the opponent's clock, start the engine's clock, start thinking, and eventually make a move.

		if state.boardState == nil {
			action = ACTION_ERROR
			state.err = errors.New("Illegal move (Board was not initialized correctly)")
			break
		}

		move, err := ParseXboardMove(command, state.boardState)
		if err != nil {
			action = ACTION_ERROR
			state.err = errors.New("Illegal move (" + err.Error() + ")")
			break
		}

		isLegal, err := state.boardState.IsMoveLegal(move)
		if !isLegal {
			action = ACTION_ERROR
			state.err = errors.New("Illegal move (" + err.Error() + ")")
		}

		logger.Printf("Applying move %s\n", MoveToString(move, state.boardState))
		state.boardState.ApplyMove(move)
		state.moveHistory = append(state.moveHistory, XboardMove{move: move})

		if !state.forceMode {
			action = ACTION_THINK_AND_MOVE
		}

	case command == "?":
		// Move now. If your engine is thinking, it should move immediately; otherwise, the command should
		// be ignored (treated as a no-op). It is permissible for your engine to always ignore the ? command.
		// The only bad consequence is that xboard's Move Now menu command will do nothing.

		action = ACTION_MOVE_NOW

	case pingRegexp.MatchString(command):
		// In this command, N is a decimal number. When you receive the command, reply by sending the string
		// pong N, where N is the same number you received. Important: You must not reply to a "ping" command
		// until you have finished executing all commands that you received before it. Pondering does not count;
		// if you receive a ping while pondering, you should reply immediately and continue pondering. Because
		// of the way xboard uses the ping command, if you implement the other commands in this protocol, you
		// should never see a "ping" command when it is your move; however, if you do, you must not send the
		// "pong" reply to xboard until after you send your move. For example, xboard may send "?" immediately
		// followed by "ping". If you implement the "?" command, you will have moved by the time you see the
		// subsequent ping command. Similarly, xboard may send a sequence like "force", "new", "ping". You must
		// not send the pong response until after you have finished executing the "new" command and are ready
		// for the new game to start.
		//
		// The ping command is new in protocol version 2 and will not be sent unless you enable it with the
		// "feature" command. Its purpose is to allow several race conditions that could occur in previous
		// versions of the protocol to be fixed, so it is highly recommended that you implement it. It is
		// especially important in simple engines that do not ponder and do not poll for input while thinking,
		// but it is needed in all engines.

		// TODO: implement this

	case command == "draw":
		// The engine's opponent offers the engine a draw. To accept the draw, send "offer draw". To decline,
		// ignore the offer (that is, send nothing). If you're playing on ICS, it's possible for the draw offer
		// to have been withdrawn by the time you accept it, so don't assume the game is over because you accept
		// a draw offer. Continue playing until xboard tells you the game is over. See also "offer draw" below.

		// TODO: implement this

		action = ACTION_NOTHING

	case resultRegexp.MatchString(command):
		// After the end of each game, xboard will send you a result command. You can use this command to trigger
		// learning. RESULT is either 1-0, 0-1, 1/2-1/2, or *, indicating whether white won, black won, the game
		// was a draw, or the game was unfinished. The COMMENT string is purely a human-readable comment; its
		// content is unspecified and subject to change. In ICS mode, it is passed through from ICS uninterpreted.
		// Example:
		//
		// result 1-0 {White mates}
		// Here are some notes on interpreting the "result" command. Some apply only to playing on ICS ("Zippy" mode).
		//
		// If you won but did not just play a mate, your opponent must have resigned or forfeited. If you lost
		// but were not just mated, you probably forfeited on time, or perhaps the operator resigned manually. If
		// there was a draw for some nonobvious reason, perhaps your opponent called your flag when he had
		// insufficient mating material (or vice versa), or perhaps the operator agreed to a draw manually.
		//
		// You will get a result command even if you already know the game ended -- for example, after you just
		// checkmated your opponent. In fact, if you send the "RESULT {COMMENT}" command (discussed below), you will
		// simply get the same thing fed back to you with "result" tacked in front. You might not always get a
		// "result *" command, however. In particular, you won't get one in local chess engine mode when the user
		// stops playing by selecting Reset, Edit Game, Exit or the like.

		action = ACTION_GAME_OVER

	case fenRegexp.MatchString(command):
		// The setboard command is the new way to set up positions, beginning in protocol version 2. It is not used
		// unless it has been selected with the feature command. Here FEN is a position in Forsythe-Edwards Notation,
		// as defined in the PGN standard. Note that this PGN standard referred to here only applies to normal Chess;
		// Obviously in variants that cannot be described by a FEN for normal Chess, e.g. because the board is not
		// 8x8, other pieces then PNBRQK participate, there are holdings that need to be specified, etc., xboard will
		// use a FEN format that is standard or suitable for that variant. In particular, in FRC or CRC, WinBoard will
		// use Shredder-FEN or X-FEN standard, i.e. it can use the rook-file indicator letter to represent a castling
		// right (like HAha) whenever it wants, but if it uses KQkq, this will always refer to the outermost rook on
		// the given side.

		// Illegal positions: Note that either setboard or edit can be used to send an illegal position to the engine.
		// The user can create any position with xboard's Edit Position command (even, say, an empty board, or a board
		// with 64 white kings and no black ones). If your engine receives a position that it considers illegal, I
		// suggest that you send the response "tellusererror Illegal position", and then respond to any attempted move
		// with "Illegal move" until the next new, edit, or setboard command.

		action = ACTION_HALT

		fenString := fenRegexp.FindStringSubmatch(command)[1]
		boardState, err := CreateBoardStateFromFENString(fenString)

		if err != nil {
			state.err = errors.New("Error (" + err.Error() + ")")
			action = ACTION_ERROR

			break
		}
		state.boardState = &boardState

	case command == "hint":
		// If the user asks for a hint, xboard sends your engine the command "hint". Your engine should respond with
		// "Hint: xxx", where xxx is a suggested move. If there is no move to suggest, you can ignore the hint command
		// (that is, treat it as a no-op).

	case command == "undo":
		// If the user asks to back up one move, xboard will send you the "undo" command. xboard will not send this
		// command without putting you in "force" mode first, so you don't have to worry about what should happen if
		// the user asks to undo a move your engine made. (GNU Chess 4 actually switches to playing the opposite color
		// in this case.)

		idx := len(state.moveHistory) - 1
		xboardMove := state.moveHistory[idx]
		state.moveHistory = state.moveHistory[:idx]
		state.boardState.UnapplyMove(xboardMove.move)

	case command == "remove":
		// If the user asks to retract a move, xboard will send you the "remove" command. It sends this command only
		// when the user is on move. Your engine should undo the last two moves (one for each player) and continue
		// playing the same color.

		idx := len(state.moveHistory) - 2
		xboardMove1 := state.moveHistory[idx]
		xboardMove2 := state.moveHistory[idx+1]
		state.moveHistory = state.moveHistory[:idx]
		state.boardState.UnapplyMove(xboardMove1.move)
		state.boardState.UnapplyMove(xboardMove2.move)
		action = ACTION_THINK_AND_MOVE

	case command == "hard":
		// Turn on pondering (thinking on the opponent's time, also known as "permanent brain"). xboard will not make
		// any assumption about what your default is for pondering or whether "new" affects this setting.

	case command == "easy":
		// Turn off pondering.

	case command == "post":
		// Turn on thinking/pondering output. See Thinking Output section.
		state.post = true

	case command == "nopost":
		// Turn off thinking/pondering output.
		state.post = false

	case command == "analyze":
		// Enter analyze mode. See Analyze Mode section.
		action = ACTION_THINK

	case nameRegexp.MatchString(command):
		// This command informs the engine of its opponent's name. When the engine is playing on a chess server, xboard
		// obtains the opponent's name from the server. When the engine is playing locally against a human user, xboard
		// obtains the user's login name from the local operating system. When the engine is playing locally against
		// another engine, xboard uses either the other engine's filename or the name that the other engine supplied in
		// the myname option to the feature command. By default, xboard uses the name command only when the engine is
		// playing on a chess server. Beginning in protocol version 2, you can change this with the name option to the
		// feature command; see below.

		state.opponentName = nameRegexp.FindStringSubmatch(command)[1]

		// TODO: We should say hi!

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
	isEnPassantCapture := false
	isKingsideCastle := false
	isQueensideCastle := false

	fromPiece := boardState.PieceAtSquare(from_sq)
	destPiece := boardState.PieceAtSquare(to_sq)

	if destPiece != 0 {
		isCapture = true
	}
	if boardState.boardInfo.enPassantTargetSquare == to_sq && boardState.boardInfo.enPassantTargetSquare > 0 {
		isEnPassantCapture = true
	}

	if boardState.sideToMove == WHITE_OFFSET && fromPiece == KING_MASK|WHITE_MASK && from_sq == SQUARE_E1 {
		if to_sq == SQUARE_G1 {
			isKingsideCastle = true
		} else if to_sq == SQUARE_C1 {
			isQueensideCastle = true
		}
	} else if boardState.sideToMove == BLACK_OFFSET && fromPiece == KING_MASK|BLACK_MASK && from_sq == SQUARE_E8 {
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
	} else if isEnPassantCapture {
		move = CreateEnPassantCapture(from_sq, to_sq)
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
	} else {
		move = CreateMove(from_sq, to_sq)
	}

	return move, nil
}

func MoveToXboardString(move Move) string {
	from := SquareToAlgebraicString(move.from)
	to := SquareToAlgebraicString(move.to)
	if move.IsPromotion() {
		var piece rune
		switch move.flags & 0x0F {
		case ROOK_MASK:
			piece = 'r'
		case QUEEN_MASK:
			piece = 'q'
		case KNIGHT_MASK:
			piece = 'n'
		case BISHOP_MASK:
			piece = 'b'
		}

		return fmt.Sprintf("%s%s%c", from, to, piece)
	}
	return fmt.Sprintf("%s%s", from, to)
}

func MoveArrayToXboardString(moves []Move) string {
	str := ""
	for i, move := range moves {
		if i != 0 {
			str += " "
		}

		str += MoveToXboardString(move)
	}
	return str

}
