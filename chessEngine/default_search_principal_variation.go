package chessEngine

import "fmt"

type PV struct {
	moves []Move
}

func (pv *PV) DeleteVariation() {
	pv.moves = nil
}

func (pv *PV) SetNewVariation(firstMove Move, continuation PV) {
	pv.DeleteVariation()
	pv.moves = append(pv.moves, firstMove)
	pv.moves = append(pv.moves, continuation.moves...)
}

func (pv *PV) GetVariationFirstMove() Move {
	return pv.moves[0]
}

func (pv PV) String() string {
	movesSliceString := fmt.Sprintf("%s", pv.moves)
	return movesSliceString[1 : len(movesSliceString)-1]
}
