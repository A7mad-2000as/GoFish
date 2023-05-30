package chessEngine

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

const (
	NumOfKillerMoves                                   = 2
	MaximumNumberOfPlies                               = 1024
	EssentialMovesOffset                        uint16 = math.MaxUint16 - 256
	PvMoveScore                                 uint16 = 65
	KillerMoveFirstSlotScore                    uint16 = 10
	KillerMoveSecondSlotScore                   uint16 = 20
	CounterMoveScore                            uint16 = 5
	HistoryHeuristicScoreUpperBound                    = int32(EssentialMovesOffset - 30)
	AspirationWindowOffset                      int16  = 35
	AspirationWindowMissTimeExtensionLowerBound        = 6
	StaticNullMovePruningPenalty                int16  = 85
	NullMovePruningDepthLimit                   int8   = 2
	RazoringDepthUpperBound                     int8   = 2
	FutilityPruningDepthUpperBound              int8   = 8
	InternalIterativeDeepeningDepthLowerBound   int8   = 4
	InternalIterativeDeepeningReductionAmount   int8   = 2
	LateMovePruningDepthUpperBound              int8   = 5
	FutilityPruningLegalMovesLowerBound         int    = 1
	SingularExtensionDepthLowerBound            int8   = 4
	SingularMoveExtensionPenalty                int16  = 125
	SignularMoveExtensionAmount                 int8   = 1
	LateMoveReductionLegalMoveLowerBound        int    = 4
	LateMoveReductionDepthLowerBound            int8   = 3
)

var FutilityBoosts = [9]int16{0, 100, 160, 220, 280, 340, 400, 460, 520}
var LateMovePruningLegalMoveLowerBounds = [6]int{0, 8, 12, 16, 20, 24}
var LateMoveReductions = [MaxDepth + 1][100]int8{}

var MvvLvaScores [7][6]uint16 = [7][6]uint16{
	{15, 14, 13, 12, 11, 10},
	{25, 24, 23, 22, 21, 20},
	{35, 34, 33, 32, 31, 30},
	{45, 44, 43, 42, 41, 40},
	{55, 54, 53, 52, 51, 50},
	{0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0},
}

type DefaultSearcher struct {
	timeManager                DefaultTimeManager
	position                   Position
	transpositionTable         DefaultTranspositionTable
	searchedNodes              uint64
	positionHashHistory        [MaximumNumberOfPlies]uint64
	positionHashHistoryCounter uint16
	sideToPlay                 uint8
	ageState                   uint8
	killerMoves                [MaxDepth + 1][NumOfKillerMoves]Move
	counterMoves               [2][64][64]Move
	historyHeuristicStats      [2][64][64]int32
}

func InitializeLateMoveReductions() {
	for depth := int8(3); depth < 100; depth++ {
		for legalMovesCounted := int8(3); legalMovesCounted < 100; legalMovesCounted++ {
			LateMoveReductions[depth][legalMovesCounted] = max(2, depth/4) + legalMovesCounted/12
		}
	}
}

func (searcher *DefaultSearcher) GetOptions() map[string]EngineOption {
	options := make(map[string]EngineOption)

	options["Transposition Table Size"] = EngineOption{
		optionType:   "spin",
		defaultValue: "64",
		minValue:     "1",
		maxValue:     "32000",
		setOption: func(sizeValue string) {
			size, err := strconv.Atoi(sizeValue)
			if err != nil {
				searcher.transpositionTable.DeleteEntries()
				searcher.transpositionTable.ResizeTable(uint64(size), EntrySize)
			}
		},
	}

	options["Clear Transposition Table"] = EngineOption{
		optionType: "button",
		setOption: func(_ string) {
			searcher.transpositionTable.ClearEntries()
		},
	}

	options["Clear Killer Moves"] = EngineOption{
		optionType: "button",
		setOption: func(_ string) {
			searcher.ClearKillerMoves()
		},
	}

	options["Clear Counter Moves"] = EngineOption{
		optionType: "button",
		setOption: func(_ string) {
			searcher.ClearCounterMoves()
		},
	}

	options["Clear History Heuristic Stats"] = EngineOption{
		optionType: "button",
		setOption: func(_ string) {
			searcher.ClearHistoryHeuristicStats()
		},
	}

	return options
}

