package define

import (
	"encoding/binary"

	"github.com/TriM-Organization/bedrock-world-operator/define"
)

// Keys on a per-sub-chunk basis.
// These are prefixed by only the chunk coordinates and subchunk ID.
const (
	KeySubChunkExistStates = '?'
	KeyDeltaUpdate         = "du"

	KeyBlockPalette = "bp"
	KeyBarrierLeft  = 'l'
	KeyBarrierRight = 'r'
	KeyMaxLimit     = 'g'

	KeyLatestSubChunk = 'm'
	KeyLatestNBT      = "m'"
)

// Index returns a byte buffer holding the written index of the sub chunk position passed.
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

// Sum converts Index(dm, position, subChunkIndex) to its []byte representation and appends p.
// Note that Sum is very necessary because all Sum do is preventing users from believing that
// "append" can create new slices (however, it not).
func Sum(dm define.Dimension, position define.ChunkPos, subChunkIndex uint8, p ...byte) []byte {
	return append(
		Index(dm, position, subChunkIndex),
		p...,
	)
}
