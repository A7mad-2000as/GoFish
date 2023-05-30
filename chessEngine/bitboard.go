package chessEngine

import (
	"fmt"
	"math/bits"
)

type Bitboard uint64
type Direction uint8

const (
	// use ^EmptyBoard instead of FullBB
	EmptyBitBoard Bitboard = 0x0
	FullBitBoard  Bitboard = 0xffffffffffffffff
	FileHBitBoard Bitboard = 0x8080808080808080
	FileABitBoard Bitboard = 0x0101010101010101
)
const (
	NORTH Direction = iota
	SOUTH
	EAST
	WEST
	North_East
	North_West
	South_East
	South_West
)

var BitboardForSquare = [65]Bitboard{
	0x8000000000000000, 0x4000000000000000, 0x2000000000000000, 0x1000000000000000,
	0x0800000000000000, 0x0400000000000000, 0x0200000000000000, 0x0100000000000000,
	0x0080000000000000, 0x0040000000000000, 0x0020000000000000, 0x0010000000000000,
	0x0008000000000000, 0x0004000000000000, 0x0002000000000000, 0x0001000000000000,
	0x0000800000000000, 0x0000400000000000, 0x0000200000000000, 0x0000100000000000,
	0x0000080000000000, 0x0000040000000000, 0x0000020000000000, 0x0000010000000000,
	0x0000008000000000, 0x0000004000000000, 0x0000002000000000, 0x0000001000000000,
	0x0000000800000000, 0x0000000400000000, 0x0000000200000000, 0x0000000100000000,
	0x0000000080000000, 0x0000000040000000, 0x0000000020000000, 0x0000000010000000,
	0x0000000008000000, 0x0000000004000000, 0x0000000002000000, 0x0000000001000000,
	0x0000000000800000, 0x0000000000400000, 0x0000000000200000, 0x0000000000100000,
	0x0000000000080000, 0x0000000000040000, 0x0000000000020000, 0x0000000000010000,
	0x0000000000008000, 0x0000000000004000, 0x0000000000002000, 0x0000000000001000,
	0x0000000000000800, 0x0000000000000400, 0x0000000000000200, 0x0000000000000100,
	0x0000000000000080, 0x0000000000000040, 0x0000000000000020, 0x0000000000000010,
	0x0000000000000008, 0x0000000000000004, 0x0000000000000002, 0x0000000000000001,
	0x0000000000000000,
}

func (bitboard Bitboard) CountSetBits() int {
	return bits.OnesCount64(uint64(bitboard))
}

func (bitboard *Bitboard) SetBit(square uint8) {
	*bitboard |= BitboardForSquare[square]
}

func (bitboard *Bitboard) ClearBit(square uint8) {
	*bitboard &= ^BitboardForSquare[square]
}

func (bitboard Bitboard) MostSignificantBit() uint8 {
	return uint8(bits.LeadingZeros64(uint64(bitboard)))
}

func (bitboard *Bitboard) PopMostSignificantBit() uint8 {
	square := bitboard.MostSignificantBit()
	bitboard.ClearBit(square)
	return square
}

func (bitboard Bitboard) GetShiftedBitBoard(b uint64, direction Direction) Bitboard {
	switch direction {
	case NORTH:
		return bitboard >> 8
	case SOUTH:
		return bitboard << 8
	case EAST:
		return (bitboard & ^FileHBitBoard) >> 1
	case WEST:
		return (bitboard & ^FileABitBoard) << 1
	case North_East:
		return (bitboard & ^FileHBitBoard) >> 9
	case North_West:
		return (bitboard & ^FileABitBoard) >> 7
	case South_East:
		return (bitboard & ^FileHBitBoard) << 7
	case South_West:
		return (bitboard & ^FileABitBoard) << 9
	default:
		return 0
	}

}
func (bitboard Bitboard) String() (formattedBitBoard string) {
	binaryRepresentation := fmt.Sprintf("%064b\n", bitboard)
	formattedBitBoard += "\n"
	for rowStartPosition := 56; rowStartPosition >= 0; rowStartPosition -= 8 {
		formattedBitBoard += fmt.Sprintf("%v | ", (rowStartPosition/8)+1)
		for position := rowStartPosition; position < rowStartPosition+8; position++ {
			charAtSquare := binaryRepresentation[position]
			if charAtSquare == '0' {
				charAtSquare = '.'
			}
			formattedBitBoard += fmt.Sprintf("%c ", charAtSquare)
		}
		formattedBitBoard += "\n"
	}

	formattedBitBoard += "   ----------------"
	formattedBitBoard += "\n    a b c d e f g h"
	formattedBitBoard += "\n"
	return formattedBitBoard
}