func (searcher *DefaultSearcher) Reset(evaluator Evaluator) {
	*searcher = DefaultSearcher{}
	searcher.transpositionTable.ResizeTable(DefaultTableSize, EntrySize)
	searcher.InitializeSearchInfo(FENStartPosition, evaluator)
}

func (searcher *DefaultSearcher) InitializeSearchInfo(fenString string, evaluator Evaluator) {
	searcher.position.LoadFEN(fenString, evaluator)
	searcher.positionHashHistoryCounter = 0
	searcher.positionHashHistory[searcher.positionHashHistoryCounter] = searcher.position.PositionHash
	searcher.ageState = 0
}

func (searcher *DefaultSearcher) ResetToNewGame() {
	searcher.transpositionTable.ClearEntries()
	searcher.ClearKillerMoves()
	searcher.ClearCounterMoves()
	searcher.ClearHistoryHeuristicStats()
}

func (searcher *DefaultSearcher) Position() *Position {
	return &searcher.position
}

func (searcher *DefaultSearcher) RecordPositionHash(positionHash uint64) {
	searcher.positionHashHistoryCounter++
	searcher.positionHashHistory[searcher.positionHashHistoryCounter] = positionHash
}

func (searcher *DefaultSearcher) EraseLatestPositionHash() {
	searcher.positionHashHistoryCounter--
}

func (searcher *DefaultSearcher) ChangeKillerMoveSlot(ply uint8, killerMove Move) {
	nonCapture := (searcher.position.SquareContent[killerMove.GetToSquare()].PieceType == NoneType)
	if nonCapture {
		if !killerMove.IsSameMove(searcher.killerMoves[ply][0]) {
			searcher.killerMoves[ply][1] = searcher.killerMoves[ply][0]
			searcher.killerMoves[ply][0] = killerMove
		}
	}
}

func (searcher *DefaultSearcher) ClearKillerMoves() {
	for depth := 0; depth < MaxDepth+1; depth++ {
		searcher.killerMoves[depth][0] = NullMove
		searcher.killerMoves[depth][1] = NullMove
	}
}

func (searcher *DefaultSearcher) ClearCounterMoves() {
	for sourceSquare := 0; sourceSquare < 64; sourceSquare++ {
		for destinationSquare := 0; destinationSquare < 64; destinationSquare++ {
			searcher.counterMoves[White][sourceSquare][destinationSquare] = NullMove
			searcher.counterMoves[Black][sourceSquare][destinationSquare] = NullMove
		}
	}
}

func (searcher *DefaultSearcher) ChangeCounterMoveSlot(previousMove Move, counterMove Move) {
	nonCapture := (searcher.position.SquareContent[counterMove.GetToSquare()].PieceType == NoneType)
	if nonCapture {
		searcher.counterMoves[searcher.position.SideToMove][previousMove.GetFromSquare()][previousMove.GetToSquare()] = counterMove
	}
}

func (searcher *DefaultSearcher) ClearHistoryHeuristicStats() {
	for sourceSquare := 0; sourceSquare < 64; sourceSquare++ {
		for destinationSquare := 0; destinationSquare < 64; destinationSquare++ {
			searcher.historyHeuristicStats[searcher.position.SideToMove][sourceSquare][destinationSquare] = 0
		}
	}
}

func (searcher *DefaultSearcher) IncreaseMoveHistoryStrength(move Move, depth int8) {
	nonCapture := (searcher.position.SquareContent[move.GetToSquare()].PieceType == NoneType)

	if nonCapture {
		searcher.historyHeuristicStats[searcher.position.SideToMove][move.GetFromSquare()][move.GetToSquare()] += int32(depth) * int32(depth)
	}

	if searcher.historyHeuristicStats[searcher.position.SideToMove][move.GetFromSquare()][move.GetToSquare()] >= HistoryHeuristicScoreUpperBound {
		searcher.ReduceHistoryHeuristicScores()
	}

}

