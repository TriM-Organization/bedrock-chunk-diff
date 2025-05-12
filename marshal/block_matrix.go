package marshal

import (
	"bytes"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// BlockMatrixToBytes write the bytes represents of blockMatrix into a bytes buffer.
func BlockMatrixToBytes(buf *bytes.Buffer, blockMatrix define.BlockMatrix) {
	if define.MatrixIsEmpty(blockMatrix) {
		buf.WriteByte(MatrixStateEmpty)
		return
	}
	buf.WriteByte(MatrixStateNotEmpty)

	w := protocol.NewWriter(buf, 0)
	for i := range define.MatrixSize {
		w.Varint32(&blockMatrix[i])
	}
}

// BytesToBlockMatrix decode BlockMatrix from bytes buffer.
func BytesToBlockMatrix(buf *bytes.Buffer) define.BlockMatrix {
	b, _ := buf.ReadByte()
	if b == MatrixStateEmpty {
		return nil
	}

	r := protocol.NewReader(buf, 0, false)
	result := define.NewMatrix[define.BlockMatrix]()

	for i := range define.MatrixSize {
		var value int32
		r.Varint32(&value)
		result[i] = value
	}

	return result
}

// DiffMatrixToBytes writes the bytes represents of diffMatrix into a bytes buffer.
func DiffMatrixToBytes(buf *bytes.Buffer, diffMatrix define.DiffMatrix) {
	if define.MatrixIsEmpty(diffMatrix) {
		buf.WriteByte(MatrixStateEmpty)
		return
	}
	buf.WriteByte(MatrixStateNotEmpty)

	w := protocol.NewWriter(buf, 0)
	for i := range define.MatrixSize {
		w.Varint32(&diffMatrix[i])
	}
}

// BytesToDiffMatrix decode DiffMatrix from bytes buffer.
func BytesToDiffMatrix(buf *bytes.Buffer) define.DiffMatrix {
	b, _ := buf.ReadByte()
	if b == MatrixStateEmpty {
		return nil
	}

	r := protocol.NewReader(buf, 0, false)
	result := define.NewMatrix[define.DiffMatrix]()

	for i := range define.MatrixSize {
		var value int32
		r.Varint32(&value)
		result[i] = value
	}

	return result
}
