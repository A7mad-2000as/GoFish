package chessEngine

const (
	CheckmateScore             int16 = 10000
	drawScore                  int16 = 0
	DrawishPositionScaleFactor int16 = 16
)

type DefaultEvaluator struct {
	evaluationData EvaluationData
}

type EvaluationData struct {
	MidgameScores           [2]int16
	EndgameScores           [2]int16
	ThreatToEnemyKingPoints [2]uint16
	EnemyKingAttackerCount  [2]uint8
}

func (evaluator *DefaultEvaluator) GetMiddleGamePieceSquareTable() *[6][64]int16 {
	return &MidGamePieceSquareTables
}

func (evaluator *DefaultEvaluator) GetEndGamePieceSquareTable() *[6][64]int16 {
	return &EndGamePieceSquareTables
}

func (evaluator *DefaultEvaluator) GetMiddleGamePieceValues() *[6]int16 {
	return &MidGamePieceValues
}

func (evaluator *DefaultEvaluator) GetEndGamePieceValues() *[6]int16 {
	return &EndGamePieceValues
}

func (evaluator *DefaultEvaluator) GetPhaseValues() *[6]int16 {
	return &PiecePhaseIncrements
}

func (evaluator *DefaultEvaluator) GetTotalPhaseWeight() int16 {
	return TotalPhaseIncrement
}

