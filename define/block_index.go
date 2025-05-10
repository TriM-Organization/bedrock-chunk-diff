package define

// SubChunkBlockIndex is an integer that range from 0 to 4095,
// and could be decode and represents as a block relative
// position to the sub chunk.
type SubChunkBlockIndex uint16

// X returns the X-axis relative coordinates of this block with respect to the sub chunk.
func (s SubChunkBlockIndex) X() uint8 {
	return uint8(s >> 8)
}

// Y returns the Y-axis relative coordinates of this block with respect to the sub chunk.
func (s SubChunkBlockIndex) Y() uint8 {
	return uint8((s - ((s >> 8) << 8)) >> 4)
}

// Z returns the Z-axis relative coordinates of this block with respect to the sub chunk.
func (s SubChunkBlockIndex) Z() uint8 {
	return uint8(s - ((s >> 4) << 4))
}

// GetIndex returns the index of this block in the sub chunk.
// The index is an integer ranging from 0 to 4095 and can be
// decoded as the relative block coordinates of this block with
// respect to this sub chunk.
func (s SubChunkBlockIndex) GetIndex() uint16 {
	return uint16(s)
}

// UpdateIndex computes the index of the block in the sub chunk
// based on the given relative coordinates of x, y, and z.
// Then, updates the index of this block.
func (s *SubChunkBlockIndex) UpdateIndex(x uint8, y uint8, z uint8) {
	*s = SubChunkBlockIndex(uint16(x)*256 + uint16(y)*16 + uint16(z))
}
