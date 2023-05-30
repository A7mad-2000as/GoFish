package chessEngine

import "fmt"

const (
	WhiteKingSideCastlingSquaresMask  = 0x600000000000000
	WhiteQueenSideCastlingSquaresMask = 0x7000000000000000
	BlackKingSideCastlingSquaresMask  = 0x6
	BlackQueenSideCastlingSquaresMask = 0x70
)

func generatePseudoLegalMoves(currentPosition *Position) (moveList MoveList) {
	for pieceType := uint8(Knight); pieceType < NoneType; pieceType++ {
		pieceBitboard := currentPosition.PiecesBitBoard[currentPosition.SideToMove][pieceType]
		for pieceBitboard != 0 {
			square := pieceBitboard.PopMostSignificantBit()
			generateNonPawnPseudoLegalMoves(currentPosition, pieceType, square, &moveList, FullBitBoard)
		}
	}

	generateKingCastlingPseudoLegalMoves(currentPosition, &moveList)
	generatePawnPseudoLegalMoves(currentPosition, &moveList)

	return moveList
}

func generatePseudoLegalCapturesAndPromotionsToQueens(currentPosition *Position) (moveList MoveList) {
	opponentSideBitboard := currentPosition.ColorsBitBoard[currentPosition.SideToMove^1]

	for pieceType := uint8(Knight); pieceType < NoneType; pieceType++ {
		pieceBitboard := currentPosition.PiecesBitBoard[currentPosition.SideToMove][pieceType]
		for pieceBitboard != 0 {
			square := pieceBitboard.PopMostSignificantBit()
			generateNonPawnPseudoLegalMoves(currentPosition, pieceType, square, &moveList, opponentSideBitboard)
		}
	}

	generatePseudoLegalPawnCapturesAndPromotionsToQueens(currentPosition, &moveList)

	return moveList

}

func generateNonPawnPseudoLegalMoves(currentPosition *Position, pieceType uint8, square uint8, moveList *MoveList, mask Bitboard) {
	sideToPlayBitboard := currentPosition.ColorsBitBoard[currentPosition.SideToMove]
	opponentSideBitboard := currentPosition.ColorsBitBoard[currentPosition.SideToMove^1]

	var pseudoLegalMoves Bitboard
	switch pieceType {
	case King:
		pseudoLegalMoves = ComputedKingMoves[square] & ^sideToPlayBitboard & mask

	case Knight:
		pseudoLegalMoves = ComputedKnightMoves[square] & ^sideToPlayBitboard & mask

	case Bishop:
		pseudoLegalMoves = GetBishopPseudoLegalMoves(square, sideToPlayBitboard|opponentSideBitboard) & ^sideToPlayBitboard & mask

	case Rook:
		pseudoLegalMoves = GetRookPseudoLegalMoves(square, sideToPlayBitboard|opponentSideBitboard) & ^sideToPlayBitboard & mask

	case Queen:
		pseudoLegalBishopLikeMoves := GetBishopPseudoLegalMoves(square, sideToPlayBitboard|opponentSideBitboard)
		pseudoLegalRookLikeMoves := GetRookPseudoLegalMoves(square, sideToPlayBitboard|opponentSideBitboard)
		pseudoLegalMoves = (pseudoLegalBishopLikeMoves | pseudoLegalRookLikeMoves) & ^sideToPlayBitboard & mask
	}

	serializeMove(pseudoLegalMoves, currentPosition, square, opponentSideBitboard, moveList)
}

func GetRookPseudoLegalMoves(square uint8, boardBlockersBitboard Bitboard) Bitboard {
	magicPackage := RookMagicPackages[square]
	blockerMask := boardBlockersBitboard & magicPackage.MaximalBlockerMask
	hashIndex := (uint64(blockerMask) * magicPackage.MagicNumber) >> magicPackage.Shift

	return ComputedRookMoves[square][hashIndex]
}

