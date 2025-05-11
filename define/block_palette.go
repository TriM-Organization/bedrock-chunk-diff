package define

import (
	"github.com/TriM-Organization/bedrock-world-operator/block"
	block_general "github.com/TriM-Organization/bedrock-world-operator/block/general"
)

// BlockPalette is the block palette for a sub chunk timeline.
// All time point in this line will share the same palette.
type BlockPalette struct {
	bp      []block_general.IndexBlockState
	mapping map[uint32]uint16
}

// NewBlockPalette creates a new block palette that only have a air block.
// Technically speaking, air does not actually exist, but it can be found by agreement.
func NewBlockPalette() *BlockPalette {
	return &BlockPalette{
		mapping: make(map[uint32]uint16),
	}
}

// AddBlock adds the block whose block runtime id is blockRuntimeID
// to the underlying block palette.
// If is exist or the given block is air, then do no operation.
func (b *BlockPalette) AddBlock(blockRuntimeID uint32) {
	if blockRuntimeID == block.AirRuntimeID {
		return
	}

	_, ok := b.mapping[blockRuntimeID]
	if ok {
		return
	}

	name, states, found := block.RuntimeIDToState(blockRuntimeID)
	if !found {
		name = "minecraft:unknown"
		states = make(map[string]any)
	}

	newerblockRuntimeID, _ := block.StateToRuntimeID(name, states)
	indexState, _ := block.RuntimeIDToIndexState(newerblockRuntimeID)

	b.bp = append(b.bp, indexState)
	b.mapping[blockRuntimeID] = uint16(len(b.bp))
}

// BlockPaletteIndex finds the index of blockRuntimeID in block palette.
// If not exist, then added it the underlying block palette.
//
// Returned index is the real index plus 1.
// If you got 0, then that means this is an air block.
// We don't save air block in block palette, and you should to pay attention to it.
func (b *BlockPalette) BlockPaletteIndex(blockRuntimeID uint32) uint16 {
	if blockRuntimeID == block.AirRuntimeID {
		return 0
	}

	idx, ok := b.mapping[blockRuntimeID]
	if ok {
		return idx
	}

	b.AddBlock(blockRuntimeID)
	return b.mapping[blockRuntimeID]
}

// BlockRuntimeID return the block runtime ID that crresponding to blockPaletteIndex.
// If target is unknown, then return the runtime ID of minecraft:unknown block.
// Will not check if blockPaletteIndex is out of index (if out of index, then runtime panic).
func (b *BlockPalette) BlockRuntimeID(blockPaletteIndex uint16) uint32 {
	if blockPaletteIndex == 0 {
		return block.AirRuntimeID
	}
	blockRuntimeID, _ := block.IndexStateToRuntimeID(b.bp[blockPaletteIndex-1])
	return blockRuntimeID
}

// BlockPaletteLen returns the length of underlying block palette.
func (b *BlockPalette) BlockPaletteLen() int {
	return len(b.bp)
}

// BlockPalette gets the deep copy of underlying block palette.
func (b *BlockPalette) BlockPalette() []block_general.IndexBlockState {
	newOne := make([]block_general.IndexBlockState, len(b.bp))
	copy(newOne, b.bp)
	return newOne
}

// SetBlockPalette sets the underlying block palette to newPalette.
// We assume all elements in newPalette is unique and all is valid Minecraft standard blocks.
func (b *BlockPalette) SetBlockPalette(newPalette []block_general.IndexBlockState) {
	b.bp = nil
	b.mapping = make(map[uint32]uint16)

	for _, value := range newPalette {
		blockRuntimeID, _ := block.IndexStateToRuntimeID(value)
		b.AddBlock(blockRuntimeID)
	}
}
