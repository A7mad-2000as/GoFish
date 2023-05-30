package chessEngine

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	Black     uint8 = 0
	White     uint8 = 1
	NoneColor uint8 = 2

	Pawn     uint8 = 0
	Knight   uint8 = 1
	Bishop   uint8 = 2
	Rook     uint8 = 3
	Queen    uint8 = 4
	King     uint8 = 5
	NoneType uint8 = 6

	White_Kingside_Castle_Right  uint8 = 0b00001000
	White_Queenside_Castle_Right uint8 = 0b00000100
	Black_Kingside_Castle_Right  uint8 = 0b00000010
	Black_Queenside_Castle_Right uint8 = 0b00000001

	A1, B1, C1, D1, E1, F1, G1, H1 = 0, 1, 2, 3, 4, 5, 6, 7
	A2, B2, C2, D2, E2, F2, G2, H2 = 8, 9, 10, 11, 12, 13, 14, 15
	A3, B3, C3, D3, E3, F3, G3, H3 = 16, 17, 18, 19, 20, 21, 22, 23
	A4, B4, C4, D4, E4, F4, G4, H4 = 24, 25, 26, 27, 28, 29, 30, 31
	A5, B5, C5, D5, E5, F5, G5, H5 = 32, 33, 34, 35, 36, 37, 38, 39
	A6, B6, C6, D6, E6, F6, G6, H6 = 40, 41, 42, 43, 44, 45, 46, 47
	A7, B7, C7, D7, E7, F7, G7, H7 = 48, 49, 50, 51, 52, 53, 54, 55
	A8, B8, C8, D8, E8, F8, G8, H8 = 56, 57, 58, 59, 60, 61, 62, 63

	NoneSquare       = 64
	FENStartPosition = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 0"
)

type RookCastleMove struct {
	fromSquare uint8
	toSquare   uint8
}

var rookCastlingMovesForKingCastlingMove = map[uint8]RookCastleMove{
	G1: {H1, F1},
	C1: {A1, D1},
	G8: {H8, F8},
	C8: {A8, D8},
}

var CharToPiece = map[byte]Piece{
	'P': {Pawn, White},
	'N': {Knight, White},
	'B': {Bishop, White},
	'R': {Rook, White},
	'Q': {Queen, White},
	'K': {King, White},
	'p': {Pawn, Black},
	'n': {Knight, Black},
	'b': {Bishop, Black},
	'r': {Rook, Black},
	'q': {Queen, Black},
	'k': {King, Black},
}

var BoardSquaresNormalAndFlipped [2][64]int = [2][64]int{
	{
		0, 1, 2, 3, 4, 5, 6, 7,
		8, 9, 10, 11, 12, 13, 14, 15,
		16, 17, 18, 19, 20, 21, 22, 23,
		24, 25, 26, 27, 28, 29, 30, 31,
		32, 33, 34, 35, 36, 37, 38, 39,
		40, 41, 42, 43, 44, 45, 46, 47,
		48, 49, 50, 51, 52, 53, 54, 55,
		56, 57, 58, 59, 60, 61, 62, 63,
	},

	{
		56, 57, 58, 59, 60, 61, 62, 63,
		48, 49, 50, 51, 52, 53, 54, 55,
		40, 41, 42, 43, 44, 45, 46, 47,
		32, 33, 34, 35, 36, 37, 38, 39,
		24, 25, 26, 27, 28, 29, 30, 31,
		16, 17, 18, 19, 20, 21, 22, 23,
		8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7,
	},
}

var BoardRanksNormalAndFlipped = [2][8]uint8{
	{Rank8, Rank7, Rank6, Rank5, Rank4, Rank3, Rank2, Rank1},
	{Rank1, Rank2, Rank3, Rank4, Rank5, Rank6, Rank7, Rank8},
}

var PieceTypeToChar = map[uint8]rune{
	Pawn:     'p',
	Knight:   'n',
	Bishop:   'b',
	Rook:     'r',
	Queen:    'q',
	King:     'k',
	NoneType: '.',
}

