package marshal

import (
	"bytes"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// BlockMatrixToBytes write the bytes represents of blockMatrix into a bytes buffer.
func BlockMatrixToBytes(buf *bytes.Buffer, blockMatrix define.BlockMatrix) {
	if define.BlockMatrixIsEmpty(blockMatrix) {
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
	result := define.NewBlockMatrix()

	for i := range define.MatrixSize {
		var value int32
		r.Varint32(&value)
		result[i] = value
	}

	return result
}

// DiffMatrixToBytes writes the bytes represents of diffMatrix into a bytes buffer.
func DiffMatrixToBytes(buf *bytes.Buffer, diffMatrix define.DiffMatrix) {
	length := uint16(len(diffMatrix))

	if length == 0 {
		buf.WriteByte(MatrixStateEmpty)
		return
	}
	buf.WriteByte(MatrixStateNotEmpty)

	w := protocol.NewWriter(buf, 0)
	w.Uint16(&length)

	for _, value := range diffMatrix {
		w.Uint16((*uint16)(&value.Index))
		w.Varint32(&value.NewPaletteID)
	}
}

// BytesToDiffMatrix decode DiffMatrix from bytes buffer.
func BytesToDiffMatrix(buf *bytes.Buffer) (result define.DiffMatrix) {
	var length uint16

	b, _ := buf.ReadByte()
	if b == MatrixStateEmpty {
		return nil
	}

	r := protocol.NewReader(buf, 0, false)
	r.Uint16(&length)

	for range length {
		var idx define.BlockIndex
		var newPaletteID int32
		r.Uint16((*uint16)(&idx))
		r.Varint32(&newPaletteID)
		result = append(result, define.SingleBlockDiff{
			Index:        idx,
			NewPaletteID: newPaletteID,
		})
	}

	return
}
