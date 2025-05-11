package marshal

import (
	"bytes"
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/utils"
)

// LayersToBytes return the bytes represents of layers.
func LayersToBytes(layers define.Layers) (result []byte, err error) {
	buf := bytes.NewBuffer(nil)

	for _, value := range layers {
		BlockMatrixToBytes(buf, value)
	}

	if buf.Len() == 0 {
		return nil, nil
	}

	result, err = utils.Gzip(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("LayersToBytes: %v", err)
	}
	return
}

// BytesToLayers decode Layers from bytes.
func BytesToLayers(in []byte) (result define.Layers, err error) {
	if len(in) == 0 {
		return result, nil
	}

	originBytes, err := utils.Ungzip(in)
	if err != nil {
		err = fmt.Errorf("BytesToLayers: %v", err)
		return
	}

	buf := bytes.NewBuffer(originBytes)
	for buf.Len() > 0 {
		result = append(result, BytesToBlockMatrix(buf))
	}

	return result, nil
}

// LayersDiffToBytes return the bytes represents of layersDiff.
func LayersDiffToBytes(layersDiff define.LayersDiff) (result []byte, err error) {
	buf := bytes.NewBuffer(nil)

	for _, value := range layersDiff {
		DiffMatrixToBytes(buf, value)
	}

	if buf.Len() == 0 {
		return nil, nil
	}

	result, err = utils.Gzip(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("LayersDiffToBytes: %v", err)
	}
	return
}

// BytesToLayersDiff decode LayersDiff from bytes.
func BytesToLayersDiff(in []byte) (result define.LayersDiff, err error) {
	if len(in) == 0 {
		return result, nil
	}

	originBytes, err := utils.Ungzip(in)
	if err != nil {
		err = fmt.Errorf("BytesToLayersDiff: %v", err)
		return
	}

	buf := bytes.NewBuffer(originBytes)
	for buf.Len() > 0 {
		result = append(result, BytesToDiffMatrix(buf))
	}

	return result, nil
}