var castlingRightsUpdateMaskBySquare = [64]uint8{
	0xb, 0xf, 0xf, 0xf, 0x3, 0xf, 0xf, 0x7,
	0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf,
	0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf,
	0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf,
	0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf,
	0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf,
	0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf,
	0xe, 0xf, 0xf, 0xf, 0xc, 0xf, 0xf, 0xd,
}

type Piece struct {
	PieceType uint8
	Color     uint8
}

type Position struct {
	// Board representations
	SquareContent  [64]Piece
	PiecesBitBoard [2][6]Bitboard
	ColorsBitBoard [2]Bitboard

	// Game state information
	SideToMove      uint8
	PositionHash    uint64
	CastlingRights  uint8
	Rule50          uint8
	EnPassantSquare uint8
	PreviousStates  [100]StateInfo
	CurrentPly      uint16
	MidGameScores   [2]int16
	EndGameScores   [2]int16
	stateStackSize  uint8
	Phase           int16
}
type StateInfo struct {
	PositionHash    uint64
	CastlingRights  uint8
	Rule50          uint8
	EnPassantSquare uint8
	CapturedPiece   Piece
	MovedPiece      Piece
}

func (position *Position) LoadFEN(FEN string, evaluator Evaluator) {
	// Reset the internal fields of the position
	position.PiecesBitBoard = [2][6]Bitboard{}
	position.ColorsBitBoard = [2]Bitboard{}
	position.SquareContent = [64]Piece{}
	position.MidGameScores = [2]int16{}
	position.EndGameScores = [2]int16{}
	position.CastlingRights = 0
	position.Phase = evaluator.GetTotalPhaseWeight()

	for square := range position.SquareContent {
		position.SquareContent[square] = Piece{PieceType: NoneType, Color: NoneColor}
	}

	// Load in each field of the FEN string.
	fields := strings.Fields(FEN)
	pieces := fields[0]
	color := fields[1]
	castling := fields[2]
	ep := fields[3]
	halfMove := fields[4]
	fullMove := fields[5]

	// Loop over each square of the board, rank by rank, from left to right,
	// loading in pieces at squares described by the FEN string.
	for index, sq := 0, uint8(56); index < len(pieces); index++ {
		char := pieces[index]
		switch char {
		case 'p', 'n', 'b', 'r', 'q', 'k', 'P', 'N', 'B', 'R', 'Q', 'K':
			piece := CharToPiece[char]
			position.placePieceAndAdjustScore(piece, sq, evaluator)
			sq++
		case '/':
			sq -= 16
		case '1', '2', '3', '4', '5', '6', '7', '8':
			sq += pieces[index] - '0'
		}
	}

	// Set the side to move for the position.
	position.SideToMove = Black
	if color == "w" {
		position.SideToMove = White
	}

	// Set the en passant square for the position.
	position.EnPassantSquare = NoneSquare
	if ep != "-" {
		position.EnPassantSquare = convertSquareNotationToSquareNumber(ep)
		if (ComputedPawnCaptures[position.SideToMove^1][position.EnPassantSquare] & (position.PiecesBitBoard[position.SideToMove][Pawn])) == 0 {
			position.EnPassantSquare = NoneSquare
		}
	}

	// Set the half move counter and game ply for the position.
	halfMoveCounter, _ := strconv.Atoi(halfMove)
	position.Rule50 = uint8(halfMoveCounter)

	gamePly, _ := strconv.Atoi(fullMove)
	gamePly *= 2
	if position.SideToMove == Black {
		gamePly--
	}
	position.CurrentPly = uint16(gamePly)

	// Set the castling rights, for the position.
	for _, char := range castling {
		switch char {
		case 'K':
			position.CastlingRights |= White_Kingside_Castle_Right
		case 'Q':
			position.CastlingRights |= White_Queenside_Castle_Right
		case 'k':
			position.CastlingRights |= Black_Kingside_Castle_Right
		case 'q':
			position.CastlingRights |= Black_Queenside_Castle_Right
		}
	}

	// Generate the zobrist hash for the position...
	position.PositionHash = ZobristSingleton.GenHash(position)
}

