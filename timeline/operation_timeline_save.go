package timeline

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/marshal"
	"github.com/TriM-Organization/bedrock-chunk-diff/utils"
	"github.com/TriM-Organization/bedrock-world-operator/block"
	"github.com/TriM-Organization/bedrock-world-operator/chunk"
)

// SaveNOP releases current chunk timeline, and don't
// do more things (will not change ths database).
func (s *ChunkTimeline) SaveNOP() {
	s.releaseFunc()
}

// Save saves current timeline into the underlying database,
// and also release current timeline.
//
// That means, if you calling Save and get a nil error,
// then this timeline is released and can't be used again.
// Also, you can't call Save multiple times.
//
// But, if Save returned non-nil error, then this object
// will not released.
//
// Note that we will not check whether it has been released,
// nor will we check whether you have called Save multiple times.
//
// Additionally, if current timeline is marked as empty,
// then calling Save will only release this object and don't do
// further operation. Note that you could use s.Empty() to check.
//
// Save must calling at the last modification of the timeline;
// otherwise, the timeline will not be able to maintain data consistency.
func (s *ChunkTimeline) Save() error {
	var success bool

	if s.isEmpty {
		s.releaseFunc()
		return nil
	}

	transaction, err := s.db.OpenTransaction()
	if err != nil {
		return fmt.Errorf("(s *ChunkTimeline) Pop: %v", err)
	}
	defer func() {
		if !success {
			_ = transaction.Discard()
			return
		}
		_ = transaction.Commit()
		s.releaseFunc()
	}()

	globalData := bytes.NewBuffer(nil)

	// Timeline Unix Time
	{
		buf := bytes.NewBuffer(nil)

		for _, value := range s.timelineUnixTime {
			temp := make([]byte, 8)
			binary.LittleEndian.PutUint64(temp, uint64(value))
			buf.Write(temp)
		}

		lengthBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(lengthBytes, uint32(buf.Len()))

		globalData.Write(lengthBytes)
		globalData.Write(buf.Bytes())
	}

	// Block Palette
	{
		buf := bytes.NewBuffer(nil)

		for _, value := range s.blockPalette.BlockPalette() {
			name, states, found := block.RuntimeIDToState(value)
			if !found {
				name = "minecraft:unknown"
			}
			utils.MarshalNBT(
				buf,
				map[string]any{
					"name":    name,
					"states":  states,
					"version": chunk.CurrentBlockVersion,
				},
				"",
			)
		}

		lengthBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(lengthBytes, uint32(buf.Len()))

		globalData.Write(lengthBytes)
		globalData.Write(buf.Bytes())
	}

	// Barrier and Max limit
	{
		result := make([]byte, 12)

		binary.LittleEndian.PutUint32(result, uint32(s.barrierLeft))
		binary.LittleEndian.PutUint32(result[4:], uint32(s.barrierRight))
		binary.LittleEndian.PutUint32(result[8:], uint32(s.maxLimit))

		globalData.Write(result)
	}

	// Save global data
	{
		gzipBytes, err := utils.Gzip(globalData.Bytes())
		if err != nil {
			return fmt.Errorf("(s *ChunkTimeline) Save: %v", err)
		}
		err = transaction.Put(
			define.Sum(s.pos, []byte(define.KeyChunkGlobalData)...),
			gzipBytes,
		)
		if err != nil {
			return fmt.Errorf("(s *ChunkTimeline) Save: %v", err)
		}
	}

	// Latest Time Point Unix Time
	{
		latestTimePointUnixTimeBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(
			latestTimePointUnixTimeBytes,
			uint64(s.timelineUnixTime[len(s.timelineUnixTime)-1]),
		)
		err = transaction.Put(
			define.Sum(s.pos, define.KeyLatestTimePointUnixTime),
			latestTimePointUnixTimeBytes,
		)
		if err != nil {
			return fmt.Errorf("(s *ChunkTimeline) Save: %v", err)
		}
	}

	// Latest Chunk
	{
		payload, err := marshal.ChunkMatrixToBytes(s.latestChunk)
		if err != nil {
			return fmt.Errorf("(s *ChunkTimeline) Save: %v", err)
		}
		err = transaction.Put(
			define.Sum(s.pos, define.KeyLatestChunk),
			payload,
		)
		if err != nil {
			return fmt.Errorf("(s *ChunkTimeline) Save: %v", err)
		}
	}

	// Latest NBT
	{
		payload, err := marshal.BlockNBTBytes(s.latestNBT)
		if err != nil {
			return fmt.Errorf("(s *ChunkTimeline) Save: %v", err)
		}
		err = transaction.Put(
			define.Sum(s.pos, []byte(define.KeyLatestNBT)...),
			payload,
		)
		if err != nil {
			return fmt.Errorf("(s *ChunkTimeline) Save: %v", err)
		}
	}

	success = true
	return nil
}
