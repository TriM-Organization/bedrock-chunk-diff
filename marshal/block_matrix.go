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
		w.Varuint32(&blockMatrix[i])
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
		r.Varuint32(&result[i])
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
		w.Varuint32(&value.IndexDelta)
	}
	for _, value := range diffMatrix {
		w.Varuint32(&value.NewPaletteID)
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
	result = make([]define.SingleBlockDiff, length)

	for index := range length {
		r.Varuint32(&result[index].IndexDelta)
	}
	for index := range length {
		r.Varuint32(&result[index].NewPaletteID)
	}

	return
}