func (position *Position) DoMove(move Move, evaluator Evaluator) (isValid bool) {
	fromSquare, toSquare, moveType, additionalMoveInfo := extractMoveInfo(move)

	state := StateInfo{
		PositionHash:    position.PositionHash,
		CastlingRights:  position.CastlingRights,
		EnPassantSquare: position.EnPassantSquare,
		Rule50:          position.Rule50,
		CapturedPiece:   position.SquareContent[toSquare],
		MovedPiece:      position.SquareContent[fromSquare],
	}

	position.CurrentPly++
	position.Rule50++

	position.PositionHash ^= ZobristSingleton.GetEnPassantFileRandomNumber(position.EnPassantSquare)
	position.EnPassantSquare = NoneSquare

	position.clearSquareUpdateHashAndAdjustScore(fromSquare, evaluator)

	switch moveType {
	case QuietMoveType:
		position.placePieceUpdateHashAndAdjustScore(Piece{PieceType: state.MovedPiece.PieceType, Color: position.SideToMove}, toSquare, evaluator)
	case CaptureMoveType:
		if additionalMoveInfo == EnPassant {
			capSq := uint8(int8(toSquare) - getPawnForwardDelta(position.SideToMove))
			state.CapturedPiece = position.SquareContent[capSq]

			position.clearSquareUpdateHashAndAdjustScore(capSq, evaluator)
			position.placePieceAndAdjustScore(Piece{PieceType: Pawn, Color: position.SideToMove}, toSquare, evaluator)
		} else {
			position.clearSquareUpdateHashAndAdjustScore(toSquare, evaluator)
			position.placePieceUpdateHashAndAdjustScore(Piece{PieceType: state.MovedPiece.PieceType, Color: position.SideToMove}, toSquare, evaluator)
		}

		position.Rule50 = 0
	case CastleMoveType:
		position.placePieceUpdateHashAndAdjustScore(Piece{PieceType: state.MovedPiece.PieceType, Color: position.SideToMove}, toSquare, evaluator)

		rookFrom, rookTo := rookCastlingMovesForKingCastlingMove[toSquare].fromSquare, rookCastlingMovesForKingCastlingMove[toSquare].toSquare
		position.clearSquareUpdateHashAndAdjustScore(rookFrom, evaluator)
		position.placePieceUpdateHashAndAdjustScore(Piece{PieceType: Rook, Color: position.SideToMove}, rookTo, evaluator)
	case PromotionMoveType:
		if state.CapturedPiece.PieceType != NoneType {
			position.clearSquareUpdateHashAndAdjustScore(toSquare, evaluator)
		}
		position.placePieceUpdateHashAndAdjustScore(Piece{Color: position.SideToMove, PieceType: additionalMoveInfo + 1}, toSquare, evaluator)
	}

	if state.MovedPiece.PieceType == Pawn {
		position.Rule50 = 0

		if abs(int8(fromSquare)-int8(toSquare)) == 16 {
			position.EnPassantSquare = uint8(int8(toSquare) - getPawnForwardDelta(position.SideToMove))
			if ComputedPawnCaptures[position.SideToMove][position.EnPassantSquare]&(position.PiecesBitBoard[position.SideToMove^1][Pawn]) == 0 {
				position.EnPassantSquare = NoneSquare
			}
		}

	}

	position.PositionHash ^= ZobristSingleton.GetCastlingRightsRandomNumber(position.CastlingRights)
	position.CastlingRights = position.CastlingRights & castlingRightsUpdateMaskBySquare[fromSquare] & castlingRightsUpdateMaskBySquare[toSquare]
	position.PositionHash ^= ZobristSingleton.GetCastlingRightsRandomNumber(position.CastlingRights)
	position.PositionHash ^= ZobristSingleton.GetEnPassantFileRandomNumber(position.EnPassantSquare)
	position.PreviousStates[position.stateStackSize] = state
	position.stateStackSize++

	occupancyBitboard := position.ColorsBitBoard[White] | position.ColorsBitBoard[Black]
	validPosition := !squaresAreCoveredByOpponent([]uint8{position.PiecesBitBoard[position.SideToMove][King].MostSignificantBit()}, position, occupancyBitboard)

	position.SideToMove ^= 1
	position.PositionHash ^= ZobristSingleton.GetSideToMoveRandomNumber()
	return validPosition

}
func (position *Position) UnDoPreviousMove(previousMove Move, evaluator Evaluator) {
	position.stateStackSize--
	stateInfoForMoveReversal := position.PreviousStates[position.stateStackSize]

	position.PositionHash = stateInfoForMoveReversal.PositionHash
	position.CastlingRights = stateInfoForMoveReversal.CastlingRights
	position.Rule50 = stateInfoForMoveReversal.Rule50
	position.EnPassantSquare = stateInfoForMoveReversal.EnPassantSquare

	position.SideToMove ^= 1
	position.CurrentPly--
	fromSquare, toSquare, moveType, additionalMoveInfo := extractMoveInfo(previousMove)
	position.placePieceAndAdjustScore(stateInfoForMoveReversal.MovedPiece, fromSquare, evaluator)
	switch moveType {
	case QuietMoveType:
		position.clearPieceAtSquareAdjustScore(toSquare, evaluator)
	case CaptureMoveType:
		if additionalMoveInfo == EnPassant {
			capturedSquare := uint8(int8(toSquare) - getPawnForwardDelta(position.SideToMove))
			position.clearPieceAtSquareAdjustScore(toSquare, evaluator)
			position.placePieceAndAdjustScore(Piece{PieceType: Pawn, Color: stateInfoForMoveReversal.CapturedPiece.Color}, capturedSquare, evaluator)
		} else {
			position.clearPieceAtSquareAdjustScore(toSquare, evaluator)
			position.placePieceAndAdjustScore(stateInfoForMoveReversal.CapturedPiece, toSquare, evaluator)
		}
	case PromotionMoveType:
		position.clearPieceAtSquareAdjustScore(toSquare, evaluator)
		if stateInfoForMoveReversal.CapturedPiece.PieceType != NoneType {
			position.placePieceAndAdjustScore(stateInfoForMoveReversal.CapturedPiece, toSquare, evaluator)
		}
	case CastleMoveType:
		position.clearPieceAtSquareAdjustScore(toSquare, evaluator)
		castleFromSquare, castleToSquare := rookCastlingMovesForKingCastlingMove[toSquare].fromSquare, rookCastlingMovesForKingCastlingMove[toSquare].toSquare
		position.clearPieceAtSquareAdjustScore(castleToSquare, evaluator)
		position.placePieceAndAdjustScore(Piece{PieceType: Rook, Color: position.SideToMove}, castleFromSquare, evaluator)
	}
}
func (position *Position) DoNullMove() {
	currentState := StateInfo{
		PositionHash:    position.PositionHash,
		CastlingRights:  position.CastlingRights,
		Rule50:          position.Rule50,
		EnPassantSquare: position.EnPassantSquare,
	}
	position.PreviousStates[position.stateStackSize] = currentState
	position.stateStackSize++
	position.clearCurrentEnPassantInfo()
	position.CurrentPly++
	position.Rule50 = 0
	position.SideToMove ^= 1
	position.PositionHash ^= ZobristSingleton.GetSideToMoveRandomNumber()
}
func (position *Position) unDoPreviousNullMove() {
	position.stateStackSize--
	stateInfoForMoveReversal := position.PreviousStates[position.stateStackSize]
	position.PositionHash = stateInfoForMoveReversal.PositionHash
	position.CastlingRights = stateInfoForMoveReversal.CastlingRights
	position.Rule50 = stateInfoForMoveReversal.Rule50
	position.EnPassantSquare = stateInfoForMoveReversal.EnPassantSquare
	position.CurrentPly--
	position.SideToMove ^= 1
}
func (position *Position) IsCurrentSideInCheck() bool {
	kingSquare := position.PiecesBitBoard[position.SideToMove][King].MostSignificantBit()
	occupancyBitboard := position.ColorsBitBoard[White] | position.ColorsBitBoard[Black]
	return squaresAreCoveredByOpponent([]uint8{uint8(kingSquare)}, position, occupancyBitboard)
}
func (position *Position) HasNoMajorOrMinorPieces() bool {
	whiteKnights := position.PiecesBitBoard[White][Knight].CountSetBits()
	blackKnights := position.PiecesBitBoard[Black][Knight].CountSetBits()

	whiteBishops := position.PiecesBitBoard[White][Bishop].CountSetBits()
	blackBishops := position.PiecesBitBoard[Black][Bishop].CountSetBits()

	whiteRooks := position.PiecesBitBoard[White][Rook].CountSetBits()
	blackRooks := position.PiecesBitBoard[Black][Rook].CountSetBits()

	whiteQueens := position.PiecesBitBoard[White][Queen].CountSetBits()
	blackQueens := position.PiecesBitBoard[Black][Queen].CountSetBits()

	totalMajorMinorPieces := whiteKnights + blackKnights + whiteBishops + blackBishops + whiteRooks + blackRooks + whiteQueens + blackQueens

	return totalMajorMinorPieces == 0
}

