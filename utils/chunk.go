package utils

import (
	"bytes"
	"fmt"

	"github.com/TriM-Organization/bedrock-world-operator/block"
	"github.com/TriM-Organization/bedrock-world-operator/chunk"
	"github.com/TriM-Organization/bedrock-world-operator/define"
)

// FromDiskChunkPayload ..
func FromDiskChunkPayload(subChunks [][]byte, r define.Range) (c *chunk.Chunk, err error) {
	c, err = chunk.DiskDecode(chunk.SerialisedData{
		SubChunks: subChunks,
		Biomes:    nil,
	}, r)
	if err != nil {
		return nil, fmt.Errorf("FromDiskChunkPayload: %v", err)
	}
	return
}

// FromNetworkChunkPayload ..
func FromNetworkChunkPayload(subChunks [][]byte, r define.Range) (c *chunk.Chunk, err error) {
	buf := bytes.NewBuffer(nil)
	for _, value := range subChunks {
		buf.Write(value)
	}

	c, err = chunk.NetworkDecode(block.AirRuntimeID, buf.Bytes(), len(subChunks), r)
	if err != nil {
		return nil, fmt.Errorf("FromNetworkChunkPayload: %v", err)
	}

	return
}

// ChunkPayload ..
func ChunkPayload(c *chunk.Chunk, e chunk.Encoding) (subChunks [][]byte, r define.Range) {
	serialisedData := chunk.Encode(c, e)
	return serialisedData.SubChunks, c.Range()
}
