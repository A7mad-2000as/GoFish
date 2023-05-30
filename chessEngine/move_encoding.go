package chessEngine

type Move uint32

const (
	CaptureMoveType   uint8 = 0
	CastleMoveType    uint8 = 1
	PromotionMoveType uint8 = 2
	QuietMoveType     uint8 = 3

	PromotionToKnight uint8 = 0
	PromotionToBishop uint8 = 1
	PromotionToRook   uint8 = 2
	PromotionToQueen  uint8 = 3

	EnPassant uint8 = 1

	MoveTypeMask   = 0xc0000000
	MoveInfoMask   = 0x30000000
	FromSquareMask = 0x0fc00000
	ToSquareMask   = 0x003f0000
	ScoreMask      = 0x0000ffff
	NoScoreMask    = 0xffff0000

	MoveTypeOffset   = 30
	MoveInfoOffset   = 28
	FromSquareOffset = 22
	ToSquareOffset   = 16

	NullMove Move = 0
)

func CreateMove(fromSquare, toSquare, moveType, moveInfo uint8) Move {
	encodedMove := uint32(moveType)<<MoveTypeOffset |
		uint32(moveInfo)<<MoveInfoOffset |
		uint32(fromSquare)<<FromSquareOffset |
		uint32(toSquare)<<ToSquareOffset

	return Move(encodedMove)
}

func (move Move) GetMoveType() uint8 {
	return uint8((move & MoveTypeMask) >> MoveTypeOffset)
}

func (move Move) GetMoveInfo() uint8 {
	return uint8((move & MoveInfoMask) >> MoveInfoOffset)
}

func (move Move) GetFromSquare() uint8 {
	return uint8((move & FromSquareMask) >> FromSquareOffset)
}

func (move Move) GetToSquare() uint8 {
	return uint8((move & ToSquareMask) >> ToSquareOffset)
}

func (move Move) GetScore() uint16 {
	return uint16(move & ScoreMask)
}

func (move *Move) ModifyMoveScore(newScore uint16) {
	encodedMove := uint32(*move)
	encodedMove = ((encodedMove & NoScoreMask) | uint32(newScore))
	*move = Move(encodedMove)
}

func (move Move) IsSameMove(otherMove Move) bool {
	return (move & NoScoreMask) == (otherMove & NoScoreMask)
}

func (move Move) String() string {
	fromSquare := move.GetFromSquare()
	toSquare := move.GetToSquare()
	moveType := move.GetMoveType()
	moveInfo := move.GetMoveInfo()

	promotionMoveSuffix := ""

	if moveType == PromotionMoveType {
		switch moveInfo {
		case PromotionToQueen:
			promotionMoveSuffix = "q"
		case PromotionToRook:
			promotionMoveSuffix = "r"
		case PromotionToBishop:
			promotionMoveSuffix = "b"
		case PromotionToKnight:
			promotionMoveSuffix = "n"
		}
	}

	return convertSquareNumberToSquareNotation(fromSquare) + convertSquareNumberToSquareNotation(toSquare) + promotionMoveSuffix

}