func (position *Position) clearCurrentEnPassantInfo() {
	position.PositionHash ^= ZobristSingleton.GetEnPassantFileRandomNumber(position.EnPassantSquare)
	position.EnPassantSquare = NoneSquare
}

func (position *Position) clearPieceAtSquareAdjustScore(square uint8, evaluator Evaluator) {
	PSQT_MG := evaluator.GetMiddleGamePieceSquareTable()
	PSQT_EG := evaluator.GetEndGamePieceSquareTable()
	PieceValueMG := evaluator.GetMiddleGamePieceValues()
	PieceValueEG := evaluator.GetEndGamePieceValues()
	PhaseValues := evaluator.GetPhaseValues()

	piece := &position.SquareContent[square]
	color := piece.Color
	pieceType := piece.PieceType
	//update board
	position.PiecesBitBoard[piece.Color][piece.PieceType].ClearBit(square)
	position.ColorsBitBoard[piece.Color].ClearBit(square)
	//update material scores
	position.MidGameScores[color] -= PieceValueMG[pieceType] + PSQT_MG[pieceType][BoardSquaresNormalAndFlipped[color][square]]
	position.EndGameScores[color] -= PieceValueEG[pieceType] + PSQT_EG[pieceType][BoardSquaresNormalAndFlipped[color][square]]
	position.Phase += PhaseValues[piece.PieceType]

	piece.PieceType = NoneType
	piece.Color = NoneColor
}

