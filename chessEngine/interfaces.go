package chessEngine

type Evaluator interface {
	EvaluatePosition(position *Position) int16
	GetMiddleGamePieceSquareTable() *[6][64]int16
	GetEndGamePieceSquareTable() *[6][64]int16
	GetMiddleGamePieceValues() *[6]int16
	GetEndGamePieceValues() *[6]int16
	GetPhaseValues() *[6]int16
	GetTotalPhaseWeight() int16
}

type EngineOption struct {
	optionType   string
	defaultValue string
	minValue     string
	maxValue     string
	fixedValues  []string
	setOption    func(optionValue string)
}

type GameSearcher interface {
	Reset(evaluator Evaluator)
	ResetToNewGame()
	GetOptions() map[string]EngineOption
	Position() *Position
	RecordPositionHash(positionHash uint64)
	InitializeSearchInfo(fenString string, evaluator Evaluator)
	InitializeTimeManager(remainingTime int64, increment int64, moveTime int64, movesToGo int16, depth uint8, nodeCount uint64)
	StartSearch(evaluator Evaluator) Move
	StopSearch()
	CleanUp()
}
