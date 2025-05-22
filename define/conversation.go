package define

import (
	"github.com/TriM-Organization/bedrock-world-operator/block"
	"github.com/TriM-Organization/bedrock-world-operator/chunk"
	"github.com/TriM-Organization/bedrock-world-operator/define"
)

// ChunkToMatrix converts a chunk to its chunk martix represents.
// blockPalette is the block palette of this chunk timeline used.
func ChunkToMatrix(c *chunk.Chunk, blockPalette *BlockPalette) (result ChunkMatrix) {
	for _, value := range c.Sub() {
		l := Layers{}

		if value.Empty() {
			if len(value.Layers()) == 0 {
				result = append(result, l)
				continue
			}
			_ = l.Layer(0)
			result = append(result, l)
			continue
		}

		for index, layer := range value.Layers() {
			newerBlockMartrix := NewBlockMatrix()

			ptr := 0
			for x := range uint8(16) {
				for y := range uint8(16) {
					for z := range uint8(16) {
						newerBlockMartrix[ptr] = blockPalette.BlockPaletteIndex(layer.At(x, y, z))
						ptr++
					}
				}
			}

			_ = l.Layer(index)
			l[index] = newerBlockMartrix
		}

		result = append(result, l)
	}

	return
}

// MatrixToChunk converts the chunk matrix to its chunk represents.
// r is the range of this chunk, and blockPalette is this chunk matrix used.
func MatrixToChunk(matrix ChunkMatrix, r define.Range, blockPalette *BlockPalette) (c *chunk.Chunk) {
	c = chunk.NewChunk(block.AirRuntimeID, r)
	sub := c.Sub()

	for subChunkIndex, subChunkLayers := range matrix {
		subChunk := sub[subChunkIndex]

		for layerIndex, blockMatrix := range subChunkLayers {
			subChunkLayer := subChunk.Layer(uint8(layerIndex))

			if BlockMatrixIsEmpty(blockMatrix) {
				continue
			}

			ptr := 0
			for x := range uint8(16) {
				for y := range uint8(16) {
					for z := range uint8(16) {
						subChunkLayer.Set(x, y, z, blockPalette.BlockRuntimeID(blockMatrix[ptr]))
						ptr++
					}
				}
			}
		}
	}

	return
}

// FromChunkNBT converts nbts to []NBTWithIndex, chunkPos is the position of this chunk.
// For all NBT block that not in this chunk (or their NBT is broken), then we will ignore.
// Note that the returned slice is not the deep copied NBTs.
func FromChunkNBT(chunkPos define.ChunkPos, nbts []map[string]any) (result []NBTWithIndex) {
	for _, value := range nbts {
		x, ok1 := value["x"].(int32)
		y, ok2 := value["y"].(int32)
		z, ok3 := value["z"].(int32)

		if !ok1 || !ok2 || !ok3 {
			continue
		}

		nbtWithIndex := NBTWithIndex{}

		xBlock, zBlock := chunkPos[0]<<4, chunkPos[1]<<4

		deltaX := x - xBlock
		deltaZ := z - zBlock
		if deltaX < 0 || deltaX > 15 || deltaZ < 0 || deltaZ > 15 {
			continue
		}

		nbtWithIndex.Index.UpdateIndex(uint8(x-xBlock), int16(y), uint8(z-zBlock))
		nbtWithIndex.NBT = value
		result = append(result, nbtWithIndex)
	}

	return
}

// ToChunkNBT converts nbts to []map[string]any.
// Note that the returned slice is not the deep copied NBTs.
func ToChunkNBT(nbts []NBTWithIndex) (result []map[string]any) {
	for _, value := range nbts {
		result = append(result, value.NBT)
	}
	return
}
