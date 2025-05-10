package marshal

import (
	"bytes"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// BlockMatrixToBytes write the bytes represents of blockMatrix into a bytes buffer.
func BlockMatrixToBytes(buf *bytes.Buffer, blockMatrix define.BlockMatrix) {
	if blockMatrix == emptyBlockMatrix {
		buf.WriteByte(MatrixStateEmpty)
		return
	}
	buf.WriteByte(MatrixStateNotEmpty)

	w := protocol.NewWriter(buf, 0)
	for i := range define.MatrixSize {
		value := uint32(blockMatrix[i])
		w.Varuint32(&value)
	}
}

// BytesToBlockMatrix decode BlockMatrix from bytes buffer.
func BytesToBlockMatrix(buf *bytes.Buffer) define.BlockMatrix {
	var result define.BlockMatrix

	b, _ := buf.ReadByte()
	if b == MatrixStateEmpty {
		return result
	}

	r := protocol.NewReader(buf, 0, false)
	for i := range define.MatrixSize {
		var value uint32
		r.Varuint32(&value)
		result[i] = uint16(value)
	}

	return result
}

// DiffMatrixToBytes writes the bytes represents of diffMatrix into a bytes buffer.
func DiffMatrixToBytes(buf *bytes.Buffer, diffMatrix define.DiffMatrix) {
	if diffMatrix == emptyDiffMatrix {
		buf.WriteByte(MatrixStateEmpty)
		return
	}
	buf.WriteByte(MatrixStateNotEmpty)

	w := protocol.NewWriter(buf, 0)
	for i := range define.MatrixSize {
		value := int32(diffMatrix[i])
		w.Varint32(&value)
	}
}

// BytesToDiffMatrix decode DiffMatrix from bytes buffer.
func BytesToDiffMatrix(buf *bytes.Buffer) define.DiffMatrix {
	var result define.DiffMatrix

	b, _ := buf.ReadByte()
	if b == MatrixStateEmpty {
		return result
	}

	r := protocol.NewReader(buf, 0, false)
	for i := range define.MatrixSize {
		var value int32
		r.Varint32(&value)
		result[i] = int(value)
	}

	return result
}
