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

// SaveNOP releases current sub chunk timeline, and
// don't do more things (will not change ths database).
func (s *SubChunkTimeline) SaveNOP() {
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
func (s *SubChunkTimeline) Save() error {
	var success bool

	if s.isEmpty {
		s.releaseFunc()
		return nil
	}

	transaction, err := s.db.OpenTransaction()
	if err != nil {
		return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
	}
	defer func() {
		if !success {
			_ = transaction.Discard()
			return
		}
		_ = transaction.Commit()
		s.releaseFunc()
	}()

	err = transaction.Put(
		define.Sum(s.pos, define.KeySubChunkExistStates),
		[]byte{1},
	)
	if err != nil {
		return fmt.Errorf("(s *SubChunkTimeline) Save: %v", err)
	}

	// Timeline Unix Time
	{
		buf := bytes.NewBuffer(nil)

		for _, value := range s.timelineUnixTime {
			temp := make([]byte, 8)
			binary.LittleEndian.PutUint64(temp, uint64(value))
			buf.Write(temp)
		}

		err = transaction.Put(
			define.Sum(s.pos, define.KeyTimelineUnixTime),
			buf.Bytes(),
		)
		if err != nil {
			return fmt.Errorf("(s *SubChunkTimeline) Save: %v", err)
		}
	}

	// Latest Time Point Unix Time
	if len(s.timelineUnixTime) > 0 {
		latestTimePointUnixTime := s.timelineUnixTime[len(s.timelineUnixTime)-1]
		if latestTimePointUnixTime == 0 {
			err = transaction.Delete(define.Sum(s.pos, define.KeyLatestTimePointUnixTime))
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Save: %v", err)
			}
		} else {
			latestTimePointUnixTimeBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(
				latestTimePointUnixTimeBytes,
				uint64(latestTimePointUnixTime),
			)
			err = transaction.Put(
				define.Sum(s.pos, define.KeyLatestTimePointUnixTime),
				latestTimePointUnixTimeBytes,
			)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Save: %v", err)
			}
		}
	}

	// Block Palette
	{
		buf := bytes.NewBuffer(nil)

		for _, value := range s.blockPalette.BlockPalette() {
			blockRuntimeID, found := block.IndexStateToRuntimeID(value)
			if !found {
				blockRuntimeID = block.ComputeBlockHash("minecraft:unknown", map[string]any{})
			}
			name, states, _ := block.RuntimeIDToState(blockRuntimeID)
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

		if buf.Len() == 0 {
			err = transaction.Delete(define.Sum(s.pos, []byte(define.KeyBlockPalette)...))
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Save: %v", err)
			}
		} else {
			err = transaction.Put(
				define.Sum(s.pos, []byte(define.KeyBlockPalette)...),
				buf.Bytes(),
			)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Save: %v", err)
			}
		}
	}

	// Barrier and Max limit
	{
		result := make([]byte, 12)

		binary.LittleEndian.PutUint32(result, uint32(s.barrierLeft))
		binary.LittleEndian.PutUint32(result[4:], uint32(s.barrierRight))
		binary.LittleEndian.PutUint32(result[8:], uint32(s.maxLimit))

		err = transaction.Put(
			define.Sum(s.pos, []byte(define.KeyBarrierAndLimit)...),
			result,
		)
		if err != nil {
			return fmt.Errorf("(s *SubChunkTimeline) Save: %v", err)
		}
	}

	// Latest Sub Chunk
	{
		payload, err := marshal.LayersToBytes(s.latestSubChunk)
		if err != nil {
			return fmt.Errorf("(s *SubChunkTimeline) Save: %v", err)
		}
		if len(payload) == 0 {
			err = transaction.Delete(define.Sum(s.pos, define.KeyLatestSubChunk))
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Save: %v", err)
			}
		} else {
			err = transaction.Put(
				define.Sum(s.pos, define.KeyLatestSubChunk),
				payload,
			)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Save: %v", err)
			}
		}
	}

	// Latest NBT
	{
		payload, err := marshal.BlockNBTBytes(s.latestNBT)
		if err != nil {
			return fmt.Errorf("(s *SubChunkTimeline) Save: %v", err)
		}
		if len(payload) == 0 {
			err = transaction.Delete(define.Sum(s.pos, []byte(define.KeyLatestNBT)...))
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Save: %v", err)
			}
		} else {
			err = transaction.Put(
				define.Sum(s.pos, []byte(define.KeyLatestNBT)...),
				payload,
			)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Save: %v", err)
			}
		}
	}

	success = true
	return nil
}
