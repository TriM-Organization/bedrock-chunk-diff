package utils

import (
	"bytes"
	"fmt"

	"github.com/TriM-Organization/bedrock-world-operator/block"
	"github.com/TriM-Organization/bedrock-world-operator/chunk"
	"github.com/TriM-Organization/bedrock-world-operator/define"
)

// FromChunkPayload ..
func FromChunkPayload(subChunks [][]byte, r define.Range, e chunk.Encoding) (c *chunk.Chunk, err error) {
	c = chunk.NewChunk(block.AirRuntimeID, r)

	for index, value := range subChunks {
		subChunk, _, err := chunk.DecodeSubChunk(bytes.NewBuffer(value), r, e)
		if err != nil {
			return nil, fmt.Errorf("FromDiskChunkPayload: %v", err)
		}
		c.SetSubChunk(subChunk, int16(index))
	}

	return
}

// ChunkPayload ..
func ChunkPayload(c *chunk.Chunk, e chunk.Encoding) (subChunks [][]byte, r define.Range) {
	serialisedData := chunk.Encode(c, e)
	return serialisedData.SubChunks, c.Range()
}
