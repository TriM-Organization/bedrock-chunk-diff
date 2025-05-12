package marshal

import (
	"bytes"
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/utils"
	operator_define "github.com/TriM-Organization/bedrock-world-operator/define"
)

// ChunkMatrixToBytes return the bytes represents of chunkMatrix.
func ChunkMatrixToBytes(chunkMatrix define.ChunkMatrix) (result []byte, err error) {
	buf := bytes.NewBuffer(nil)

	for _, value := range chunkMatrix {
		LayersToBytes(buf, value)
	}

	if buf.Len() == 0 {
		return nil, nil
	}

	result, err = utils.Gzip(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("ChunkMatrixToBytes: %v", err)
	}
	return
}

// BytesToChunkMatrix decode ChunkMatrix from bytes.
// r is the count of sub chunks that this chunk have.
func BytesToChunkMatrix(in []byte, r operator_define.Range) (result define.ChunkMatrix, err error) {
	result = make(define.ChunkMatrix, (r.Height()>>4)+1)

	if len(in) == 0 {
		return result, nil
	}

	originBytes, err := utils.Ungzip(in)
	if err != nil {
		err = fmt.Errorf("BytesToChunkMatrix: %v", err)
		return
	}

	ptr := 0
	buf := bytes.NewBuffer(originBytes)
	for buf.Len() > 0 {
		result[ptr] = BytesToLayers(buf)
		ptr++
	}

	return result, nil
}

// ChunkDiffMatrixToBytes return the bytes represents of chunkDiffMatrix.
func ChunkDiffMatrixToBytes(chunkDiffMatrix define.ChunkDiffMatrix) (result []byte, err error) {
	buf := bytes.NewBuffer(nil)

	for _, value := range chunkDiffMatrix {
		LayersDiffToBytes(buf, value)
	}

	if buf.Len() == 0 {
		return nil, nil
	}

	result, err = utils.Gzip(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("ChunkDiffMatrixToBytes: %v", err)
	}
	return
}

// BytesToChunkDiffMatrix decode ChunkDiffMatrix from bytes.
// r is the count of sub chunks that this chunk have.
func BytesToChunkDiffMatrix(in []byte, r operator_define.Range) (result define.ChunkDiffMatrix, err error) {
	result = make(define.ChunkDiffMatrix, (r.Height()>>4)+1)

	if len(in) == 0 {
		return result, nil
	}

	originBytes, err := utils.Ungzip(in)
	if err != nil {
		err = fmt.Errorf("BytesToChunkDiffMatrix: %v", err)
		return
	}

	ptr := 0
	buf := bytes.NewBuffer(originBytes)
	for buf.Len() > 0 {
		result[ptr] = BytesToLayersDiff(buf)
		ptr++
	}

	return result, nil
}
