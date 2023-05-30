package chessEngine

import (
	"math/bits"
)

type MagicPackage struct {
	MagicNumber        uint64
	Shift              uint8
	MaximalBlockerMask Bitboard
}

var RookMagicPackages [64]MagicPackage
var BishopMagicPackages [64]MagicPackage

func OccupyRookMagicNumbers() {
	randomNumberGenerator := RandomNumberGenerator{}

	for square := uint8(0); square < 64; square++ {
		magicPackage := &RookMagicPackages[square]

		magicPackage.MaximalBlockerMask = GetRookBlockerMask(square)
		numOfBlockerMaskBits := magicPackage.MaximalBlockerMask.CountSetBits()
		magicPackage.Shift = uint8(64 - numOfBlockerMaskBits)

		blockerMasks := make([]Bitboard, 1<<numOfBlockerMaskBits)
		blockerMask := EmptyBitBoard
		blockerMaskNumber := 0

		for ok := true; ok; ok = (blockerMask != 0) {
			blockerMasks[blockerMaskNumber] = blockerMask
			blockerMaskNumber++
			blockerMask = (blockerMask - magicPackage.MaximalBlockerMask) & magicPackage.MaximalBlockerMask
		}

		randomNumberGenerator.Seed(RandomNumberGeneratorSeeds[Rank(square)])
		correctMagicNumberNotFound := true

		for correctMagicNumberNotFound {
			magicNumberAttempt := randomNumberGenerator.SparseRandom64()

			magicPackage.MagicNumber = magicNumberAttempt
			correctMagicNumberNotFound = false

			ComputedRookMoves[square] = [4096]Bitboard{}

			for blockerMaskNumber = 0; blockerMaskNumber < (1 << numOfBlockerMaskBits); blockerMaskNumber++ {
				hashIndex := (uint64(blockerMasks[blockerMaskNumber]) * magicNumberAttempt) >> magicPackage.Shift
				moveBitboard := getRookMoveBitboard(square, blockerMasks[blockerMaskNumber])

				if ComputedRookMoves[square][hashIndex] != EmptyBitBoard && ComputedRookMoves[square][hashIndex] != moveBitboard {
					correctMagicNumberNotFound = true
					break
				}

				ComputedRookMoves[square][hashIndex] = moveBitboard
			}
		}
	}
}

func OccupyBishopMagicNumbers() {
	randomNumberGenerator := RandomNumberGenerator{}

	for square := uint8(0); square < 64; square++ {
		magicPackage := &BishopMagicPackages[square]

		magicPackage.MaximalBlockerMask = getBishopBlockerMask(square)
		numOfBlockerMaskBits := magicPackage.MaximalBlockerMask.CountSetBits()
		magicPackage.Shift = uint8(64 - numOfBlockerMaskBits)

		blockerMasks := make([]Bitboard, 1<<numOfBlockerMaskBits)
		blockerMask := EmptyBitBoard
		blockerMaskNumber := 0

		for ok := true; ok; ok = (blockerMask != 0) {
			blockerMasks[blockerMaskNumber] = blockerMask
			blockerMaskNumber++
			blockerMask = (blockerMask - magicPackage.MaximalBlockerMask) & magicPackage.MaximalBlockerMask
		}

		randomNumberGenerator.Seed(RandomNumberGeneratorSeeds[Rank(square)])
		correctMagicNumberNotFound := true

		for correctMagicNumberNotFound {
			magicNumberAttempt := randomNumberGenerator.SparseRandom64()
			magicPackage.MagicNumber = magicNumberAttempt
			correctMagicNumberNotFound = false

			ComputedBishopMoves[square] = [512]Bitboard{}

			for blockerMaskNumber = 0; blockerMaskNumber < (1 << numOfBlockerMaskBits); blockerMaskNumber++ {
				hashIndex := (uint64(blockerMasks[blockerMaskNumber]) * magicNumberAttempt) >> magicPackage.Shift
				moveBitboard := getBishopMoveBitboard(square, blockerMasks[blockerMaskNumber])

				if ComputedBishopMoves[square][hashIndex] != EmptyBitBoard && ComputedBishopMoves[square][hashIndex] != moveBitboard {
					correctMagicNumberNotFound = true
					break
				}

				ComputedBishopMoves[square][hashIndex] = moveBitboard
			}
		}
	}
}

func GetRookBlockerMask(square uint8) Bitboard {
	mask := EmptyBitBoard
	mask |= SetFileMasks[File(square)]
	mask &= ClearRankMasks[Rank1]
	mask &= ClearRankMasks[Rank8]
	mask |= SetRankMasks[Rank(square)]
	if square%8 == 0 {
		mask.ClearBit(uint8(0))
		mask.ClearBit(uint8(square + 7))
		mask.ClearBit(uint8(56))
	} else if (square+1)%8 == 0 {
		mask.ClearBit(uint8(7))
		mask.ClearBit(uint8(square - 7))
		mask.ClearBit(uint8(63))
	} else {
		mask &= ClearFileMasks[FileA]
		mask &= ClearFileMasks[FileH]
	}
	mask.ClearBit(square)

	return mask
}

func getBishopBlockerMask(square uint8) Bitboard {
	mask := EmptyBitBoard
	mask |= MainDiagonalMasks[File(square)-Rank(square)+7]
	mask |= AntiDiagonalMasks[14-(Rank(square)+File(square))]
	mask &= ClearRankMasks[Rank1]
	mask &= ClearRankMasks[Rank8]
	mask &= ClearFileMasks[FileA]
	mask &= ClearFileMasks[FileH]
	mask.ClearBit(square)

	return mask
}

func getRookMoveBitboard(square uint8, occupancy Bitboard) Bitboard {
	piece := uint64(BitboardForSquare[square])
	occupied := uint64(occupancy)

	rankMask := uint64(SetRankMasks[Rank(square)])
	fileMask := uint64(SetFileMasks[File(square)])

	westMoves := (occupied & rankMask) - 2*piece
	eastMoves := bits.Reverse64(bits.Reverse64(occupied&rankMask) - 2*bits.Reverse64(piece))
	westEastMoves := (westMoves ^ eastMoves) & rankMask

	northMoves := (occupied & fileMask) - 2*piece
	southMoves := bits.Reverse64(bits.Reverse64(occupied&fileMask) - 2*bits.Reverse64(piece))
	northSouthMoves := (northMoves ^ southMoves) & fileMask

	return Bitboard(westEastMoves | northSouthMoves)
}

func getBishopMoveBitboard(square uint8, occupancy Bitboard) Bitboard {
	piece := uint64(BitboardForSquare[square])
	occupied := uint64(occupancy)

	mainDiagonalMask := uint64(MainDiagonalMasks[File(square)-Rank(square)+7])
	antiDiagonalMask := uint64(AntiDiagonalMasks[14-(Rank(square)+File(square))])

	southWestMoves := (occupied & mainDiagonalMask) - 2*piece
	NorthEastMoves := bits.Reverse64(bits.Reverse64(occupied&mainDiagonalMask) - 2*bits.Reverse64(piece))
	mainDiagonalMoves := (southWestMoves ^ NorthEastMoves) & mainDiagonalMask

	southEastMoves := (occupied & antiDiagonalMask) - 2*piece
	northWestMoves := bits.Reverse64(bits.Reverse64(occupied&antiDiagonalMask) - 2*bits.Reverse64(piece))
	antiDiagonalMoves := (southEastMoves ^ northWestMoves) & antiDiagonalMask

	return Bitboard(mainDiagonalMoves | antiDiagonalMoves)

}
