package define

import (
	"bytes"
	"encoding/binary"
)

// BlockIndex is a integer that ranging from 0 to 4095,
// and can be decode and represents as a  relative position
// to a sub chunk.
type BlockIndex uint16

// X returns the X-axis relative coordinates of this block with respect to the sub chunk.
func (b BlockIndex) X() uint8 {
	return uint8(b >> 8)
}

// Y returns the Y-axis relative coordinates of this block with respect to the sub chunk.
func (b BlockIndex) Y() uint8 {
	return uint8((b >> 4) & 15)
}

// Z returns the Z-axis relative coordinates of this block with respect to the sub chunk.
func (b BlockIndex) Z() uint8 {
	return uint8(b & 15)
}

// UpdateIndex computes the index of the block in the sub
// chunk based on the given relative coordinates of x, y,
// and z. Then, updates the index of this block.
func (b *BlockIndex) UpdateIndex(x uint8, y uint8, z uint8) {
	*b = BlockIndex(uint16(z) | (uint16(y) << 4) | (uint16(x) << 8))
}

// Marshal encode BlockIndex to its bytes represents, and writes to bytes buffer.
func (b BlockIndex) Marshal(buf *bytes.Buffer) {
	result := make([]byte, 2)
	binary.LittleEndian.PutUint16(result, uint16(b))
	buf.Write(result)
}

// Unmarshal decode a BlockIndex from the underlying bytes buffer.
func (b *BlockIndex) Unmarshal(buf *bytes.Buffer) {
	temp := make([]byte, 2)
	_, _ = buf.Read(temp)
	*b = BlockIndex(binary.LittleEndian.Uint16(temp))
}

// ChunkBlockIndex holds two (u)int16 integers
// which the first one is ranging from 0 to 4095,
// and the second one is the sub chunk Y position
// of this block.
//
// Those two integers could be decode and represents
// as a block relative position to the chunk.
type ChunkBlockIndex struct {
	blockIndex      BlockIndex
	inWhichSubChunk int16
}

// X returns the X-axis relative coordinates of this block with respect to the chunk.
func (c ChunkBlockIndex) X() uint8 {
	return c.blockIndex.X()
}

// Y returns the Y-axis relative coordinates of this block with respect to the chunk.
func (c ChunkBlockIndex) Y() int16 {
	return int16(c.blockIndex.Y()) + (c.inWhichSubChunk << 4)
}

// Z returns the Z-axis relative coordinates of this block with respect to the chunk.
func (c ChunkBlockIndex) Z() uint8 {
	return c.blockIndex.Z()
}

// UpdateIndex computes the index of the block in the chunk
// based on the given relative coordinates of x, y, and z.
// Then, updates the index of this block.
func (c *ChunkBlockIndex) UpdateIndex(x uint8, y int16, z uint8) {
	c.inWhichSubChunk = y >> 4
	relativeY := uint8(y - (c.inWhichSubChunk << 4))
	c.blockIndex.UpdateIndex(x, relativeY, z)
}

// Marshal encode ChunkBlockIndex to its bytes represents, and writes to bytes buffer.
func (c ChunkBlockIndex) Marshal(buf *bytes.Buffer) {
	c.blockIndex.Marshal(buf)
	result := make([]byte, 2)
	binary.LittleEndian.PutUint16(result, uint16(c.inWhichSubChunk))
	buf.Write(result)
}

// Unmarshal decode a ChunkBlockIndex from the underlying bytes buffer.
func (c *ChunkBlockIndex) Unmarshal(buf *bytes.Buffer) {
	c.blockIndex.Unmarshal(buf)
	temp := make([]byte, 2)
	_, _ = buf.Read(temp)
	c.inWhichSubChunk = int16(binary.LittleEndian.Uint16(temp))
}