func (position *Position) clearSquareUpdateHashAndAdjustScore(square uint8, evaluator Evaluator) {
	piece := position.SquareContent[square]
	position.PositionHash ^= ZobristSingleton.GetPieceSquareRandomNumber(piece, square)
	position.clearPieceAtSquareAdjustScore(square, evaluator)
}
func (position *Position) placePieceAndAdjustScore(piece Piece, toSquare uint8, evaluator Evaluator) {
	PSQT_MG := evaluator.GetMiddleGamePieceSquareTable()
	PSQT_EG := evaluator.GetEndGamePieceSquareTable()
	PieceValueMG := evaluator.GetMiddleGamePieceValues()
	PieceValueEG := evaluator.GetEndGamePieceValues()
	PhaseValues := evaluator.GetPhaseValues()

	//update board
	position.PiecesBitBoard[piece.Color][piece.PieceType].SetBit(toSquare)
	position.ColorsBitBoard[piece.Color].SetBit(toSquare)
	position.SquareContent[toSquare].PieceType = piece.PieceType
	position.SquareContent[toSquare].Color = piece.Color

	//update material scores
	position.MidGameScores[piece.Color] += PieceValueMG[piece.PieceType] + PSQT_MG[piece.PieceType][BoardSquaresNormalAndFlipped[piece.Color][toSquare]]
	position.EndGameScores[piece.Color] += PieceValueEG[piece.PieceType] + PSQT_EG[piece.PieceType][BoardSquaresNormalAndFlipped[piece.Color][toSquare]]
	position.Phase -= PhaseValues[piece.PieceType]
}
func (position *Position) placePieceUpdateHashAndAdjustScore(piece Piece, toSquare uint8, evaluator Evaluator) {
	position.PositionHash ^= ZobristSingleton.GetPieceSquareRandomNumber(piece, toSquare)
	position.placePieceAndAdjustScore(piece, toSquare, evaluator)
}
func getPawnForwardDelta(color uint8) int8 {
	if color == White {
		return 8
	}
	return -8
}
func extractMoveInfo(move Move) (fromSquare uint8, toSquare uint8, moveType uint8, additionalMoveInfo uint8) {
	fromSquare = move.GetFromSquare()
	toSquare = move.GetToSquare()
	moveType = move.GetMoveType()
	additionalMoveInfo = move.GetMoveInfo()
	return
}
func (pos *Position) MoveIsPseduoLegal(move Move) bool {
	fromSq, toSq := move.GetFromSquare(), move.GetToSquare()
	moved := pos.SquareContent[fromSq]
	captured := pos.SquareContent[toSq]

	toBB := BitboardForSquare[toSq]
	allBB := pos.ColorsBitBoard[White] | pos.ColorsBitBoard[Black]
	sideToMove := pos.SideToMove

	if moved.Color != sideToMove ||
		captured.PieceType == King ||
		captured.Color == sideToMove {
		return false
	}

	if moved.PieceType == Pawn {
		if fromSq > 55 || fromSq < 8 {
			return false
		}

		// Credit to the Stockfish team for the idea behind this section of code to
		// verify pseduo-legal pawn moves.
		if ((ComputedPawnCaptures[sideToMove][fromSq] & toBB & allBB) == 0) &&
			!((fromSq+uint8(getPawnForwardDelta(sideToMove)) == toSq) && (captured.PieceType == NoneType)) &&
			!((fromSq+uint8(getPawnForwardDelta(sideToMove)*2) == toSq) &&
				captured.PieceType == NoneType &&
				pos.SquareContent[toSq-uint8(getPawnForwardDelta(sideToMove))].PieceType == NoneType &&
				isDoublePushAllowed(fromSq, sideToMove)) {
			return false
		}
	} else {
		if (moved.PieceType == Knight && ((ComputedKnightMoves[fromSq] & toBB) == 0)) ||
			(moved.PieceType == Bishop && ((GetBishopPseudoLegalMoves(fromSq, allBB) & toBB) == 0)) ||
			(moved.PieceType == Rook && ((GetRookPseudoLegalMoves(fromSq, allBB) & toBB) == 0)) ||
			(moved.PieceType == Queen && (((GetBishopPseudoLegalMoves(fromSq, allBB) | GetRookPseudoLegalMoves(fromSq, allBB)) & toBB) == 0)) ||
			(moved.PieceType == King && ((ComputedKingMoves[fromSq] & toBB) == 0)) {
			return false
		}
	}

	return true
}
func isDoublePushAllowed(fromSquare uint8, sideToMove uint8) bool {
	if sideToMove == White {
		return Rank(fromSquare) == Rank2
	}
	return Rank(fromSquare) == Rank7
}

