package define

import (
	"encoding/binary"
)

// Keys on a per-sub-chunk basis.
// These are prefixed by only the chunk coordinates and subchunk ID.
const (
	KeySubChunkExistStates     = '?'
	KeyTimelineUnixTime        = 't'
	KeyLatestTimePointUnixTime = 'T'

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
func Index(pos DimSubChunk) []byte {
	x, z, dim := uint32(pos.ChunkPos[0]), uint32(pos.ChunkPos[1]), uint16(pos.Dimension)
	b := make([]byte, 11)

	b[0] = pos.SubChunkIndex
	binary.LittleEndian.PutUint32(b[1:], x)
	binary.LittleEndian.PutUint32(b[5:], z)
	binary.LittleEndian.PutUint16(b[9:], dim)

	return b
}

// Sum converts Index(pos) to its []byte representation and appends p.
// Note that Sum is very necessary because all Sum do is preventing users from believing that
// "append" can create new slices (however, it not).
func Sum(pos DimSubChunk, p ...byte) []byte {
	return append(Index(pos), p...)
}
