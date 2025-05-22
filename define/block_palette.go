package define

import (
	"github.com/TriM-Organization/bedrock-world-operator/block"
)

// BlockPalette is the block palette for a chunk timeline.
// All time point in this line will share the same palette.
type BlockPalette struct {
	bp      []uint32
	mapping map[uint32]uint32
}

// NewBlockPalette creates a new block palette that only have a air block.
// Technically speaking, air does not actually exist, but it can be found by agreement.
func NewBlockPalette() *BlockPalette {
	return &BlockPalette{
		mapping: make(map[uint32]uint32),
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

	b.bp = append(b.bp, blockRuntimeID)
	b.mapping[blockRuntimeID] = uint32(len(b.bp))
}

// BlockPaletteIndex finds the index of blockRuntimeID in block palette.
// If not exist, then added it the underlying block palette.
//
// Returned index is the real index plus 1.
// If you got 0, then that means this is an air block.
// We don't save air block in block palette, and you should to pay attention to it.
func (b *BlockPalette) BlockPaletteIndex(blockRuntimeID uint32) uint32 {
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
// Will not check if blockPaletteIndex is out of index (if out of index, then runtime panic).
func (b *BlockPalette) BlockRuntimeID(blockPaletteIndex uint32) uint32 {
	if blockPaletteIndex == 0 {
		return block.AirRuntimeID
	}
	return b.bp[blockPaletteIndex-1]
}

// BlockPaletteLen returns the length of underlying block palette.
func (b *BlockPalette) BlockPaletteLen() int {
	return len(b.bp)
}

// BlockPalette gets the deep copy of underlying block palette.
// The block palette is actually multiple block runtime IDs.
func (b *BlockPalette) BlockPalette() []uint32 {
	newOne := make([]uint32, len(b.bp))
	copy(newOne, b.bp)
	return newOne
}

// SetBlockPalette sets the underlying block palette to newPalette.
// We assume all elements (block runtime IDs) in newPalette is unique
// and all is valid Minecraft standard blocks.
func (b *BlockPalette) SetBlockPalette(newPalette []uint32) {
	b.bp = newPalette
	b.mapping = make(map[uint32]uint32)
	for key, value := range b.bp {
		b.mapping[value] = uint32(key)
	}
}
