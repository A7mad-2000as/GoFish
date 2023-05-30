package chessEngine

const (
	DefaultTableSize = 64 * 1024 * 1024
	EntriesPerIndex  = 2
	EntrySize        = 16

	UpperBoundEntryType uint8 = 1
	LowerBoundEntryType uint8 = 2
	ExactEntryType      uint8 = 3

	MateThreshold = 9000
)

type TableEntry struct {
	HashValue     uint64
	BestMove      Move
	DepthOfSearch uint8
	Score         int16
	EntryInfo     uint8
}

type DefaultTranspositionTable struct {
	entries      []TableEntry
	numOfEntries uint64
}

func (entry TableEntry) GetEntryType() uint8 {
	return entry.EntryInfo & 0x03
}

func (entry TableEntry) GetEntryAge() uint8 {
	return (entry.EntryInfo & 0x0c) >> 2
}

func (entry *TableEntry) SetEntryType(entryType uint8) {
	entry.EntryInfo &= 0xfc
	entry.EntryInfo |= entryType
}

func (entry *TableEntry) SetEntryAge(entryAge uint8) {
	entry.EntryInfo &= 0xf3
	entry.EntryInfo |= (entryAge << 2)
}

func (entry *TableEntry) ReadEntryInfo(ttMove *Move, hash uint64, pliesFromRoot uint8, requiredDepth uint8, alpha int16, beta int16) (int16, bool) {
	if entry.HashValue == hash {
		entryScore := entry.Score
		*ttMove = entry.BestMove

		if entry.DepthOfSearch >= requiredDepth {
			if entryScore > MateThreshold {
				entryScore -= int16(pliesFromRoot)
			} else if entryScore < -MateThreshold {
				entryScore += int16(pliesFromRoot)
			}

			switch entry.GetEntryType() {
			case UpperBoundEntryType:
				if entryScore <= alpha {
					return alpha, true
				}
			case LowerBoundEntryType:
				if entryScore >= beta {
					return beta, true
				}
			case ExactEntryType:
				return entryScore, true
			}
		}

		return entryScore, false
	}

	return int16(0), false
}

func (entry *TableEntry) ModifyTableEntry(move Move, searchScore int16, hash uint64, pliesFromRoot uint8, requiredDepth uint8, entryType uint8, entryAge uint8) {
	entry.HashValue = hash
	entry.BestMove = move
	entry.DepthOfSearch = requiredDepth
	entry.SetEntryType(entryType)
	entry.SetEntryAge(entryAge)

	if searchScore > MateThreshold {
		searchScore += int16(pliesFromRoot)
	} else if searchScore < -MateThreshold {
		searchScore -= int16(pliesFromRoot)
	}
	entry.Score = searchScore
}

func (table *DefaultTranspositionTable) ResizeTable(tableSize uint64, entrySize uint64) {
	table.numOfEntries = tableSize / entrySize
	table.entries = make([]TableEntry, table.numOfEntries)
}

func (table *DefaultTranspositionTable) DeleteEntries() {
	table.entries = nil
	table.numOfEntries = 0
}

func (table *DefaultTranspositionTable) ClearEntries() {
	for i := uint64(0); i < table.numOfEntries; i++ {
		table.entries[i] = *new(TableEntry)
	}
}

func (table *DefaultTranspositionTable) GetEntryToRead(hash uint64) *TableEntry {
	tableIndex := hash % table.numOfEntries

	if tableIndex == table.numOfEntries-1 {
		return &table.entries[tableIndex]
	}

	firstEntry := table.entries[tableIndex]
	if firstEntry.HashValue == hash {
		return &table.entries[tableIndex]
	}
	return &table.entries[tableIndex+1]
}

func (table *DefaultTranspositionTable) GetEntryToReplace(hash uint64, requiredDepth uint8, ageState uint8) *TableEntry {
	tableIndex := hash % table.numOfEntries

	if tableIndex == table.numOfEntries-1 {
		return &table.entries[tableIndex]
	}

	firstEntry := table.entries[tableIndex]
	if firstEntry.GetEntryAge() != ageState || firstEntry.DepthOfSearch <= requiredDepth {
		return &table.entries[tableIndex]
	}
	return &table.entries[tableIndex+1]
}
