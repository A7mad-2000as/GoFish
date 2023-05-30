package chessEngine

var StandardPieceValuesScaled [7]int16 = [7]int16{
	100,
	300,
	300,
	500,
	900,
	CheckmateScore,
	0,
}

func (position *Position) See(move Move) int16 {
	gain := [32]int16{}
	depth := uint8(0)
	sideToMove := position.SideToMove ^ 1
	toSquare := move.GetToSquare()
	fromSquare := move.GetFromSquare()
	currentAttackerPieceType := position.SquareContent[fromSquare].PieceType
	currentAttackerBitBoard := BitboardForSquare[fromSquare]
	occupiedBitBoard := position.ColorsBitBoard[White] | position.ColorsBitBoard[Black]
	allAttackers := getAllAttackers(toSquare, occupiedBitBoard, position)
	mayUnCoverXrayAttacks := occupiedBitBoard & ^(position.PiecesBitBoard[White][Knight] | position.PiecesBitBoard[White][King] | position.PiecesBitBoard[Black][Knight] | position.PiecesBitBoard[Black][King])
	alreadyChecked := EmptyBitBoard

	gain[depth] = StandardPieceValuesScaled[position.SquareContent[toSquare].PieceType]
	for {
		depth++
		gain[depth] = StandardPieceValuesScaled[currentAttackerPieceType] - gain[depth-1]
		if max(-gain[depth-1], gain[depth]) < 0 { // no matter what, the current player would be at a disadvantage
			break
		}
		allAttackers &= ^currentAttackerBitBoard
		occupiedBitBoard &= ^currentAttackerBitBoard
		alreadyChecked |= currentAttackerBitBoard
		if currentAttackerBitBoard&mayUnCoverXrayAttacks != 0 {
			allAttackers |= getPossibleXrayAttackers(toSquare, occupiedBitBoard, position) & ^alreadyChecked // this works because we did "occupiedBitBoard &= ^currentAttackerBitBoard" so we may have uncovered an X-ray attack
		}
		sideToMove = sideToMove ^ 1
		currentAttackerBitBoard = getLeastValuableAttacker(allAttackers, sideToMove, &currentAttackerPieceType, position)
		if currentAttackerBitBoard == 0 {
			break
		}
	}
	for depth--; depth > 0; depth-- {
		gain[depth-1] = -max(-gain[depth-1], gain[depth])
	}

	return gain[0]
}

func getAllAttackers(toSquare uint8, occupiedBitBoard Bitboard, position *Position) (allAttackers Bitboard) {
	allAttackers |= getAttackersForColor(White, toSquare, occupiedBitBoard, position)
	allAttackers |= getAttackersForColor(Black, toSquare, occupiedBitBoard, position)
	return allAttackers
}
func getAttackersForColor(attackerColor, square uint8, occupiedBitBoard Bitboard, position *Position) (attackers Bitboard) {
	bishopAttacks := GetBishopPseudoLegalMoves(square, occupiedBitBoard) & (position.PiecesBitBoard[attackerColor][Bishop] | position.PiecesBitBoard[attackerColor][Queen])
	rookAttacks := GetRookPseudoLegalMoves(square, occupiedBitBoard) & (position.PiecesBitBoard[attackerColor][Rook] | position.PiecesBitBoard[attackerColor][Queen])
	knightAttacks := ComputedKnightMoves[square] & position.PiecesBitBoard[attackerColor][Knight]
	kingAttacks := ComputedKingMoves[square] & position.PiecesBitBoard[attackerColor][King]
	pawnAttacks := ComputedPawnCaptures[attackerColor^1][square] & position.PiecesBitBoard[attackerColor][Pawn]
	attackers |= bishopAttacks
	attackers |= rookAttacks
	attackers |= knightAttacks
	attackers |= kingAttacks
	attackers |= pawnAttacks
	return attackers
}
func getPossibleXrayAttackers(square uint8, occupiedBitBoard Bitboard, position *Position) (possibleXrayAttackers Bitboard) {
	possibleXrayAttackers |= GetBishopPseudoLegalMoves(square, occupiedBitBoard) & (position.PiecesBitBoard[White][Bishop] | position.PiecesBitBoard[Black][Bishop] | position.PiecesBitBoard[White][Queen] | position.PiecesBitBoard[Black][Queen])
	possibleXrayAttackers |= GetRookPseudoLegalMoves(square, occupiedBitBoard) & (position.PiecesBitBoard[White][Rook] | position.PiecesBitBoard[Black][Rook] | position.PiecesBitBoard[White][Queen] | position.PiecesBitBoard[Black][Queen])
	return possibleXrayAttackers
}
func getLeastValuableAttacker(allAttackers Bitboard, color uint8, currentAttackerPieceType *uint8, position *Position) Bitboard {
	for *currentAttackerPieceType = Pawn; *currentAttackerPieceType <= King; *currentAttackerPieceType++ {
		attackers := allAttackers & position.PiecesBitBoard[color][*currentAttackerPieceType]
		if attackers != 0 {
			return attackers & -attackers // this will return one attackers only
		}
	}
	return 0
}
