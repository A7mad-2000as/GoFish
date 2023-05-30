package chessEngine

type KingSafetyZone struct {
	OuterDefenseRing Bitboard
	InnerDefenseRing Bitboard
}

var CheckForIsolatedPawnOnFileMasks [8]Bitboard
var CheckDoublePawnOnSquareMask [2][64]Bitboard
var CheckPassedPawnOnSquareMask [2][64]Bitboard
var CheckOutpostOnSquareMask [2][64]Bitboard
var KingSafetyZonesOnSquareMask [64]KingSafetyZone

func InitEvaluationRelatedMasks() {
	for file := FileA; file <= FileH; file++ {
		computeCheckIsolatedPawnOnFileMask(uint8(file))
	}
	for square := 0; square < 64; square++ {
		computeCheckDoublePawnOnSquareMask(uint8(square))
		computeOutpostOnSquareMask(uint8(square))
		computeKingSafetyZonesOnSquareMask(uint8(square))
		computeCheckPassedPawnOnSquareMask(uint8(square))
	}
}

func computeCheckIsolatedPawnOnFileMask(file uint8) {
	var bitboardForFile Bitboard = SetFileMasks[file]
	var CheckIsolatedPawnOnFileMask Bitboard = ((bitboardForFile & ClearFileMasks[FileA]) << 1) | ((bitboardForFile & ClearFileMasks[FileH]) >> 1)
	CheckForIsolatedPawnOnFileMasks[file] = CheckIsolatedPawnOnFileMask
}
func computeCheckDoublePawnOnSquareMask(square uint8) {
	var bitboardForFile Bitboard = SetFileMasks[File(square)]
	squareRank := int(Rank(square))
	var whiteMask Bitboard = bitboardForFile
	for rank := 0; rank <= squareRank; rank++ {
		whiteMask &= ClearRankMasks[rank]
	}
	CheckDoublePawnOnSquareMask[White][square] = whiteMask
	var blackMask Bitboard = bitboardForFile
	for rank := 7; rank >= squareRank; rank-- {
		blackMask &= ClearRankMasks[rank]
	}
	CheckDoublePawnOnSquareMask[Black][square] = blackMask

}
func computeCheckPassedPawnOnSquareMask(square uint8) {
	var currentFileAndAdjacentFilesMask Bitboard = CheckForIsolatedPawnOnFileMasks[File(square)] | SetFileMasks[File(square)]
	squareRank := int(Rank(square))
	whiteFrontSpanMask := currentFileAndAdjacentFilesMask
	for rank := 0; rank <= squareRank; rank++ {
		whiteFrontSpanMask &= ClearRankMasks[rank]
	}
	CheckPassedPawnOnSquareMask[White][square] = whiteFrontSpanMask

	blackFrontSpanMask := currentFileAndAdjacentFilesMask
	for rank := 7; rank >= squareRank; rank-- {
		blackFrontSpanMask &= ClearRankMasks[rank]
	}
	CheckPassedPawnOnSquareMask[Black][square] = blackFrontSpanMask

}
func computeOutpostOnSquareMask(square uint8) {
	var bitboardForFile Bitboard = SetFileMasks[File(square)]
	var currentAdjacentFilesMask Bitboard = CheckForIsolatedPawnOnFileMasks[File(square)]
	squareRank := int(Rank(square))
	whiteMask := currentAdjacentFilesMask
	for rank := 0; rank <= squareRank; rank++ {
		whiteMask &= ClearRankMasks[rank]
	}

	CheckOutpostOnSquareMask[White][square] = whiteMask & ^bitboardForFile

	blackMask := currentAdjacentFilesMask
	for rank := 7; rank >= squareRank; rank-- {
		blackMask &= ClearRankMasks[rank]
	}
	CheckOutpostOnSquareMask[Black][square] = blackMask & ^bitboardForFile
}
func computeKingSafetyZonesOnSquareMask(square uint8) {
	squareBitboard := BitboardForSquare[square]
	var aroundKingZone Bitboard = ((squareBitboard & ClearFileMasks[FileH]) >> 1) | ((squareBitboard & (ClearFileMasks[FileG] & ClearFileMasks[FileH])) >> 2)
	aroundKingZone |= ((squareBitboard & ClearFileMasks[FileA]) << 1) | ((squareBitboard & (ClearFileMasks[FileB] & ClearFileMasks[FileA])) << 2)
	aroundKingZone |= squareBitboard
	aroundKingZone |= (aroundKingZone >> 8) | (aroundKingZone >> 16)
	aroundKingZone |= (aroundKingZone << 8) | (aroundKingZone << 16)

	outerRing := aroundKingZone & ^(ComputedKingMoves[square] | squareBitboard)
	innerRing := ComputedKingMoves[square] | squareBitboard
	KingSafetyZonesOnSquareMask[square] = KingSafetyZone{
		OuterDefenseRing: outerRing,
		InnerDefenseRing: innerRing,
	}
}