func (searcher *DefaultSearcher) DecreaseMoveHistoryStrength(move Move) {
	nonCapture := (searcher.position.SquareContent[move.GetToSquare()].PieceType == NoneType)

	if nonCapture && searcher.historyHeuristicStats[searcher.position.SideToMove][move.GetFromSquare()][move.GetToSquare()] > 0 {
		searcher.historyHeuristicStats[searcher.position.SideToMove][move.GetFromSquare()][move.GetToSquare()]--
	}
}

func (searcher *DefaultSearcher) ReduceHistoryHeuristicScores() {
	for sourceSquare := 0; sourceSquare < 64; sourceSquare++ {
		for destinationSquare := 0; destinationSquare < 64; destinationSquare++ {
			searcher.historyHeuristicStats[searcher.position.SideToMove][sourceSquare][destinationSquare] /= 2
		}
	}
}

func (searcher *DefaultSearcher) isThreeFoldRepetition() bool {
	// positionRepetitions := 1

	// for positionHashIndex := uint16(0); positionHashIndex < searcher.positionHashHistoryCounter; positionHashIndex++ {
	// 	if searcher.positionHashHistory[positionHashIndex] == searcher.position.PositionHash {
	// 		positionRepetitions++
	// 		if positionRepetitions == 3 {
	// 			return true
	// 		}
	// 	}
	// }
	// return false

	for repPly := uint16(0); repPly < searcher.positionHashHistoryCounter; repPly++ {
		if searcher.positionHashHistory[repPly] == searcher.position.PositionHash {
			return true
		}
	}
	return false

}

func (searcher *DefaultSearcher) AssignScoresToMoves(moves *MoveList, pvFirstMove Move, depth uint8, previousMove Move) {
	for moveIndex := uint8(0); moveIndex < moves.Size; moveIndex++ {
		move := &moves.Moves[moveIndex]
		pieceToBeMoved := searcher.position.SquareContent[move.GetFromSquare()].PieceType
		pieceToBeCaptured := searcher.position.SquareContent[move.GetToSquare()].PieceType

		if move.IsSameMove(pvFirstMove) {
			move.ModifyMoveScore(EssentialMovesOffset + PvMoveScore)
		} else if pieceToBeCaptured != NoneType {
			move.ModifyMoveScore(EssentialMovesOffset + MvvLvaScores[pieceToBeCaptured][pieceToBeMoved])
		} else if move.IsSameMove(searcher.killerMoves[depth][0]) {
			move.ModifyMoveScore(EssentialMovesOffset - KillerMoveFirstSlotScore)
		} else if move.IsSameMove(searcher.killerMoves[depth][1]) {
			move.ModifyMoveScore(EssentialMovesOffset - KillerMoveSecondSlotScore)
		} else {
			moveScore := uint16(0)
			moveHistoryStrength := uint16(searcher.historyHeuristicStats[searcher.position.SideToMove][move.GetFromSquare()][move.GetToSquare()])
			moveScore += moveHistoryStrength

			if move.IsSameMove(searcher.counterMoves[searcher.position.SideToMove][previousMove.GetFromSquare()][previousMove.GetToSquare()]) {
				moveScore += CounterMoveScore
			}

			move.ModifyMoveScore(moveScore)
		}
	}
}

func OrderHighestScoredMove(destinationIndex uint8, moves *MoveList) {
	highestScore := moves.Moves[destinationIndex].GetScore()
	highestScoreIndex := destinationIndex

	for i := destinationIndex; i < moves.Size; i++ {
		if moves.Moves[i].GetScore() > highestScore {
			highestScore = moves.Moves[i].GetScore()
			highestScoreIndex = i
		}
	}

	temp := moves.Moves[destinationIndex]
	moves.Moves[destinationIndex] = moves.Moves[highestScoreIndex]
	moves.Moves[highestScoreIndex] = temp
}

