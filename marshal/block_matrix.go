package marshal

import (
	"bytes"
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/utils"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// BlockMatrixToBytes return the bytes represents of blockMatrix.
func BlockMatrixToBytes(blockMatrix define.BlockMatrix) []byte {
	buf := bytes.NewBuffer(nil)
	w := protocol.NewWriter(buf, 0)

	for i := range define.MatrixSize {
		value := uint32(blockMatrix[i])
		w.Varuint32(&value)
	}

	return utils.Gzip(buf.Bytes())
}

// BytesToBlockMatrix decode BlockMatrix from bytes.
func BytesToBlockMatrix(in []byte) (result define.BlockMatrix, err error) {
	originBytes, err := utils.Ungzip(in)
	if err != nil {
		err = fmt.Errorf("BytesToBlockMatrix: %v", err)
		return
	}

	buf := bytes.NewBuffer(originBytes)
	r := protocol.NewReader(buf, 0, false)

	for i := range define.MatrixSize {
		var value uint32
		r.Varuint32(&value)
		result[i] = uint16(value)
	}
	return
}

// DiffMatrixToBytes return the bytes represents of diffMatrix.
func DiffMatrixToBytes(diffMatrix define.DiffMatrix) []byte {
	buf := bytes.NewBuffer(nil)
	w := protocol.NewWriter(buf, 0)

	for i := range define.MatrixSize {
		value := int32(diffMatrix[i])
		w.Varint32(&value)
	}

	return utils.Gzip(buf.Bytes())
}

// BytesToDiffMatrix decode DiffMatrix from bytes.
func BytesToDiffMatrix(in []byte) (result define.DiffMatrix, err error) {
	originBytes, err := utils.Ungzip(in)
	if err != nil {
		err = fmt.Errorf("BytesToDiffMatrix: %v", err)
		return
	}

	buf := bytes.NewBuffer(originBytes)
	r := protocol.NewReader(buf, 0, false)

	for i := range define.MatrixSize {
		var value int32
		r.Varint32(&value)
		result[i] = int(value)
	}
	return
}
