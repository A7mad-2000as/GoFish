package chessEngine

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

const (
	EngineName   = "GoFish"
	Author       = "Ahmad & Yazan"
	InfiniteTime = -1
	NoValue      = 0
	MaxDepth     = 100
)

type UciInterface struct {
	gameSearcher GameSearcher
	evaluator    Evaluator
}

func (uciInterface *UciInterface) ReInitialize() {
	uciInterface.gameSearcher.Reset(uciInterface.evaluator)
}

func (uciInterface *UciInterface) respondToUciCommand() {
	fmt.Println("id name", EngineName)
	fmt.Println("id author", Author)

	engineOptions := uciInterface.gameSearcher.GetOptions()

	for optionName, option := range engineOptions {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("option name %s type %s ", optionName, option.optionType))
		if option.optionType != "button" {
			sb.WriteString(fmt.Sprintf("default %s ", option.defaultValue))
		}

		if option.optionType == "spin" {
			sb.WriteString(fmt.Sprintf("min %s max %s ", option.minValue, option.maxValue))
		} else if option.optionType == "combo" {
			for _, fixedValue := range option.fixedValues {
				sb.WriteString(fmt.Sprintf("var %s ", fixedValue))
			}
		}

		optionOffer := strings.TrimSpace(sb.String())
		fmt.Println(optionOffer)
	}

	fmt.Println("uciok")
}

func (uciInterface *UciInterface) respondToSetOptionCommand(setOptionCommand string) {
	commandFields := strings.Fields(setOptionCommand)
	gettingOptionValue := false

	var optionNameBuilder strings.Builder
	var optionValueBuilder strings.Builder
	for _, commandField := range commandFields {
		if commandField == "value" {
			gettingOptionValue = true
		} else if commandField != "name" && !gettingOptionValue {
			optionNameBuilder.WriteString(commandField + " ")
		} else if commandField != "name" {
			optionValueBuilder.WriteString(commandField + " ")
		}
	}

	optionName := strings.TrimSpace(optionNameBuilder.String())
	optionValue := strings.TrimSpace(optionValueBuilder.String())

	engineOptions := uciInterface.gameSearcher.GetOptions()

	engineOptions[optionName].setOption(optionValue)

}

func (uciInterface *UciInterface) respondToIsReadyCommand() {
	fmt.Println("readyok")
}

func (UciInterface *UciInterface) respondToUciNewGameCommand() {
	UciInterface.gameSearcher.ResetToNewGame()
}

func (uciInterface *UciInterface) respondToPositionCommand(positionCommand string) {
	fenString := ""
	movesString := ""

	if strings.HasPrefix(positionCommand, "startpos") {
		fenString = FENStartPosition
		movesString = strings.TrimPrefix(positionCommand, "startpos ")
	} else if strings.HasPrefix(positionCommand, "fen") {
		commandInfo := strings.TrimPrefix(positionCommand, "fen ")
		commandFields := strings.Fields(commandInfo)
		fenString = strings.Join(commandFields[0:6], " ")
		movesString = strings.Join(commandFields[6:], " ")
	}

	uciInterface.gameSearcher.InitializeSearchInfo(fenString, uciInterface.evaluator)

	if strings.HasPrefix(movesString, "moves") {
		uciMoves := strings.TrimSpace(strings.TrimPrefix(movesString, "moves "))
		for _, uciMove := range strings.Fields(uciMoves) {
			move := convertUciMoveIntoEncodedMove(uciInterface.gameSearcher.Position(), uciMove)
			uciInterface.gameSearcher.Position().DoMove(move, uciInterface.evaluator)

			positionHash := uciInterface.gameSearcher.Position().PositionHash
			uciInterface.gameSearcher.RecordPositionHash(positionHash)

			uciInterface.gameSearcher.Position().stateStackSize--
		}
	}
}