func (searcher *DefaultSearcher) InitializeTimeManager(remainingTime int64, increment int64, moveTime int64, movesToGo int16, depth uint8, nodeCount uint64) {
	searcher.timeManager.Initialize(remainingTime, increment, moveTime, movesToGo, depth, nodeCount)
}

func getPresentableScore(nodeScore int16) string {
	if nodeScore > MateThreshold || nodeScore < -MateThreshold {
		halfMovesToMate := CheckmateScore - abs(nodeScore)
		fullMovesToMate := (halfMovesToMate / 2) + (halfMovesToMate % 2)
		return fmt.Sprintf("mate %d", fullMovesToMate*(nodeScore/abs(nodeScore)))
	} else {
		return fmt.Sprintf("cp %d", nodeScore)
	}
}

func (searcher *DefaultSearcher) StartSearch(evaluator Evaluator) Move {
	bestMove := NullMove
	pv := PV{}
	searcher.ageState ^= 1
	searcher.sideToPlay = searcher.position.SideToMove
	searcher.searchedNodes = 0
	searchTime := int64(0)
	aspirationWindowMissTimeExtension := false
	alpha := -CheckmateScore
	beta := CheckmateScore

	searcher.ReduceHistoryHeuristicScores()
	searcher.timeManager.StartMoveTimeAllocation(searcher.position.CurrentPly)

	for depth := uint8(1); searcher.timeManager.nodeCount > 0 && depth <= MaxDepth && depth <= searcher.timeManager.depth; depth++ {
		pv.DeleteVariation()

		searchStartInstant := time.Now()
		nodeScore := searcher.Negamax(evaluator, int8(depth), 0, alpha, beta, &pv, true, NullMove, NullMove, false)
		searchDuration := time.Since(searchStartInstant)

		if searcher.timeManager.endSearch {
			if bestMove == NullMove && depth == 1 {
				bestMove = pv.GetVariationFirstMove()
			}
			break
		}

		if nodeScore >= beta || nodeScore <= alpha { // Outside aspiration window
			alpha = -CheckmateScore
			beta = CheckmateScore
			depth--

			if depth >= AspirationWindowMissTimeExtensionLowerBound && !aspirationWindowMissTimeExtension {
				searcher.timeManager.ChangeMoveAllocatedTime(searcher.timeManager.moveAllocatedTime * 13 / 10)
				aspirationWindowMissTimeExtension = true
			}
			continue
		}

		alpha = nodeScore - AspirationWindowOffset
		beta = nodeScore + AspirationWindowOffset

		searchTime += searchDuration.Milliseconds()
		bestMove = pv.GetVariationFirstMove()

		fmt.Printf("info depth %d score %s nodes %d nps %d time %d pv %s\n", depth, getPresentableScore(nodeScore), searcher.searchedNodes, uint64(float64(1000*searcher.searchedNodes)/float64(searchTime)), searchTime, pv)
	}

	return bestMove
}

func (searcher *DefaultSearcher) StopSearch() {
	searcher.timeManager.endSearch = true
}