func (defaultClassicEvaluator *DefaultEvaluator) EvaluatePosition(position *Position) int16 {
	if isDrawnState(position) {
		return drawScore
	}
	allBitBoard := position.ColorsBitBoard[position.SideToMove] | position.ColorsBitBoard[position.SideToMove^1]
	var phaseValue = position.Phase
	defaultClassicEvaluator.evaluationData = EvaluationData{
		MidgameScores: position.MidGameScores,
		EndgameScores: position.EndGameScores,
	}
	for allBitBoard != 0 {
		pieceSquare := allBitBoard.PopMostSignificantBit()
		pieceType := position.SquareContent[pieceSquare].PieceType
		pieceColor := position.SquareContent[pieceSquare].Color
		switch pieceType {
		case Pawn:
			defaultClassicEvaluator.evaluatePawnAtSquare(position, pieceColor, pieceSquare)
		case Knight:
			defaultClassicEvaluator.evaluateKnightAtSquare(position, pieceColor, pieceSquare)
		case Bishop:
			defaultClassicEvaluator.evaluateBishopAtSquare(position, pieceColor, pieceSquare)
		case Rook:
			defaultClassicEvaluator.evaluateRookAtSquare(position, pieceColor, pieceSquare)
		case Queen:
			defaultClassicEvaluator.evaluateQueenAtSquare(position, pieceColor, pieceSquare)
		}
	}
	for color := Black; color <= White; color++ {
		if position.PiecesBitBoard[color][Bishop].CountSetBits() >= 2 {
			defaultClassicEvaluator.evaluationData.MidgameScores[color] += MidGameBishopPairBonus
			defaultClassicEvaluator.evaluationData.EndgameScores[color] += EndgameBishopPairBonus
		}
		defaultClassicEvaluator.evaluateKingAtSquare(position, color, position.PiecesBitBoard[color][King].MostSignificantBit())
	}
	defaultClassicEvaluator.evaluationData.MidgameScores[position.SideToMove] += MidGameTempoBonus

	currentMidGameScore := defaultClassicEvaluator.evaluationData.MidgameScores[position.SideToMove] - defaultClassicEvaluator.evaluationData.MidgameScores[position.SideToMove^1]
	currentEndGameScore := defaultClassicEvaluator.evaluationData.EndgameScores[position.SideToMove] - defaultClassicEvaluator.evaluationData.EndgameScores[position.SideToMove^1]

	scaledPhaseValue := (phaseValue*256 + (TotalPhaseIncrement / 2)) / TotalPhaseIncrement
	currentScore := int16(((int32(currentMidGameScore) * (int32(256) - int32(scaledPhaseValue))) + (int32(currentEndGameScore) * int32(scaledPhaseValue))) / int32(256))

	if isDrawishState(position) {
		return currentScore / DrawishPositionScaleFactor
	}

	return currentScore
}
func (defaultClassicEvaluator *DefaultEvaluator) evaluatePawnAtSquare(position *Position, color uint8, square uint8) {
	enemyPawns := position.PiecesBitBoard[color^1][Pawn]
	sideToMovePawn := position.PiecesBitBoard[color][Pawn]
	fileOfSq := File(square)
	isIsolated := CheckForIsolatedPawnOnFileMasks[fileOfSq]&sideToMovePawn == 0
	isDoubled := CheckDoublePawnOnSquareMask[color][square]&sideToMovePawn != 0
	isPassedAndNotBlockedByFriendlyPawn := CheckPassedPawnOnSquareMask[color][square]&enemyPawns == 0 && sideToMovePawn&CheckDoublePawnOnSquareMask[color][square] == 0

	if isIsolated {
		defaultClassicEvaluator.evaluationData.MidgameScores[color] -= MidGameIsolatedPawnPenalty
		defaultClassicEvaluator.evaluationData.EndgameScores[color] -= EndGameIsolatedPawnPenalty
	}
	if isDoubled {
		defaultClassicEvaluator.evaluationData.MidgameScores[color] -= MidGameDoubledPawnPenalty
		defaultClassicEvaluator.evaluationData.EndgameScores[color] -= EndGameDoubledPawnPenalty
	}
	if isPassedAndNotBlockedByFriendlyPawn {
		defaultClassicEvaluator.evaluationData.MidgameScores[color] += MidGamePassedPawnSquareTables[BoardSquaresNormalAndFlipped[color][square]]
		defaultClassicEvaluator.evaluationData.EndgameScores[color] += EndGamePassedPawnSquareTables[BoardSquaresNormalAndFlipped[color][square]]
	}
}
func (defaultClassicEvaluator *DefaultEvaluator) evaluateKnightAtSquare(position *Position, color uint8, square uint8) {
	var enemyPawns Bitboard = position.PiecesBitBoard[color^1][Pawn]
	var sideToMovePawns Bitboard = position.PiecesBitBoard[color][Pawn]

	// outPost Evaluation
	noEnemyCanAttackKnight := CheckOutpostOnSquareMask[color][square]&enemyPawns == 0
	isTheKnightProtectedByFriendlyPawn := ComputedPawnCaptures[color^1][square]&sideToMovePawns != 0
	if noEnemyCanAttackKnight && isTheKnightProtectedByFriendlyPawn &&
		BoardRanksNormalAndFlipped[color][Rank(square)] >= Rank5 {
		defaultClassicEvaluator.evaluationData.MidgameScores[color] += MidGameKnightOnOutpostBonus
		defaultClassicEvaluator.evaluationData.EndgameScores[color] += EndGameKnightOnOutpostBonus
	}
	// mobility evaluation
	var sideToMoveBitBoard Bitboard = position.ColorsBitBoard[color]
	var knightMoves Bitboard = ComputedKnightMoves[square] & ^sideToMoveBitBoard
	var knightSafeMoves Bitboard = filterMoveAndKeepTheSafeMoves(knightMoves, color, enemyPawns)
	mobility := int16(knightSafeMoves.CountSetBits())
	defaultClassicEvaluator.evaluationData.MidgameScores[color] += (mobility - 4) * MidGameMobilityScoresPerPiece[Knight]
	defaultClassicEvaluator.evaluationData.EndgameScores[color] += (mobility - 4) * EndGameMobilityScoresPerPiece[Knight]

	// attacks on enemy king evaluation
	defaultClassicEvaluator.evaluateAttacksOnEnemyKing(position, knightSafeMoves, color, Knight)
}
func (defaultClassicEvaluator *DefaultEvaluator) evaluateBishopAtSquare(position *Position, color uint8, square uint8) {
	enemyPawns := position.PiecesBitBoard[color^1][Pawn]
	sideToMovePawns := position.PiecesBitBoard[color][Pawn]
	sideToMoveBitBoard := position.ColorsBitBoard[color]
	allBitBoard := position.ColorsBitBoard[position.SideToMove] | position.ColorsBitBoard[position.SideToMove^1]
	// outPost Evaluation
	noEnemyCanAttackBishop := CheckOutpostOnSquareMask[color][square]&enemyPawns == 0
	isTheBishopProtectedByFriendlyPawn := ComputedPawnCaptures[color^1][square]&sideToMovePawns != 0
	if noEnemyCanAttackBishop && isTheBishopProtectedByFriendlyPawn &&
		BoardRanksNormalAndFlipped[color][Rank(square)] >= Rank5 {
		defaultClassicEvaluator.evaluationData.MidgameScores[color] += MidGameBishopOnOutpostBonus
		defaultClassicEvaluator.evaluationData.EndgameScores[color] += EndGameBishopOnOutpostBonus
	}

	//mobility evaluation
	var bishopMoves Bitboard = GetBishopPseudoLegalMoves(square, allBitBoard) & ^sideToMoveBitBoard
	mobility := int16(bishopMoves.CountSetBits())
	defaultClassicEvaluator.evaluationData.MidgameScores[color] += (mobility - 7) * MidGameMobilityScoresPerPiece[Bishop]
	defaultClassicEvaluator.evaluationData.EndgameScores[color] += (mobility - 7) * EndGameMobilityScoresPerPiece[Bishop]

	// attacks on enemy king evaluation
	defaultClassicEvaluator.evaluateAttacksOnEnemyKing(position, bishopMoves, color, Bishop)

}
func (defaultClassicEvaluator *DefaultEvaluator) evaluateRookAtSquare(position *Position, color uint8, square uint8) {
	enemyKingSquare := position.PiecesBitBoard[color^1][King].MostSignificantBit()
	allPawns := position.PiecesBitBoard[color][Pawn] | position.PiecesBitBoard[color^1][Pawn]
	sideToMoveBitBoard := position.ColorsBitBoard[color]
	allBitBoard := position.ColorsBitBoard[color] | position.ColorsBitBoard[color^1]

	if BoardRanksNormalAndFlipped[color][Rank(square)] == Rank7 && BoardRanksNormalAndFlipped[color][Rank(enemyKingSquare)] >= Rank7 {
		defaultClassicEvaluator.evaluationData.EndgameScores[color] += EndGameBonusForRookOrQueenOnSeventhRank
	}

	if SetFileMasks[File(square)]&allPawns == 0 {
		defaultClassicEvaluator.evaluationData.MidgameScores[color] += MidGameRookOnOpenFileBonus
	}

	rookMoves := GetRookPseudoLegalMoves(square, allBitBoard) & ^sideToMoveBitBoard
	mobility := int16(rookMoves.CountSetBits())
	defaultClassicEvaluator.evaluationData.MidgameScores[color] += (mobility - 7) * MidGameMobilityScoresPerPiece[Rook]
	defaultClassicEvaluator.evaluationData.EndgameScores[color] += (mobility - 7) * EndGameMobilityScoresPerPiece[Rook]
	defaultClassicEvaluator.evaluateAttacksOnEnemyKing(position, rookMoves, color, Rook)

}
func (defaultClassicEvaluator *DefaultEvaluator) evaluateQueenAtSquare(position *Position, color uint8, square uint8) {
	enemyKingSquare := position.PiecesBitBoard[color^1][King].MostSignificantBit()
	sideToMoveBitBoard := position.ColorsBitBoard[color]
	allBitBoard := position.ColorsBitBoard[color] | position.ColorsBitBoard[color^1]

	if BoardRanksNormalAndFlipped[color][Rank(square)] == Rank7 && BoardRanksNormalAndFlipped[color][Rank(enemyKingSquare)] >= Rank7 {
		defaultClassicEvaluator.evaluationData.EndgameScores[color] += EndGameBonusForRookOrQueenOnSeventhRank
	}
	queenMoves := (GetBishopPseudoLegalMoves(square, allBitBoard) | GetRookPseudoLegalMoves(square, allBitBoard)) & ^sideToMoveBitBoard
	mobility := int16(queenMoves.CountSetBits())

	defaultClassicEvaluator.evaluationData.MidgameScores[color] += (mobility - 14) * MidGameMobilityScoresPerPiece[Queen]
	defaultClassicEvaluator.evaluationData.EndgameScores[color] += (mobility - 14) * EndGameMobilityScoresPerPiece[Queen]

	defaultClassicEvaluator.evaluateAttacksOnEnemyKing(position, queenMoves, color, Queen)
}