func GetBishopPseudoLegalMoves(square uint8, boardBlockersBitboard Bitboard) Bitboard {
	magicPackage := BishopMagicPackages[square]
	blockerMask := boardBlockersBitboard & magicPackage.MaximalBlockerMask
	hashIndex := (uint64(blockerMask) * magicPackage.MagicNumber) >> magicPackage.Shift

	return ComputedBishopMoves[square][hashIndex]
}

func serializeMove(moveBitboard Bitboard, currentPosition *Position, square uint8, opponentSideBitboard Bitboard, moveList *MoveList) {

	for moveBitboard != 0 {
		destinationSquare := moveBitboard.PopMostSignificantBit()
		destinationSquareBitboard := BitboardForSquare[destinationSquare]

		if (destinationSquareBitboard & opponentSideBitboard) != 0 {
			moveList.AddMove(CreateMove(square, destinationSquare, CaptureMoveType, uint8(0)))
		} else {
			moveList.AddMove(CreateMove(square, destinationSquare, QuietMoveType, uint8(0)))
		}
	}
}

func generatePawnPseudoLegalMoves(currentPosition *Position, moveList *MoveList) {
	pawnBitboard := currentPosition.PiecesBitBoard[currentPosition.SideToMove][Pawn]
	sideToPlayBitboard := currentPosition.ColorsBitBoard[currentPosition.SideToMove]
	opponentSideBitboard := currentPosition.ColorsBitBoard[currentPosition.SideToMove^1]

	for pawnBitboard != 0 {
		square := pawnBitboard.PopMostSignificantBit()

		pawnSingleAdvances := ComputedPawnAdvances[currentPosition.SideToMove][square] & ^(sideToPlayBitboard | opponentSideBitboard)

		var pawnDoubleAdvances Bitboard
		if currentPosition.SideToMove == White {
			pawnDoubleAdvances = ((pawnSingleAdvances & SetRankMasks[Rank3]) >> NorthOffset) & ^(sideToPlayBitboard | opponentSideBitboard)
		} else {
			pawnDoubleAdvances = ((pawnSingleAdvances & SetRankMasks[Rank6]) << SouthOffset) & ^(sideToPlayBitboard | opponentSideBitboard)
		}

		pawnAdvances := pawnSingleAdvances | pawnDoubleAdvances
		pawnCaptures := ComputedPawnCaptures[currentPosition.SideToMove][square] & (opponentSideBitboard | BitboardForSquare[currentPosition.EnPassantSquare])

		for pawnAdvances != 0 {
			destinationSquare := pawnAdvances.PopMostSignificantBit()
			if BitboardForSquare[destinationSquare]&(SetRankMasks[Rank1]|SetRankMasks[Rank8]) != 0 {
				moveList.AddMove(CreateMove(square, destinationSquare, PromotionMoveType, PromotionToQueen))
				moveList.AddMove(CreateMove(square, destinationSquare, PromotionMoveType, PromotionToRook))
				moveList.AddMove(CreateMove(square, destinationSquare, PromotionMoveType, PromotionToBishop))
				moveList.AddMove(CreateMove(square, destinationSquare, PromotionMoveType, PromotionToKnight))
				continue
			}
			moveList.AddMove(CreateMove(square, destinationSquare, QuietMoveType, uint8(0)))
		}

		for pawnCaptures != 0 {
			destinationSquare := pawnCaptures.PopMostSignificantBit()
			if BitboardForSquare[destinationSquare]&(SetRankMasks[Rank1]|SetRankMasks[Rank8]) != 0 {
				moveList.AddMove(CreateMove(square, destinationSquare, PromotionMoveType, PromotionToQueen))
				moveList.AddMove(CreateMove(square, destinationSquare, PromotionMoveType, PromotionToRook))
				moveList.AddMove(CreateMove(square, destinationSquare, PromotionMoveType, PromotionToBishop))
				moveList.AddMove(CreateMove(square, destinationSquare, PromotionMoveType, PromotionToKnight))
				continue
			}
			if destinationSquare == currentPosition.EnPassantSquare {
				moveList.AddMove(CreateMove(square, destinationSquare, CaptureMoveType, EnPassant))
			} else {
				moveList.AddMove(CreateMove(square, destinationSquare, CaptureMoveType, uint8(0)))
			}
		}
	}
}