func (searcher *DefaultSearcher) Negamax(evaluator Evaluator, depth int8, ply uint8, alpha int16, beta int16, pv *PV, nullMovePruningRequired bool, previousMove Move, singularMoveExtensionMove Move, singularMoveExtendedSearch bool) int16 {
	searcher.searchedNodes++

	if ply >= MaxDepth {
		return evaluator.EvaluatePosition(&searcher.position)
	}

	if searcher.searchedNodes >= searcher.timeManager.nodeCount {
		searcher.timeManager.endSearch = true
	}

	if searcher.searchedNodes&2047 == 0 {
		searcher.timeManager.SetMoveTimeIsUp()
	}

	if searcher.timeManager.endSearch {
		return 0
	}

	onTreeRoot := (ply == 0)
	inCheck := searcher.position.IsCurrentSideInCheck()
	isCurrentNodePv := beta-alpha != 1
	continuationPv := PV{}
	futilityPruningPossibility := false
	transpositionTableHashMatch := false

	if inCheck {
		depth++
	}

	if depth <= 0 {
		searcher.searchedNodes--
		return searcher.QuiescenceSearch(evaluator, alpha, beta, ply, pv, ply)
	}

	if !onTreeRoot && ((searcher.position.Rule50 >= 100 && !(inCheck && ply == 1)) || searcher.isThreeFoldRepetition()) {
		return drawScore
	}

	transpostionTableMove := NullMove
	transpostionTableEntry := searcher.transpositionTable.GetEntryToRead(searcher.position.PositionHash)
	transpostionTableScore, transpositionTableScoreValid := transpostionTableEntry.ReadEntryInfo(&transpostionTableMove, searcher.position.PositionHash, ply, uint8(depth), alpha, beta)
	transpositionTableHashMatch = transpostionTableEntry.HashValue == searcher.position.PositionHash

	if transpositionTableScoreValid && !onTreeRoot && !singularMoveExtensionMove.IsSameMove(transpostionTableMove) {
		return transpostionTableScore
	}

	if abs(beta) < MateThreshold && !inCheck && !isCurrentNodePv {
		currentPositionStaticEvaluation := evaluator.EvaluatePosition(&searcher.position)
		penalizedEvaluation := currentPositionStaticEvaluation - StaticNullMovePruningPenalty*int16(depth)
		if penalizedEvaluation >= beta {
			return penalizedEvaluation
		}
	}

	if nullMovePruningRequired && depth >= NullMovePruningDepthLimit && !searcher.position.HasNoMajorOrMinorPieces() && !inCheck && !isCurrentNodePv {
		searcher.position.DoNullMove()
		searcher.RecordPositionHash(searcher.position.PositionHash)

		reductionAmount := 3 + depth/6
		nullMovePruningScore := -searcher.Negamax(evaluator, depth-1-reductionAmount, ply+1, -beta, -beta+1, &continuationPv, false, NullMove, NullMove, singularMoveExtendedSearch)
		searcher.EraseLatestPositionHash()
		searcher.position.unDoPreviousNullMove()
		continuationPv.DeleteVariation()

		if searcher.timeManager.endSearch {
			return 0
		}

		if nullMovePruningScore >= beta && abs(nullMovePruningScore) < MateThreshold {
			return beta
		}
	}

	if depth <= RazoringDepthUpperBound && !inCheck && !isCurrentNodePv {
		currentPositionStaticEvaluation := evaluator.EvaluatePosition(&searcher.position)
		boostedScore := currentPositionStaticEvaluation + FutilityBoosts[depth]*3

		if boostedScore < alpha {
			razoredScore := searcher.QuiescenceSearch(evaluator, alpha, beta, ply, &PV{}, 0)
			if razoredScore < alpha {
				return alpha
			}
		}
	}

	if depth <= FutilityPruningDepthUpperBound && alpha < MateThreshold && beta < MateThreshold && !inCheck && !isCurrentNodePv {
		currentPositionStaticEvaluation := evaluator.EvaluatePosition(&searcher.position)
		boost := FutilityBoosts[depth]
		futilityPruningPossibility = currentPositionStaticEvaluation+boost <= alpha
	}

	if depth >= InternalIterativeDeepeningDepthLowerBound && (isCurrentNodePv || transpostionTableEntry.GetEntryType() == LowerBoundEntryType) && transpostionTableMove.IsSameMove(NullMove) {
		searcher.Negamax(evaluator, depth-InternalIterativeDeepeningReductionAmount-1, ply+1, -beta, -alpha, &continuationPv, true, NullMove, NullMove, singularMoveExtendedSearch)
		if len(continuationPv.moves) > 0 {
			transpostionTableMove = continuationPv.GetVariationFirstMove()
			continuationPv.DeleteVariation()
		}
	}

	pseudoLegalMoves := generatePseudoLegalMoves(&searcher.position)
	searcher.AssignScoresToMoves(&pseudoLegalMoves, transpostionTableMove, ply, previousMove)

	legalMoveCount := 0
	transpositionTableEntryType := UpperBoundEntryType
	highestScore := -CheckmateScore
	bestMove := NullMove

	for i := uint8(0); i < pseudoLegalMoves.Size; i++ {
		OrderHighestScoredMove(i, &pseudoLegalMoves)
		currentMove := pseudoLegalMoves.Moves[i]
		if currentMove.IsSameMove(singularMoveExtensionMove) {
			continue
		}

		if !searcher.position.DoMove(currentMove, evaluator) {
			searcher.position.UnDoPreviousMove(currentMove, evaluator)
			continue
		}

		legalMoveCount++

		if depth <= LateMovePruningDepthUpperBound && legalMoveCount > LateMovePruningLegalMoveLowerBounds[depth] && !inCheck && !isCurrentNodePv {
			if !(searcher.position.IsCurrentSideInCheck() || currentMove.GetMoveType() == PromotionMoveType) {
				searcher.position.UnDoPreviousMove(currentMove, evaluator)
				continue
			}
		}

		if futilityPruningPossibility && legalMoveCount > FutilityPruningLegalMovesLowerBound && currentMove.GetMoveType() != CaptureMoveType && currentMove.GetMoveType() != PromotionMoveType && !searcher.position.IsCurrentSideInCheck() {
			searcher.position.UnDoPreviousMove(currentMove, evaluator)
			continue
		}

		searcher.RecordPositionHash(searcher.position.PositionHash)

		score := int16(0)
		if legalMoveCount == 1 {
			effectiveDepth := depth - 1

			if !singularMoveExtendedSearch &&
				depth >= SingularExtensionDepthLowerBound &&
				transpostionTableMove.IsSameMove(currentMove) &&
				(transpostionTableEntry.GetEntryType() == ExactEntryType || transpostionTableEntry.GetEntryType() == LowerBoundEntryType) &&
				isCurrentNodePv && transpositionTableHashMatch {
				searcher.position.UnDoPreviousMove(currentMove, evaluator)
				searcher.EraseLatestPositionHash()

				penalizedScore := transpostionTableScore - SingularMoveExtensionPenalty
				reductionAmount := 3 + depth/6

				reducedSearchScore := searcher.Negamax(evaluator, depth-1-reductionAmount, ply+1, penalizedScore, penalizedScore+1, &PV{}, true, previousMove, currentMove, true)
				if reducedSearchScore <= penalizedScore {
					effectiveDepth += SignularMoveExtensionAmount
				}

				searcher.position.DoMove(currentMove, evaluator)
				searcher.RecordPositionHash(searcher.position.PositionHash)
			}
			score = -searcher.Negamax(evaluator, effectiveDepth, ply+1, -beta, -alpha, &continuationPv, true, currentMove, NullMove, singularMoveExtendedSearch)
		} else {
			reductionAmount := int8(0)
			if legalMoveCount >= LateMoveReductionLegalMoveLowerBound && depth >= LateMoveReductionDepthLowerBound && !(inCheck || currentMove.GetMoveType() == CaptureMoveType) && !isCurrentNodePv {
				reductionAmount = LateMoveReductions[depth][legalMoveCount]
			}

			score = -searcher.Negamax(evaluator, depth-1-reductionAmount, ply+1, -(alpha + 1), -alpha, &continuationPv, true, currentMove, NullMove, singularMoveExtendedSearch)

			if score > alpha && reductionAmount > 0 {
				score = -searcher.Negamax(evaluator, depth-1, ply+1, -(alpha + 1), -alpha, &continuationPv, true, currentMove, NullMove, singularMoveExtendedSearch)
				if score > alpha {
					score = -searcher.Negamax(evaluator, depth-1, ply+1, -beta, -alpha, &continuationPv, true, currentMove, NullMove, singularMoveExtendedSearch)
				}
			} else if score > alpha && score < beta {
				score = -searcher.Negamax(evaluator, depth-1, ply+1, -beta, -alpha, &continuationPv, true, currentMove, NullMove, singularMoveExtendedSearch)
			}
		}

		searcher.position.UnDoPreviousMove(currentMove, evaluator)
		searcher.EraseLatestPositionHash()

		if score > highestScore {
			highestScore = score
			bestMove = currentMove
		}

		if score >= beta {
			transpositionTableEntryType = LowerBoundEntryType
			searcher.ChangeKillerMoveSlot(ply, currentMove)
			searcher.ChangeCounterMoveSlot(previousMove, currentMove)
			searcher.IncreaseMoveHistoryStrength(currentMove, depth)
			break
		} else {
			searcher.DecreaseMoveHistoryStrength(currentMove)
		}

		if score > alpha {
			alpha = score
			transpositionTableEntryType = ExactEntryType
			pv.SetNewVariation(currentMove, continuationPv)
			searcher.IncreaseMoveHistoryStrength(currentMove, depth)
		} else {
			searcher.DecreaseMoveHistoryStrength(currentMove)
		}
		continuationPv.DeleteVariation()
	}

	if legalMoveCount == 0 {
		if inCheck {
			return -CheckmateScore + int16(ply)
		}
		return drawScore
	}

	if !searcher.timeManager.endSearch {
		tableEntry := searcher.transpositionTable.GetEntryToReplace(searcher.position.PositionHash, uint8(depth), searcher.ageState)
		tableEntry.ModifyTableEntry(bestMove, highestScore, searcher.position.PositionHash, ply, uint8(depth), transpositionTableEntryType, searcher.ageState)
	}

	return highestScore
}

