package chessEngine

const (
	seedValue                  = 1
	NoEnPassantOnAnyFile uint8 = 8
)

type zobrist struct {
	pieceSquareRandomNumbers    [768]uint64
	enPassantFileRandomNumbers  [9]uint64
	castlingRightsRandomNumbers [16]uint64
	sideToMoveRandomNumber      uint64
}

var ZobristSingleton zobrist

func (zobrist *zobrist) populateRandomNumbers() {
	var prng RandomNumberGenerator
	prng.Seed(seedValue)
	for pieceSquareIndex := 0; pieceSquareIndex < 768; pieceSquareIndex++ {
		zobrist.pieceSquareRandomNumbers[pieceSquareIndex] = prng.Random64()
	}
	for fileIndex := 0; fileIndex < 8; fileIndex++ {
		zobrist.enPassantFileRandomNumbers[fileIndex] = prng.Random64()
	}
	zobrist.enPassantFileRandomNumbers[NoEnPassantOnAnyFile] = prng.Random64()
	for castlingRightsMaskIndex := 0; castlingRightsMaskIndex < 16; castlingRightsMaskIndex++ {
		zobrist.castlingRightsRandomNumbers[castlingRightsMaskIndex] = prng.Random64()
	}
	zobrist.sideToMoveRandomNumber = prng.Random64()

}

var enPassantFileIndices = [65]uint8{
	8, 8, 8, 8, 8, 8, 8, 8,
	8, 8, 8, 8, 8, 8, 8, 8,
	0, 1, 2, 3, 4, 5, 6, 7,
	8, 8, 8, 8, 8, 8, 8, 8,
	8, 8, 8, 8, 8, 8, 8, 8,
	0, 1, 2, 3, 4, 5, 6, 7,
	8, 8, 8, 8, 8, 8, 8, 8,
	8, 8, 8, 8, 8, 8, 8, 8,
	8,
}

func (zobrist *zobrist) getEnPassantFile(EnPassantSquare uint8) uint8 {
	return enPassantFileIndices[EnPassantSquare]

}

// convert 3d to 1d (z * height * width) + (y * width) + x)
func (zobrist *zobrist) GetPieceSquareRandomNumber(piece Piece, square uint8) uint64 {
	return zobrist.pieceSquareRandomNumbers[(uint16(piece.PieceType)*2+uint16(piece.Color))*64+uint16(square)]
}
func (zobrist *zobrist) GetEnPassantFileRandomNumber(EnPassantSquare uint8) uint64 {
	return zobrist.enPassantFileRandomNumbers[zobrist.getEnPassantFile(EnPassantSquare)]
}
func (zobrist *zobrist) GetCastlingRightsRandomNumber(castlingRights uint8) uint64 {
	return zobrist.castlingRightsRandomNumbers[castlingRights]
}
func (zobrist *zobrist) GetSideToMoveRandomNumber() uint64 {
	return zobrist.sideToMoveRandomNumber
}

func (zobrist *zobrist) GenHash(position *Position) (hash uint64) {
	for sq := uint8(0); sq < 64; sq++ {
		piece := position.SquareContent[sq]
		if piece.PieceType != NoneType {
			hash ^= zobrist.GetPieceSquareRandomNumber(piece, sq)
		}
	}

	hash ^= zobrist.GetEnPassantFileRandomNumber(position.EnPassantSquare)
	hash ^= zobrist.GetCastlingRightsRandomNumber(position.CastlingRights)

	if position.SideToMove == White {
		hash ^= zobrist.GetSideToMoveRandomNumber()
	}

	return hash
}

func InitializeZobristHashing() {
	ZobristSingleton = zobrist{}
	ZobristSingleton.populateRandomNumbers()
}
