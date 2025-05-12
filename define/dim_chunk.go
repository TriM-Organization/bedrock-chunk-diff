package define

import (
	"encoding/binary"

	"github.com/TriM-Organization/bedrock-world-operator/define"
)

// DimChunk ..
type DimChunk struct {
	Dimension define.Dimension
	ChunkPos  define.ChunkPos
}

// IndexBlockDu returns a bytes holding the written index of the chunk position passed,
// but specially for blocks delta update used key to index.
func IndexBlockDu(pos DimChunk, timeID uint) []byte {
	timeIDBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(timeIDBytes, uint32(timeID))
	return append(
		Sum(pos, []byte(KeyBlockDeltaUpdate)...),
		timeIDBytes...,
	)
}

// IndexNBTDu returns a bytes holding the written index of the chunk position passed,
// but specially for NBTs delta update used key to index.
func IndexNBTDu(pos DimChunk, timeID uint) []byte {
	timeIDBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(timeIDBytes, uint32(timeID))
	return append(
		Sum(pos, []byte(KeyNBTDeltaUpdate)...),
		timeIDBytes...,
	)
}