func generatePseudoLegalPawnCapturesAndPromotionsToQueens(currentPosition *Position, moveList *MoveList) {
	pawnBitboard := currentPosition.PiecesBitBoard[currentPosition.SideToMove][Pawn]
	sideToPlayBitboard := currentPosition.ColorsBitBoard[currentPosition.SideToMove]
	opponentSideBitboard := currentPosition.ColorsBitBoard[currentPosition.SideToMove^1] | BitboardForSquare[currentPosition.EnPassantSquare]

	for pawnBitboard != 0 {
		square := pawnBitboard.PopMostSignificantBit()
		pawnSingleAdvances := ComputedPawnAdvances[currentPosition.SideToMove][square] & ^(opponentSideBitboard | sideToPlayBitboard)
		advanceDestinationSquare := pawnSingleAdvances.PopMostSignificantBit()
		if (BitboardForSquare[advanceDestinationSquare] & (SetRankMasks[Rank1] | SetRankMasks[Rank8])) != 0 {
			moveList.AddMove(CreateMove(square, advanceDestinationSquare, PromotionMoveType, PromotionToQueen))
		}

		pawnCaptures := ComputedPawnCaptures[currentPosition.SideToMove][square] & opponentSideBitboard

		for pawnCaptures != 0 {
			captureDestinationSquare := pawnCaptures.PopMostSignificantBit()

			if (BitboardForSquare[captureDestinationSquare] & (SetRankMasks[Rank1] | SetRankMasks[Rank8])) != 0 {
				moveList.AddMove(CreateMove(square, captureDestinationSquare, PromotionMoveType, PromotionToQueen))
				continue
			}

			if captureDestinationSquare == currentPosition.EnPassantSquare {
				moveList.AddMove(CreateMove(square, captureDestinationSquare, CaptureMoveType, EnPassant))
			} else {
				moveList.AddMove(CreateMove(square, captureDestinationSquare, CaptureMoveType, uint8(0)))
			}
		}
	}
}

func generateKingCastlingPseudoLegalMoves(currentPosition *Position, moveList *MoveList) {
	occupancyBitboard := currentPosition.ColorsBitBoard[currentPosition.SideToMove] | currentPosition.ColorsBitBoard[currentPosition.SideToMove^1]

	if currentPosition.SideToMove == White {
		if (occupancyBitboard&WhiteKingSideCastlingSquaresMask) == 0 &&
			(!squaresAreCoveredByOpponent([]uint8{E1, F1, G1}, currentPosition, occupancyBitboard)) &&
			(currentPosition.CastlingRights&White_Kingside_Castle_Right) != 0 {
			moveList.AddMove(CreateMove(E1, G1, CastleMoveType, uint8(0)))
		}

		if (occupancyBitboard&WhiteQueenSideCastlingSquaresMask) == 0 &&
			(!squaresAreCoveredByOpponent([]uint8{C1, D1, E1}, currentPosition, occupancyBitboard)) &&
			(currentPosition.CastlingRights&White_Queenside_Castle_Right) != 0 {
			moveList.AddMove(CreateMove(E1, C1, CastleMoveType, uint8(0)))
		}
	} else {
		if (occupancyBitboard&BlackKingSideCastlingSquaresMask) == 0 &&
			(!squaresAreCoveredByOpponent([]uint8{E8, F8, G8}, currentPosition, occupancyBitboard)) &&
			(currentPosition.CastlingRights&Black_Kingside_Castle_Right) != 0 {
			moveList.AddMove(CreateMove(E8, G8, CastleMoveType, uint8(0)))
		}

		if (occupancyBitboard&BlackQueenSideCastlingSquaresMask) == 0 &&
			(!squaresAreCoveredByOpponent([]uint8{C8, D8, E8}, currentPosition, occupancyBitboard)) &&
			(currentPosition.CastlingRights&Black_Queenside_Castle_Right) != 0 {
			moveList.AddMove(CreateMove(E8, C8, CastleMoveType, uint8(0)))
		}
	}
}