func (uciInterface *UciInterface) respondToGoCommand(goCommand string) {
	commandFields := strings.Fields(goCommand)

	sideToPlay := uciInterface.gameSearcher.Position().SideToMove
	colorIndicator := ""

	if sideToPlay == White {
		colorIndicator = "w"
	} else {
		colorIndicator = "b"
	}

	remainingTime, increment, movesToGo := int(InfiniteTime), int(NoValue), int(NoValue)
	depth, nodeCount, moveTime := uint64(MaxDepth), uint64(math.MaxUint64), uint64(NoValue)

	for index, commandField := range commandFields {
		switch commandField {
		case "movestogo":
			movesToGo, _ = strconv.Atoi(commandFields[index+1])
		case "depth":
			depth, _ = strconv.ParseUint(commandFields[index+1], 10, 8)
		case "nodes":
			nodeCount, _ = strconv.ParseUint(commandFields[index+1], 10, 64)
		case "movetime":
			moveTime, _ = strconv.ParseUint(commandFields[index+1], 10, 64)
		case colorIndicator + "time":
			remainingTime, _ = strconv.Atoi(commandFields[index+1])
		case colorIndicator + "inc":
			increment, _ = strconv.Atoi(commandFields[index+1])
		}
	}

	uciInterface.gameSearcher.InitializeTimeManager(
		int64(remainingTime),
		int64(increment),
		int64(moveTime),
		int16(movesToGo),
		uint8(depth),
		nodeCount,
	)

	bestMoveEngineResponse := uciInterface.gameSearcher.StartSearch(uciInterface.evaluator)
	fmt.Printf("bestmove %v\n", bestMoveEngineResponse)
}

func (uciInterface *UciInterface) respondToStopCommand() {
	uciInterface.gameSearcher.StopSearch()
}

func (uciInterface *UciInterface) respondToQuitCommand() {
	uciInterface.gameSearcher.CleanUp()
}

func convertUciMoveIntoEncodedMove(position *Position, uciMove string) Move {
	fromSquare := convertSquareNotationToSquareNumber(uciMove[0:2])
	toSquare := convertSquareNotationToSquareNumber(uciMove[2:4])

	pieceType := position.SquareContent[fromSquare].PieceType

	var moveType uint8
	var moveInfo uint8

	uciMoveLastChar := uciMove[len(uciMove)-1]
	if uciMoveLastChar == 'q' || uciMoveLastChar == 'r' || uciMoveLastChar == 'b' || uciMoveLastChar == 'n' {

		moveType = CastleMoveType

		switch uciMoveLastChar {
		case 'q':
			moveInfo = PromotionToQueen
		case 'r':
			moveInfo = PromotionToRook
		case 'b':
			moveInfo = PromotionToBishop
		case 'n':
			moveInfo = PromotionToKnight
		}
	} else if (uciMove == "e1g1" || uciMove == "e8g8" || uciMove == "e1c1" || uciMove == "e8c8") && pieceType == King {
		moveType = CastleMoveType
	} else if position.SquareContent[toSquare].PieceType != NoneType {
		moveType = CaptureMoveType
		if toSquare == position.EnPassantSquare && pieceType == Pawn {
			moveInfo = EnPassant
		}
	} else {
		moveType = QuietMoveType
	}

	return CreateMove(fromSquare, toSquare, moveType, moveInfo)
}

func (uciInterface *UciInterface) Run() {
	uciInterface.respondToUciCommand()
	uciInterface.ReInitialize()

	consoleReader := bufio.NewReader(os.Stdin)

	for {
		userCommand, _ := consoleReader.ReadString('\n')
		command := strings.TrimSpace(strings.Replace(userCommand, "\r\n", "\n", -1))

		if command == "uci" {
			uciInterface.respondToUciCommand()
		} else if strings.HasPrefix(command, "setoption") {
			uciInterface.respondToSetOptionCommand(strings.TrimPrefix(command, "setoption "))
		} else if command == "isready" {
			uciInterface.respondToIsReadyCommand()
		} else if command == "ucinewgame" {
			uciInterface.respondToUciNewGameCommand()
		} else if strings.HasPrefix(command, "position") {
			uciInterface.respondToPositionCommand(strings.TrimPrefix(command, "position "))
		} else if strings.HasPrefix(command, "go") {
			go uciInterface.respondToGoCommand(strings.TrimPrefix(command, "go "))
		} else if command == "stop" {
			uciInterface.respondToStopCommand()
		} else if command == "quit" {
			uciInterface.respondToQuitCommand()
			break
		}
	}
}