func (defaultClassicEvaluator *DefaultEvaluator) evaluateKingAtSquare(position *Position, color uint8, square uint8) {
	threatPointOnSideToMoveKing := defaultClassicEvaluator.evaluationData.ThreatToEnemyKingPoints[color^1]
	kingFile := SetFileMasks[File(square)]
	kingLeftFile, kingRightFile := ((kingFile & ClearFileMasks[FileA]) << 1), ((kingFile & ClearFileMasks[FileH]) >> 1)
	sideToMovePawns := position.PiecesBitBoard[color][Pawn]

	var semipOpenFilePenality uint16 = 0
	if kingFile&sideToMovePawns == 0 {
		semipOpenFilePenality += uint16(SemiOpenFileBesideKingPenalty)
	}
	if kingLeftFile != 0 && kingLeftFile&sideToMovePawns == 0 {
		semipOpenFilePenality += uint16(SemiOpenFileBesideKingPenalty)
	}
	if kingRightFile != 0 && kingRightFile&sideToMovePawns == 0 {
		semipOpenFilePenality += uint16(SemiOpenFileBesideKingPenalty)
	}

	finalPenalty := int16(((threatPointOnSideToMoveKing + semipOpenFilePenality) * (threatPointOnSideToMoveKing + semipOpenFilePenality)) / 4)
	if defaultClassicEvaluator.evaluationData.EnemyKingAttackerCount[color^1] >= 2 && position.PiecesBitBoard[color^1][Queen] != 0 {
		defaultClassicEvaluator.evaluationData.MidgameScores[color] -= finalPenalty
	}
}

