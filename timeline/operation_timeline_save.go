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

// Save saves current timeline into the underlying database.
// Save must calling at the last modification of the timeline;
// otherwise, the timeline will not be able to maintain data consistency.
func (s *SubChunkTimeline) Save() error {
	var success bool

	if s.isEmpty {
		return nil
	}

	transaction, err := s.db.OpenTransaction()
	if err != nil {
		return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
	}
	defer func() {
		if !success {
			transaction.Discard()
			return
		}
		_ = transaction.Commit()
	}()

	err = transaction.Put(
		define.Sum(s.dm, s.position, s.subChunkIndex, define.KeySubChunkExistStates),
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
			define.Sum(s.dm, s.position, s.subChunkIndex, define.KeyTimelineUnixTime),
			buf.Bytes(),
		)
		if err != nil {
			return fmt.Errorf("(s *SubChunkTimeline) Save: %v", err)
		}
	}

	// Block Palette
	{
		buf := bytes.NewBuffer(nil)

		for _, value := range s.blockPalette {
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
			err = transaction.Delete(define.Sum(s.dm, s.position, s.subChunkIndex, []byte(define.KeyBlockPalette)...))
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Save: %v", err)
			}
		} else {
			err = transaction.Put(
				define.Sum(s.dm, s.position, s.subChunkIndex, []byte(define.KeyBlockPalette)...),
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
			define.Sum(s.dm, s.position, s.subChunkIndex, []byte(define.KeyBarrierAndLimit)...),
			result,
		)
		if err != nil {
			return fmt.Errorf("(s *SubChunkTimeline) Save: %v", err)
		}
	}

	// Latest Sub Chunk
	{
		payload := marshal.LayersToBytes(s.latestSubChunk)
		if len(payload) == 0 {
			err = transaction.Delete(define.Sum(s.dm, s.position, s.subChunkIndex, define.KeyLatestSubChunk))
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Save: %v", err)
			}
		} else {
			err = transaction.Put(
				define.Sum(s.dm, s.position, s.subChunkIndex, define.KeyLatestSubChunk),
				payload,
			)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Save: %v", err)
			}
		}
	}

	// Latest NBT
	{
		payload := marshal.BlockNBTBytes(s.latestNBT)
		if len(payload) == 0 {
			err = transaction.Delete(define.Sum(s.dm, s.position, s.subChunkIndex, []byte(define.KeyLatestNBT)...))
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Save: %v", err)
			}
		} else {
			err = transaction.Put(
				define.Sum(s.dm, s.position, s.subChunkIndex, []byte(define.KeyLatestNBT)...),
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
