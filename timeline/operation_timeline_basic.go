package timeline

import "github.com/TriM-Organization/bedrock-world-operator/block"

// Empty returns whether this timeline is empty or not.
// If is empty, then calling Save will result in no operation.
func (s *SubChunkTimeline) Empty() bool {
	return s.isEmpty
}

// SetMaxLimit sets the timeline could record how many time point.
// maxLimit must bigger than 0. If less, then set the limit to 1.
//
// Note that calling SetMaxLimit will not change the empty states
// of this timeline.
func (s *SubChunkTimeline) SetMaxLimit(maxLimit uint) {
	s.maxLimit = max(maxLimit, 1)
}

// BlockPaletteIndex finds the index of blockRuntimeID in block palette.
// If not exist, then added it the underlying block palette.
//
// Returned index is the real index plus 1.
// If you got 0, then that means this is an air block.
// We don't save air block in block palette, and you should to pay attention to it.
func (s *SubChunkTimeline) BlockPaletteIndex(blockRuntimeID uint32) uint16 {
	if blockRuntimeID == block.AirRuntimeID {
		return 0
	}

	idx, ok := s.blockPaletteMapping[blockRuntimeID]
	if ok {
		return idx
	}

	name, states, found := block.RuntimeIDToState(blockRuntimeID)
	if !found {
		name = "minecraft:unknown"
		states = make(map[string]any)
	}

	blockRuntimeID, _ = block.StateToRuntimeID(name, states)
	indexState, _ := block.RuntimeIDToIndexState(blockRuntimeID)

	s.blockPalette = append(s.blockPalette, indexState)
	idx = uint16(len(s.blockPaletteMapping) + 1)
	s.blockPaletteMapping[blockRuntimeID] = idx

	return idx
}

// BlockRuntimeID return the block runtime ID that crresponding to blockPaletteIndex.
// If target is unknown, then return the runtime ID of minecraft:unknown block.
// Will not check if blockPaletteIndex is out of index (if out of index, then runtime panic).
func (s *SubChunkTimeline) BlockRuntimeID(blockPaletteIndex uint16) uint32 {
	if blockPaletteIndex == 0 {
		return block.AirRuntimeID
	}

	blockRuntimeID, found := block.IndexStateToRuntimeID(s.blockPalette[blockPaletteIndex-1])
	if !found {
		return block.ComputeBlockHash("minecraft:unknown", map[string]any{})
	}

	return blockRuntimeID
}