func (defaultClassicEvaluator *DefaultEvaluator) evaluateAttacksOnEnemyKing(position *Position, moves Bitboard, color uint8, piece uint8) {
	var attacksOnEnemyKingOuterRing Bitboard = moves & KingSafetyZonesOnSquareMask[position.PiecesBitBoard[color^1][King].MostSignificantBit()].OuterDefenseRing
	var attacksOnEnemyKingInnerRing Bitboard = moves & KingSafetyZonesOnSquareMask[position.PiecesBitBoard[color^1][King].MostSignificantBit()].InnerDefenseRing
	if attacksOnEnemyKingOuterRing != 0 || attacksOnEnemyKingInnerRing != 0 {
		defaultClassicEvaluator.evaluationData.EnemyKingAttackerCount[color]++
		defaultClassicEvaluator.evaluationData.ThreatToEnemyKingPoints[color] += uint16(attacksOnEnemyKingOuterRing.CountSetBits()) * uint16(OuterRingAttackScorePerPiece[piece])
		defaultClassicEvaluator.evaluationData.ThreatToEnemyKingPoints[color] += uint16(attacksOnEnemyKingInnerRing.CountSetBits()) * uint16(InnerRingAttackScorePerPiece[piece])
	}
}

func filterMoveAndKeepTheSafeMoves(moves Bitboard, color uint8, enemyPawns Bitboard) Bitboard {
	safeMoves := moves
	for enemyPawns != 0 {
		square := enemyPawns.PopMostSignificantBit()
		safeMoves &= ^ComputedPawnCaptures[color^1][square]
	}
	return safeMoves
}

func isDrawnState(position *Position) bool {
	whiteKnightCount := position.PiecesBitBoard[White][Knight].CountSetBits()
	whiteBishopCount := position.PiecesBitBoard[White][Bishop].CountSetBits()

	blackKnightCount := position.PiecesBitBoard[Black][Knight].CountSetBits()
	blackBishopCount := position.PiecesBitBoard[Black][Bishop].CountSetBits()

	totalPawnsCount := position.PiecesBitBoard[White][Pawn].CountSetBits() + position.PiecesBitBoard[Black][Pawn].CountSetBits()
	totalKnightsCount := whiteKnightCount + blackKnightCount
	totalBishopsCount := whiteBishopCount + blackBishopCount
	totalRooksCount := position.PiecesBitBoard[White][Rook].CountSetBits() + position.PiecesBitBoard[Black][Rook].CountSetBits()
	totalQueensCount := position.PiecesBitBoard[White][Queen].CountSetBits() + position.PiecesBitBoard[Black][Queen].CountSetBits()

	majorPiecesCount := totalRooksCount + totalQueensCount
	minorPiecesCount := totalKnightsCount + totalBishopsCount

	if totalPawnsCount+majorPiecesCount+minorPiecesCount == 0 {
		// King vs King
		return true
	} else if majorPiecesCount+totalPawnsCount == 0 {
		if minorPiecesCount == 1 {
			// King & minorPiece vs King
			return true
		} else if minorPiecesCount == 2 && whiteKnightCount == 1 && blackKnightCount == 1 {
			// King & Knight vs King & Knight
			return true
		} else if minorPiecesCount == 2 && whiteBishopCount == 1 && blackBishopCount == 1 {
			// King & Bishop vs King & Bishop and bishops are on the same square color
			return isSquareLight(position.PiecesBitBoard[White][Bishop].MostSignificantBit()) == isSquareLight(position.PiecesBitBoard[Black][Bishop].MostSignificantBit())
		}
	}

	return false
}

