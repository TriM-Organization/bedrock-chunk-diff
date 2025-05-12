package define

import (
	"encoding/binary"
)

// Keys on a per-chunk basis.
// These are prefixed by only the chunk coordinates.
const (
	KeyChunkGlobalData = "tbplrg"

	KeyBlockDeltaUpdate = "du"
	KeyNBTDeltaUpdate   = "du'"

	KeyLatestTimePointUnixTime = 'T'
	KeyLatestChunk             = 'm'
	KeyLatestNBT               = "m'"
)

// Index returns a bytes holding the written index of the chunk position passed.
//
// Different from standard Minecraft world, we write the x and z position of this
// chunk first, and then 2 bytes to represents the dimension id of this chunk.
//
// Therefore, we use and return 10 bytes in total.
func Index(pos DimChunk) []byte {
	x, z, dim := uint32(pos.ChunkPos[0]), uint32(pos.ChunkPos[1]), uint16(pos.Dimension)
	b := make([]byte, 10)

	binary.LittleEndian.PutUint32(b, x)
	binary.LittleEndian.PutUint32(b[4:], z)
	binary.LittleEndian.PutUint16(b[8:], dim)

	return b
}

// Sum converts Index(pos) to its []byte representation and appends p.
// Note that Sum is very necessary because all Sum do is preventing users
// from believing that "append" can create new slices (however, it not).
func Sum(pos DimChunk, p ...byte) []byte {
	return append(Index(pos), p...)
}