func squaresAreCoveredByOpponent(squares []uint8, currentPosition *Position, occupancyBitboard Bitboard) bool {
	opponentQueens := currentPosition.PiecesBitBoard[currentPosition.SideToMove^1][Queen]
	opponentRooks := currentPosition.PiecesBitBoard[currentPosition.SideToMove^1][Rook]
	opponentKnights := currentPosition.PiecesBitBoard[currentPosition.SideToMove^1][Knight]
	opponentBishops := currentPosition.PiecesBitBoard[currentPosition.SideToMove^1][Bishop]
	opponentPawns := currentPosition.PiecesBitBoard[currentPosition.SideToMove^1][Pawn]
	opponentKing := currentPosition.PiecesBitBoard[currentPosition.SideToMove^1][King]

	for _, square := range squares {
		coveredDiagonalSquares := GetBishopPseudoLegalMoves(square, occupancyBitboard)
		coveredOrthogonalSquares := GetRookPseudoLegalMoves(square, occupancyBitboard)
		coveredKnightDistanceSquares := ComputedKnightMoves[square]
		coveredKingMovesSquares := ComputedKingMoves[square]
		coveredPawnCapturesSquares := ComputedPawnCaptures[currentPosition.SideToMove][square]

		if (coveredDiagonalSquares&(opponentQueens|opponentBishops)) != 0 ||
			(coveredOrthogonalSquares&(opponentQueens|opponentRooks)) != 0 ||
			(coveredKnightDistanceSquares&opponentKnights) != 0 ||
			(coveredKingMovesSquares&opponentKing) != 0 ||
			(coveredPawnCapturesSquares&opponentPawns) != 0 {
			return true
		}
	}

	return false
}

func DividePerft(currentPosition *Position, depth uint8, divisionPoint uint8, evaluator Evaluator) uint64 {
	if depth == 0 {
		return 1
	}

	pseudoLegalMoves := generatePseudoLegalMoves(currentPosition)
	totalVariations := uint64(0)
	for i := uint8(0); i < pseudoLegalMoves.Size; i++ {
		pseudoLegalMove := pseudoLegalMoves.Moves[i]

		if currentPosition.DoMove(pseudoLegalMove, evaluator) {
			variationsUnderNode := DividePerft(currentPosition, depth-1, divisionPoint, evaluator)
			if depth == divisionPoint {
				fmt.Printf("%v: %v\n", pseudoLegalMove, variationsUnderNode)
			}
			totalVariations += variationsUnderNode
		}

		currentPosition.UnDoPreviousMove(pseudoLegalMove, evaluator)
	}

	return totalVariations
}

func Perft(currentPosition *Position, depth uint8, evaluator Evaluator) uint64 {
	if depth == 0 {
		return 1
	}

	pseudoLegalMoves := generatePseudoLegalMoves(currentPosition)
	totalVariations := uint64(0)

	for i := uint8(0); i < pseudoLegalMoves.Size; i++ {
		pseudoLegalMove := pseudoLegalMoves.Moves[i]

		if currentPosition.DoMove(pseudoLegalMove, evaluator) {
			variationsUnderNode := Perft(currentPosition, depth-1, evaluator)
			totalVariations += variationsUnderNode
		}

		currentPosition.UnDoPreviousMove(pseudoLegalMove, evaluator)
	}

	return totalVariations
}
