package define

import (
	"bytes"
	"encoding/binary"
)

// ChunkBlockIndex holds two (u)int16 integers
// which the first one is ranging from 0 to 400,
// and the second one is the sub chunk Y position
// of this block.
//
// Those two integers could be decode and represents
// as a block relative position to the chunk.
type ChunkBlockIndex struct {
	indexInChunk uint16
	inWhichChunk int16
}

// X returns the X-axis relative coordinates of this block with respect to the chunk.
func (c ChunkBlockIndex) X() uint8 {
	return uint8(c.indexInChunk >> 8)
}

// Y returns the Y-axis relative coordinates of this block with respect to the chunk.
func (c ChunkBlockIndex) Y() int16 {
	idx := int16(c.indexInChunk)
	return ((idx - ((idx >> 8) << 8)) >> 4) + int16(c.inWhichChunk)<<4
}

// Z returns the Z-axis relative coordinates of this block with respect to the chunk.
func (c ChunkBlockIndex) Z() uint8 {
	return uint8(c.indexInChunk - ((c.indexInChunk >> 4) << 4))
}

// UpdateIndex computes the index of the block in the chunk
// based on the given relative coordinates of x, y, and z.
// Then, updates the index of this block.
func (c *ChunkBlockIndex) UpdateIndex(x uint8, y int16, z uint8) {
	c.inWhichChunk = y >> 4
	relativeY := y - int16(c.inWhichChunk)<<4
	c.indexInChunk = uint16(x)*256 + uint16(relativeY)*16 + uint16(z)
}

// Marshal encode ChunkBlockIndex to its bytes represents, and writes to bytes buffer.
func (c ChunkBlockIndex) Marshal(buf *bytes.Buffer) {
	result := make([]byte, 4)
	binary.LittleEndian.PutUint16(result, c.indexInChunk)
	binary.LittleEndian.PutUint16(result[2:], uint16(c.inWhichChunk))
	buf.Write(result)
}

// Unmarshal decode a ChunkBlockIndex from the underlying bytes buffer.
func (c *ChunkBlockIndex) Unmarshal(buf *bytes.Buffer) {
	temp := make([]byte, 2)
	_, _ = buf.Read(temp)
	c.indexInChunk = binary.LittleEndian.Uint16(temp)
	_, _ = buf.Read(temp)
	c.inWhichChunk = int16(binary.LittleEndian.Uint16(temp))
}
