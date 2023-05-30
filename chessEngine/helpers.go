package chessEngine

import (
	"golang.org/x/exp/constraints"
)

var RandomNumberGeneratorSeeds [8]uint64 = [8]uint64{728, 10316, 55013, 32803, 12281, 15100, 16645, 255}

func File(square uint8) uint8 {
	return square % 8
}

func Rank(square uint8) uint8 {
	return square / 8
}

func convertSquareNotationToSquareNumber(coordinate string) uint8 {
	fileNumber := int(coordinate[0] - 'a')
	rankNumber := int(coordinate[1] - '1')
	return uint8(rankNumber*8 + fileNumber)
}

func convertSquareNumberToSquareNotation(square uint8) string {
	return string(rune('a'+File(square))) + string(rune('1'+Rank(square)))
}

func abs[Int constraints.Integer](n Int) Int {
	if n < 0 {
		return -n
	}
	return n
}

func max[Int constraints.Integer](a, b Int) Int {
	if a > b {
		return a
	}
	return b
}

type RandomNumberGenerator struct {
	seed uint64
}

func (prng *RandomNumberGenerator) Seed(seed uint64) {
	prng.seed = seed
}

func (prng *RandomNumberGenerator) Random64() uint64 {
	prng.seed ^= prng.seed >> 12
	prng.seed ^= prng.seed << 25
	prng.seed ^= prng.seed >> 27
	return prng.seed * 2685821657736338717
}

func (prng *RandomNumberGenerator) SparseRandom64() uint64 {
	return prng.Random64() & prng.Random64() & prng.Random64()
}
