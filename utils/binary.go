package utils

import "encoding/binary"

// Uint32BinaryAdd decode a little endian uint32Bytes as uint32,
// and computes the result of uint32+delta, and then return its
// little endian bytes result. If len(uint32Bytes) < 4, then use
// defaultIfNotExist instead.
func Uint32BinaryAdd(uint32Bytes []byte, defaultIfNotExist []byte, delta int32) []byte {
	if len(uint32Bytes) < 4 {
		uint32Bytes = defaultIfNotExist
	}

	originCount := binary.LittleEndian.Uint32(uint32Bytes)
	originCount = uint32(int32(originCount) + delta)

	newer := make([]byte, 4)
	binary.LittleEndian.PutUint32(newer, originCount)

	return newer
}
