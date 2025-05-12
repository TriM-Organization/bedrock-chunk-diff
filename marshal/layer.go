package marshal

import (
	"bytes"
	"encoding/binary"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
)

// LayersToBytes writes the bytes represents of layers into a bytes buffer.
func LayersToBytes(buf *bytes.Buffer, layers define.Layers) {
	lengthBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthBytes, uint32(len(layers)))
	buf.Write(lengthBytes)
	for _, value := range layers {
		BlockMatrixToBytes(buf, value)
	}
}

// BytesToLayers decode Layers from bytes buffer.
func BytesToLayers(buf *bytes.Buffer) define.Layers {
	lengthBytes := make([]byte, 4)
	_, _ = buf.Read(lengthBytes)
	length := int(binary.LittleEndian.Uint32(lengthBytes))

	result := define.Layers{}
	for i := range length {
		result.Layer(i)
		result[i] = BytesToBlockMatrix(buf)
	}

	return result
}

// LayersDiffToBytes writes the bytes represents of layersDiff into a bytes buffer.
func LayersDiffToBytes(buf *bytes.Buffer, layersDiff define.LayersDiff) {
	lengthBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthBytes, uint32(len(layersDiff)))
	buf.Write(lengthBytes)
	for _, value := range layersDiff {
		DiffMatrixToBytes(buf, value)
	}
}

// BytesToLayersDiff decode LayersDiff from bytes buffer.
func BytesToLayersDiff(buf *bytes.Buffer) define.LayersDiff {
	lengthBytes := make([]byte, 4)
	_, _ = buf.Read(lengthBytes)
	length := int(binary.LittleEndian.Uint32(lengthBytes))

	result := define.LayersDiff{}
	for i := range length {
		result.Layer(i)
		result[i] = BytesToDiffMatrix(buf)
	}

	return result
}