func isDrawishState(pos *Position) bool {
	whiteKnightCount := pos.PiecesBitBoard[White][Knight].CountSetBits()
	whiteBishopCount := pos.PiecesBitBoard[White][Bishop].CountSetBits()
	whiteRookCount := pos.PiecesBitBoard[White][Rook].CountSetBits()
	whiteQueenCount := pos.PiecesBitBoard[White][Queen].CountSetBits()

	blackKnightCount := pos.PiecesBitBoard[Black][Knight].CountSetBits()
	blackBishopCount := pos.PiecesBitBoard[Black][Bishop].CountSetBits()
	blackRookCount := pos.PiecesBitBoard[Black][Rook].CountSetBits()
	blackQueenCount := pos.PiecesBitBoard[Black][Queen].CountSetBits()

	totalPawnsCount := pos.PiecesBitBoard[White][Pawn].CountSetBits() + pos.PiecesBitBoard[Black][Pawn].CountSetBits()
	totalKnightsCount := whiteKnightCount + blackKnightCount
	totalBishopsCount := whiteBishopCount + blackKnightCount
	totalRooksCount := whiteRookCount + blackRookCount
	totalQueensCount := whiteQueenCount + blackQueenCount

	whiteMinorPiecesCount := whiteBishopCount + whiteKnightCount
	blackMinorPiecesCount := blackBishopCount + blackKnightCount

	totalMajorPiecesCount := totalRooksCount + totalQueensCount
	totalMinorPiecesCount := totalKnightsCount + totalBishopsCount
	totalPiecesCount := totalMajorPiecesCount + totalMinorPiecesCount

	if totalPawnsCount == 0 {
		if totalPiecesCount == 2 && blackQueenCount == 1 && whiteQueenCount == 1 {
			// KQ v KQ
			return true
		} else if totalPiecesCount == 2 && blackRookCount == 1 && whiteRookCount == 1 {
			// KR v KR
			return true
		} else if totalPiecesCount == 2 && whiteMinorPiecesCount == 1 && blackMinorPiecesCount == 1 {
			// KN v KB
			// KB v KB
			return true
		} else if totalPiecesCount == 3 && ((whiteQueenCount == 1 && blackRookCount == 2) || (blackQueenCount == 1 && whiteRookCount == 2)) {
			// KQ v KRR
			return true
		} else if totalPiecesCount == 3 && ((whiteQueenCount == 1 && blackBishopCount == 2) || (blackQueenCount == 1 && whiteBishopCount == 2)) {
			// KQ vs KBB
			return true
		} else if totalPiecesCount == 3 && ((whiteQueenCount == 1 && blackKnightCount == 2) || (blackQueenCount == 1 && whiteKnightCount == 2)) {
			// KQ vs KNN
			return true
		} else if totalPiecesCount <= 3 && ((whiteKnightCount == 2 && blackMinorPiecesCount <= 1) || (blackKnightCount == 2 && whiteMinorPiecesCount <= 1)) {
			// KNN v KN, KNN v KB, KNN v K
			return true
		} else if totalPiecesCount == 3 &&
			((whiteQueenCount == 1 && blackRookCount == 1 && blackMinorPiecesCount == 1) ||
				(blackQueenCount == 1 && whiteRookCount == 1 && whiteMinorPiecesCount == 1)) {
			// KQ vs KRN, KQ vs KRB
			return true
		} else if totalPiecesCount == 3 &&
			((whiteRookCount == 1 && blackRookCount == 1 && blackMinorPiecesCount == 1) ||
				(blackRookCount == 1 && whiteRookCount == 1 && whiteMinorPiecesCount == 1)) {
			// KR vs KRB, KR vs KRN
		} else if totalPiecesCount == 4 &&
			((whiteRookCount == 2 && blackRookCount == 1 && blackMinorPiecesCount == 1) ||
				(blackRookCount == 2 && whiteRookCount == 1 && whiteMinorPiecesCount == 1)) {
			// KRR v KRB, KRR v KRN
			return true
		}
	}

	return false
}

func isSquareLight(square uint8) bool {
	fileNumber := File(square)
	rankNumber := File(square)
	return (fileNumber+rankNumber)%2 != 0
}
