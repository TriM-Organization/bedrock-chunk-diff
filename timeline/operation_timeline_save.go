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

// Save saves current timeline into the underlying database,
// and also release current timeline.
//
// Read only timeline should also calling Save to release the
// resource. But read only timeline calling this function will
// only release but don't do further operation.
// Additionally, empty non read only timeline is also follow the
// same behavior.
// Note that you could use s.Empty() and s.ReadOnly() to check.
//
// If you calling Save and get a nil error,
// then this timeline is released and can't be used again.
// Also, you can't call Save multiple times.
//
// But, if Save returned non-nil error, then this object
// will not released.
//
// Note that we will not check whether it has been released,
// nor will we check whether you have called Save multiple times.
//
// Save must calling at the last modification of the timeline;
// otherwise, the timeline will not be able to maintain data consistency.
func (s *ChunkTimeline) Save() error {
	var success bool

	if s.isEmpty || s.isReadOnly {
		s.releaseFunc()
		return nil
	}

	tran, err := s.db.OpenTransaction()
	if err != nil {
		return fmt.Errorf("(s *ChunkTimeline) Save: %v", err)
	}
	defer func() {
		if !success {
			_ = tran.Discard()
			return
		}
		_ = tran.Commit()
		s.releaseFunc()
	}()

	globalData := bytes.NewBuffer(nil)

	// Chunk Index
	{
		keyBytes := define.Index(s.pos)
		bucket := tran.(*transaction).tx.Bucket(DatabaseKeyChunkIndex)

		if bucket.Get(keyBytes) == nil {
			err = bucket.Put(
				DatabaseKeyChunkCount,
				utils.Uint32BinaryAdd(bucket.Get(DatabaseKeyChunkCount), make([]byte, 4), 1),
			)
			if err != nil {
				return fmt.Errorf("(s *ChunkTimeline) Save: %v", err)
			}

			err = bucket.Put(keyBytes, []byte{1})
			if err != nil {
				return fmt.Errorf("(s *ChunkTimeline) Save: %v", err)
			}
		}
	}

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
		err = tran.Put(
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
		err = tran.Put(
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
		err = tran.Put(
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
		err = tran.Put(
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