func (position Position) String() (boardStr string) {
	boardStr += "\n"
	for rankStartPos := 56; rankStartPos >= 0; rankStartPos -= 8 {
		boardStr += fmt.Sprintf("%v | ", (rankStartPos/8)+1)
		for index := rankStartPos; index < rankStartPos+8; index++ {
			piece := position.SquareContent[index]
			pieceChar := PieceTypeToChar[piece.PieceType]
			if piece.Color == White {
				pieceChar = unicode.ToUpper(pieceChar)
			}

			boardStr += fmt.Sprintf("%c ", pieceChar)
		}
		boardStr += "\n"
	}

	boardStr += "   ----------------"
	boardStr += "\n    a b c d e f g h"

	boardStr += "\n\n"
	if position.SideToMove == White {
		boardStr += "turn: white\n"
	} else {
		boardStr += "turn: black\n"
	}

	boardStr += "castling rights: "
	if position.CastlingRights&White_Kingside_Castle_Right != 0 {
		boardStr += "K"
	}
	if position.CastlingRights&White_Queenside_Castle_Right != 0 {
		boardStr += "Q"
	}
	if position.CastlingRights&Black_Kingside_Castle_Right != 0 {
		boardStr += "k"
	}
	if position.CastlingRights&Black_Queenside_Castle_Right != 0 {
		boardStr += "q"
	}

	boardStr += "\nen passant: "
	if position.EnPassantSquare == NoneSquare {
		boardStr += "none"
	} else {
		boardStr += convertSquareNumberToSquareNotation(position.EnPassantSquare)
	}

	boardStr += fmt.Sprintf("\nfen: %s", position.GenFEN())
	boardStr += fmt.Sprintf("\nzobrist hash: 0x%x", position.PositionHash)
	boardStr += fmt.Sprintf("\nrule 50: %d\n", position.Rule50)
	boardStr += fmt.Sprintf("game ply: %d\n", position.CurrentPly)
	return boardStr
}

