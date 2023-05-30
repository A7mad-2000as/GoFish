package chessEngine

const MovesCapacity = 1024

type MoveList struct {
	Moves [MovesCapacity]Move
	Size  uint8
}

func (moveList *MoveList) AddMove(move Move) {
	moveList.Moves[moveList.Size] = move
	moveList.Size++
}
