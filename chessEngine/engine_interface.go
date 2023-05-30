package chessEngine

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	MaxPerftDepth   = 6
	mainMenuMessage = `
Please enter a command:
- uci : Start the UCI protocol to communicate with the engine
- seeBoardState: Display the current board position
- changePosition <fen>: Change the current position via an FEN string
- perft <x>: Performance test of the move generation to depth x
- dividePerft <x>: Divide performance test of the move generation to depth x
- evaluatePosition: Get the static evaluation of the current position
- exit: Exit the main menu and quit the program`
)

type EngineInterface struct {
	GameSearcher GameSearcher
	Evaluator    Evaluator
}

func NewCustomEngineInterface(GameSearcher GameSearcher, Evaluator Evaluator) EngineInterface {
	ComputePieceMoveTables()
	InitializeZobristHashing()

	return EngineInterface{
		GameSearcher: GameSearcher,
		Evaluator:    Evaluator,
	}
}

func NewDefaultEngineInterface() EngineInterface {
	ComputePieceMoveTables()
	InitializeZobristHashing()
	InitEvaluationRelatedMasks()
	InitializeLateMoveReductions()

	defaultGameSearcher := DefaultSearcher{}
	defaultEvaluator := DefaultEvaluator{}

	return EngineInterface{
		GameSearcher: &defaultGameSearcher,
		Evaluator:    &defaultEvaluator,
	}
}

func reflectFenString(position *Position, fenString string, evaluator Evaluator) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("Invalid FEN String")
			fmt.Println(fenString)
		}
	}()

	position.LoadFEN(fenString, evaluator)
}

func runPerft(perftCommand string, position *Position, evaluator Evaluator) {
	requiredDepth, e := strconv.Atoi(perftCommand)

	if e != nil {
		fmt.Println("Depth is invalid")
	}

	if requiredDepth > MaxPerftDepth {
		fmt.Printf("Max value of depth is %v\n", MaxPerftDepth)
		return
	}

	startTimeInstant := time.Now()
	numberOfVariations := Perft(position, uint8(requiredDepth), evaluator)
	calculationTimeDuration := time.Since(startTimeInstant)

	fmt.Printf("Number of variations: %v\n", numberOfVariations)
	fmt.Printf("Execution time: %vs\n", calculationTimeDuration.Seconds())
}

func runDividePerft(dperftCommand string, position *Position, evaluator Evaluator) {
	requiredDepth, e := strconv.Atoi(dperftCommand)

	if e != nil {
		fmt.Println("Depth is invalid")
	}

	if requiredDepth > MaxPerftDepth {
		fmt.Printf("Max value of depth is %v\n", requiredDepth)
		return
	}

	startTimeInstant := time.Now()
	numberOfVariations := DividePerft(position, uint8(requiredDepth), uint8(requiredDepth), evaluator)
	calculationTimeDuration := time.Since(startTimeInstant)

	fmt.Printf("Number of variations: %v\n", numberOfVariations)
	fmt.Printf("Execution time: %vs\n", calculationTimeDuration.Seconds())
}

func (engineInterface *EngineInterface) StartEngine() {
	consoleReader := bufio.NewReader(os.Stdin)
	uciInterface := UciInterface{
		gameSearcher: engineInterface.GameSearcher,
		evaluator:    engineInterface.Evaluator,
	}

	uciInterface.gameSearcher.InitializeSearchInfo(FENStartPosition, uciInterface.evaluator)
	fmt.Println(mainMenuMessage)

	for {
		userCommand, _ := consoleReader.ReadString('\n')
		command := strings.TrimSpace(strings.Replace(userCommand, "\r\n", "\n", -1))

		if command == "uci" {
			uciInterface.Run()
		} else if command == "seeBoardState" {
			fmt.Println(uciInterface.gameSearcher.Position())
		} else if strings.HasPrefix(command, "changePosition") {
			fenString := strings.TrimPrefix(command, "changePosition ")
			reflectFenString(uciInterface.gameSearcher.Position(), strings.TrimSpace(fenString), uciInterface.evaluator)
		} else if command == "exit" {
			break
		} else if strings.HasPrefix(command, "perft") {
			perftCommand := strings.TrimPrefix(command, "perft ")
			runPerft(perftCommand, uciInterface.gameSearcher.Position(), engineInterface.Evaluator)
		} else if strings.HasPrefix(command, "dividePerft") {
			dividePerftCommand := strings.TrimPrefix(command, "dividePerft ")
			runDividePerft(dividePerftCommand, uciInterface.gameSearcher.Position(), engineInterface.Evaluator)
		} else if command == "evaluatePosition" {
			fmt.Println(uciInterface.evaluator.EvaluatePosition(uciInterface.gameSearcher.Position()))
		} else {
			fmt.Println("Invalid input")
			fmt.Println(mainMenuMessage)
		}
	}
}
