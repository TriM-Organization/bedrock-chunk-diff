package define

import (
	"encoding/binary"

	"github.com/TriM-Organization/bedrock-world-operator/define"
)

// Keys on a per-sub-chunk basis.
// These are prefixed by only the chunk coordinates and subchunk ID.
const (
	KeySubChunkExistStates = '?'
	KeyTimelineUnixTime    = 't'

	KeyBlockDeltaUpdate = "du"
	KeyNBTDeltaUpdate   = "du'"

	KeyBlockPalette    = "bp"
	KeyBarrierAndLimit = "lrg"

	KeyLatestSubChunk = 'm'
	KeyLatestNBT      = "m'"
)

// Index returns a bytes holding the written index of the sub chunk position passed.
//
// Different from standard Minecraft world, we write subChunkIndex first, and then
// is the x and z position of this sub chunk.
// Additionally, we always write the dimension id and only use two bytes for it.
//
// Therefore, we use and return 11 bytes in total.
func Index(dm define.Dimension, position define.ChunkPos, subChunkIndex uint8) []byte {
	x, z, dim := uint32(position[0]), uint32(position[1]), uint16(dm)
	b := make([]byte, 11)

	b[0] = subChunkIndex
	binary.LittleEndian.PutUint32(b[1:], x)
	binary.LittleEndian.PutUint32(b[5:], z)
	binary.LittleEndian.PutUint16(b[9:], dim)

	return b
}

// IndexBlockDu returns a bytes holding the written index of the sub chunk position passed,
// but specially for blocks delta update used key to index.
func IndexBlockDu(dm define.Dimension, position define.ChunkPos, subChunkIndex uint8, timeID uint) []byte {
	timeIDBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(timeIDBytes, uint32(timeID))
	return append(
		Sum(
			dm, position, subChunkIndex,
			[]byte(KeyBlockDeltaUpdate)...,
		),
		timeIDBytes...,
	)
}

// IndexNBTDu returns a bytes holding the written index of the sub chunk position passed,
// but specially for NBTs delta update used key to index.
func IndexNBTDu(dm define.Dimension, position define.ChunkPos, subChunkIndex uint8, timeID uint) []byte {
	timeIDBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(timeIDBytes, uint32(timeID))
	return append(
		Sum(
			dm, position, subChunkIndex,
			[]byte(KeyNBTDeltaUpdate)...,
		),
		timeIDBytes...,
	)
}

// Sum converts Index(dm, position, subChunkIndex) to its []byte representation and appends p.
// Note that Sum is very necessary because all Sum do is preventing users from believing that
// "append" can create new slices (however, it not).
func Sum(dm define.Dimension, position define.ChunkPos, subChunkIndex uint8, p ...byte) []byte {
	return append(
		Index(dm, position, subChunkIndex),
		p...,
	)
}
