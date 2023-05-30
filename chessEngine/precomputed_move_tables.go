package chessEngine

const (
	FileA = 0
	FileB = 1
	FileC = 2
	FileD = 3
	FileE = 4
	FileF = 5
	FileG = 6
	FileH = 7

	Rank1 = 0
	Rank2 = 1
	Rank3 = 2
	Rank4 = 3
	Rank5 = 4
	Rank6 = 5
	Rank7 = 6
	Rank8 = 7

	NorthOffset = 8
	SouthOffset = 8
	EastOffset  = 1
	WestOffset  = 1
)

var MainDiagonalMasks = [15]Bitboard{
	0x80,
	0x8040,
	0x804020,
	0x80402010,
	0x8040201008,
	0x804020100804,
	0x80402010080402,
	0x8040201008040201,
	0x4020100804020100,
	0x2010080402010000,
	0x1008040201000000,
	0x804020100000000,
	0x402010000000000,
	0x201000000000000,
	0x100000000000000,
}

var AntiDiagonalMasks = [15]Bitboard{
	0x1,
	0x102,
	0x10204,
	0x1020408,
	0x102040810,
	0x10204081020,
	0x1020408102040,
	0x102040810204080,
	0x204081020408000,
	0x408102040800000,
	0x810204080000000,
	0x1020408000000000,
	0x2040800000000000,
	0x4080000000000000,
	0x8000000000000000,
}

var SetFileMasks = [8]Bitboard{}
var SetRankMasks = [8]Bitboard{}
var ClearFileMasks = [8]Bitboard{}
var ClearRankMasks = [8]Bitboard{}

// look at slider_moves.go to understand the sizes 4096 and 512 for slider pieces, i.e. rook and bishop
var ComputedRookMoves = [64][4096]Bitboard{}
var ComputedBishopMoves = [64][512]Bitboard{}

var ComputedKnightMoves = [64]Bitboard{}
var ComputedPawnAdvances = [2][64]Bitboard{}
var ComputedPawnCaptures = [2][64]Bitboard{}
var ComputedKingMoves = [64]Bitboard{}

func ComputePieceMoveTables() {
	initializeFileAndRankMasks()
	occupySliderMoves()

	for square := uint8(0); square < 64; square++ {
		occupyKnightMoves(square)
		occupyKingMoves(square)
		occupyPawnMoves(square)
	}
}

func initializeFileAndRankMasks() {
	for i := uint8(0); i < 57; i += 8 {
		InitialSetBitBoard := FullBitBoard
		InitialClearBitBoard := EmptyBitBoard

		for j := i; j < i+8; j++ {
			InitialClearBitBoard.SetBit(j)
			InitialSetBitBoard.ClearBit(j)
		}

		SetRankMasks[i/8] = InitialClearBitBoard
		ClearRankMasks[i/8] = InitialSetBitBoard
	}

	for i := uint8(0); i < 8; i++ {
		InitialSetBitBoard := FullBitBoard
		InitialClearBitBoard := EmptyBitBoard

		for j := i; j < 64; j += 8 {
			InitialClearBitBoard.SetBit(j)
			InitialSetBitBoard.ClearBit(j)
		}

		SetFileMasks[i] = InitialClearBitBoard
		ClearFileMasks[i] = InitialSetBitBoard
	}
}

func occupyKnightMoves(square uint8) {
	pieceBitboard := BitboardForSquare[square]

	northNorthEast := (pieceBitboard & ClearFileMasks[FileH]) >> NorthOffset >> NorthOffset >> EastOffset
	northNorthWest := (pieceBitboard & ClearFileMasks[FileA]) >> NorthOffset >> NorthOffset << WestOffset
	northEastEast := (pieceBitboard & ClearFileMasks[FileG] & ClearFileMasks[FileH]) >> NorthOffset >> EastOffset >> EastOffset
	northWestWest := (pieceBitboard & ClearFileMasks[FileB] & ClearFileMasks[FileA]) >> NorthOffset << WestOffset << WestOffset
	southSouthEast := (pieceBitboard & ClearFileMasks[FileH]) << SouthOffset << SouthOffset >> EastOffset
	southsouthWest := (pieceBitboard & ClearFileMasks[FileA]) << SouthOffset << SouthOffset << WestOffset
	southEastEast := (pieceBitboard & ClearFileMasks[FileG] & ClearFileMasks[FileH]) << SouthOffset >> EastOffset >> EastOffset
	southWestWest := (pieceBitboard & ClearFileMasks[FileB] & ClearFileMasks[FileA]) << SouthOffset << WestOffset << WestOffset

	movesUnion := northNorthEast | northNorthWest | northEastEast | northWestWest | southSouthEast | southsouthWest | southEastEast | southWestWest

	ComputedKnightMoves[square] = movesUnion
}

func occupyKingMoves(square uint8) {
	pieceBitboard := BitboardForSquare[square]

	north := pieceBitboard >> NorthOffset
	south := pieceBitboard << SouthOffset
	east := (pieceBitboard & ClearFileMasks[FileH]) >> EastOffset
	west := (pieceBitboard & ClearFileMasks[FileA]) << WestOffset
	northEast := (pieceBitboard & ClearFileMasks[FileH]) >> NorthOffset >> EastOffset
	northWest := (pieceBitboard & ClearFileMasks[FileA]) >> NorthOffset << WestOffset
	southEast := (pieceBitboard & ClearFileMasks[FileH]) << SouthOffset >> EastOffset
	southWest := (pieceBitboard & ClearFileMasks[FileA]) << SouthOffset << WestOffset

	movesUnion := north | south | east | west | northEast | northWest | southEast | southWest

	ComputedKingMoves[square] = movesUnion
}

func occupyPawnMoves(square uint8) {
	pieceBitboard := BitboardForSquare[square]

	whiteAdvance := pieceBitboard >> NorthOffset
	blackAdvance := pieceBitboard << SouthOffset

	whiteCaptureEast := (pieceBitboard & ClearFileMasks[FileH]) >> NorthOffset >> EastOffset
	whiteCaptureWest := (pieceBitboard & ClearFileMasks[FileA]) >> NorthOffset << WestOffset
	blackCaptureEast := (pieceBitboard & ClearFileMasks[FileH]) << SouthOffset >> EastOffset
	blackCaptureWest := (pieceBitboard & ClearFileMasks[FileA]) << SouthOffset << WestOffset

	ComputedPawnAdvances[White][square] = whiteAdvance
	ComputedPawnAdvances[Black][square] = blackAdvance

	ComputedPawnCaptures[White][square] = whiteCaptureEast | whiteCaptureWest
	ComputedPawnCaptures[Black][square] = blackCaptureEast | blackCaptureWest

}

func occupySliderMoves() {
	OccupyRookMagicNumbers()
	OccupyBishopMagicNumbers()
}