func (searcher *DefaultSearcher) QuiescenceSearch(evaluator Evaluator, alpha int16, beta int16, maximumAllowablePly uint8, pv *PV, ply uint8) int16 {
	searcher.searchedNodes++
	if maximumAllowablePly+ply >= MaxDepth {
		return evaluator.EvaluatePosition(&searcher.position)
	}
	if searcher.searchedNodes >= searcher.timeManager.nodeCount {
		searcher.timeManager.endSearch = true
	}
	if searcher.searchedNodes&2047 == 0 {
		searcher.timeManager.SetMoveTimeIsUp()
	}
	if searcher.timeManager.endSearch {
		return 0
	}

	highestScore := evaluator.EvaluatePosition(&searcher.position)
	inCheck := ply <= 2 && searcher.position.IsCurrentSideInCheck()

	if highestScore >= beta && !inCheck {
		return highestScore
	}

	if highestScore > alpha {
		alpha = highestScore
	}

	pseudoLegalMoves := MoveList{}

	if inCheck {
		pseudoLegalMoves = generatePseudoLegalMoves(&searcher.position)
	} else {
		pseudoLegalMoves = generatePseudoLegalCapturesAndPromotionsToQueens(&searcher.position)
	}

	searcher.AssignScoresToMoves(&pseudoLegalMoves, NullMove, maximumAllowablePly, NullMove)
	continuationPv := PV{}

	for i := uint8(0); i < pseudoLegalMoves.Size; i++ {
		OrderHighestScoredMove(i, &pseudoLegalMoves)
		currentMove := pseudoLegalMoves.Moves[i]
		staticExchangeEvaluationResult := searcher.position.See(currentMove)

		if staticExchangeEvaluationResult < 0 {
			continue
		}

		if !searcher.position.DoMove(currentMove, evaluator) {
			searcher.position.UnDoPreviousMove(currentMove, evaluator)
			continue
		}

		score := -searcher.QuiescenceSearch(evaluator, -beta, -alpha, maximumAllowablePly, &continuationPv, ply+1)
		searcher.position.UnDoPreviousMove(currentMove, evaluator)

		if score > highestScore {
			highestScore = score
		}

		if score >= beta {
			break
		}
		if score > alpha {
			alpha = score
			pv.SetNewVariation(currentMove, continuationPv)
		}

		continuationPv.DeleteVariation()
	}

	return highestScore
}

func (searcher *DefaultSearcher) CleanUp() {
	searcher.transpositionTable.DeleteEntries()
}
