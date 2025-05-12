package marshal

import (
	"bytes"
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/utils"
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
func BytesToChunkMatrix(in []byte) (result define.ChunkMatrix, err error) {
	if len(in) == 0 {
		return result, nil
	}

	originBytes, err := utils.Ungzip(in)
	if err != nil {
		err = fmt.Errorf("BytesToChunkMatrix: %v", err)
		return
	}

	buf := bytes.NewBuffer(originBytes)
	for buf.Len() > 0 {
		result = append(result, BytesToLayers(buf))
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
func BytesToChunkDiffMatrix(in []byte) (result define.ChunkDiffMatrix, err error) {
	if len(in) == 0 {
		return result, nil
	}

	originBytes, err := utils.Ungzip(in)
	if err != nil {
		err = fmt.Errorf("BytesToChunkDiffMatrix: %v", err)
		return
	}

	buf := bytes.NewBuffer(originBytes)
	for buf.Len() > 0 {
		result = append(result, BytesToLayersDiff(buf))
	}

	return result, nil
}