func (pos Position) GenFEN() string {
	positionStr := strings.Builder{}

	for rankStartPos := 56; rankStartPos >= 0; rankStartPos -= 8 {
		emptySquares := 0
		for sq := rankStartPos; sq < rankStartPos+8; sq++ {
			piece := pos.SquareContent[sq]
			if piece.PieceType == NoneType {
				emptySquares++
			} else {
				// If we have some consecutive empty squares, add then to the FEN
				// string board, and reset the empty squares counter.
				if emptySquares > 0 {
					positionStr.WriteString(strconv.Itoa(emptySquares))
					emptySquares = 0
				}

				piece := pos.SquareContent[sq]
				pieceChar := PieceTypeToChar[piece.PieceType]
				if piece.Color == White {
					pieceChar = unicode.ToUpper(pieceChar)
				}

				positionStr.WriteRune(pieceChar)
			}
		}

		if emptySquares > 0 {
			positionStr.WriteString(strconv.Itoa(emptySquares))
			emptySquares = 0
		}

		positionStr.WriteString("/")

	}

	sideToMove := ""
	castlingRights := ""
	epSquare := ""

	if pos.SideToMove == White {
		sideToMove = "w"
	} else {
		sideToMove = "b"
	}

	if pos.CastlingRights&White_Kingside_Castle_Right != 0 {
		castlingRights += "K"
	}
	if pos.CastlingRights&White_Queenside_Castle_Right != 0 {
		castlingRights += "Q"
	}
	if pos.CastlingRights&Black_Kingside_Castle_Right != 0 {
		castlingRights += "k"
	}
	if pos.CastlingRights&Black_Queenside_Castle_Right != 0 {
		castlingRights += "q"
	}

	if castlingRights == "" {
		castlingRights = "-"
	}

	if pos.EnPassantSquare == NoneSquare {
		epSquare = "-"
	} else {
		epSquare = convertSquareNumberToSquareNotation(pos.EnPassantSquare)
	}

	fullMoveCount := pos.CurrentPly / 2
	if pos.CurrentPly%2 != 0 {
		fullMoveCount = pos.CurrentPly/2 + 1
	}

	return fmt.Sprintf(
		"%s %s %s %s %d %d",
		strings.TrimSuffix(positionStr.String(), "/"),
		sideToMove, castlingRights, epSquare,
		pos.Rule50, fullMoveCount,
	)
}
