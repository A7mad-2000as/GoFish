package chessEngine

import "time"

const (
	MinimumExpectedPliesLeft                 = 10
	AverageExpectedPliesLeft                 = 40
	MinimumMoveTimeAllocation                = 100
	AverageAdjustMoveTime                    = 150
	MoveAllocatedTimeOfRemainingTimeFraction = 8
)

type DefaultTimeManager struct {
	remainingTime     int64
	increment         int64
	moveTime          int64
	movesToGo         int16
	depth             uint8
	nodeCount         uint64
	endSearch         bool
	moveAllocatedTime int64
	searchStopInstant time.Time
}

func (timeManager *DefaultTimeManager) Initialize(remainingTime int64, increment int64, moveTime int64, movesToGo int16, depth uint8, nodeCount uint64) {
	timeManager.remainingTime = remainingTime
	timeManager.increment = increment
	timeManager.moveTime = moveTime
	timeManager.movesToGo = movesToGo
	timeManager.depth = depth
	timeManager.nodeCount = nodeCount
}

func (timeManager *DefaultTimeManager) StartMoveTimeAllocation(plyNumber uint16) {
	timeManager.endSearch = false

	if timeManager.moveTime != 0 {
		timeManager.remainingTime = 0
		timeManager.searchStopInstant = time.Now().Add(time.Duration(timeManager.moveTime) * time.Millisecond)
		return
	}

	if timeManager.remainingTime < 0 {
		return
	}

	moveAllocatedTime := int64(0)
	if timeManager.movesToGo != 0 {
		moveAllocatedTime = timeManager.remainingTime / int64(timeManager.movesToGo)
	} else {
		if timeManager.increment > 0 {
			moveAllocatedTime = timeManager.remainingTime / max(MinimumExpectedPliesLeft, AverageExpectedPliesLeft-int64(plyNumber))
		} else {
			moveAllocatedTime = timeManager.remainingTime / AverageExpectedPliesLeft
		}
	}

	moveAllocatedTime += (3 * timeManager.increment) / 4

	if moveAllocatedTime >= timeManager.remainingTime {
		moveAllocatedTime = timeManager.remainingTime - AverageAdjustMoveTime
	}

	if moveAllocatedTime <= 0 {
		moveAllocatedTime = MinimumMoveTimeAllocation
	}

	timeManager.searchStopInstant = time.Now().Add(time.Duration(moveAllocatedTime) * time.Millisecond)
	timeManager.moveAllocatedTime = moveAllocatedTime
}

func (timeManager *DefaultTimeManager) ChangeMoveAllocatedTime(newMoveAllocatedTime int64) {
	if timeManager.movesToGo != 0 {
		return
	}

	if newMoveAllocatedTime >= timeManager.remainingTime/MoveAllocatedTimeOfRemainingTimeFraction {
		newMoveAllocatedTime = timeManager.remainingTime / MoveAllocatedTimeOfRemainingTimeFraction
	}

	timeManager.moveAllocatedTime = newMoveAllocatedTime
	timeManager.searchStopInstant = time.Now().Add(time.Duration(newMoveAllocatedTime) * time.Millisecond)
}

func (timeManager *DefaultTimeManager) SetMoveTimeIsUp() {
	if timeManager.remainingTime < 0 || timeManager.endSearch {
		return
	}

	if time.Now().After(timeManager.searchStopInstant) {
		timeManager.endSearch = true
	}
}
