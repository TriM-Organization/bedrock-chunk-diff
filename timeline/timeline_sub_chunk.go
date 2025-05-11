package timeline

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/marshal"
	"github.com/TriM-Organization/bedrock-world-operator/chunk"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

const DefaultMaxLimit = 7

// SubChunkTimeline records the timeline of a sub chunk,
// and it contains the change logs about this sub chunk
// on this timeline.
// In other words, the SubChunkTimeline holds the history
// of this sub chunk.
//
// Note that it's unsafe for multiple thread to access this
// struct due to we don't use mutex to ensure the operation
// is atomic.
//
// So, it's your responsibility to make ensure there is only
// one thread is using this object.
type SubChunkTimeline struct {
	db          DB
	pos         define.DimSubChunk
	releaseFunc func()
	isEmpty     bool

	timelineUnixTime []int64
	blockPalette     *define.BlockPalette

	ptr             uint
	currentSubChunk define.Layers
	currentNBT      []define.NBTWithIndex

	barrierLeft  uint
	barrierRight uint
	maxLimit     uint

	latestSubChunk define.Layers
	latestNBT      []define.NBTWithIndex
}

// NewSubChunkTimeline gets the timeline of a sub chunk who is at pos.
//
// Note that if timeline of current sub chunk is not exist, then we will not create a timeline
// but return an empty one so you can modify it. The time to create the timeline is only when you
// save a timeline that not empty to the database.
//
// subChunkIndex is an integer that bigger then -1.
// For example, if a block is at (x,23,z) and is in Overworld, then it is in a sub chunk
// whose Y position is 23>>4 = 1. However, this is not the index of this sub chunk,
// we need use (23>>4) - (dm.Range()[0]>>4) to get the index, which is 1-4=-3.
//
// Important:
//
//   - Once any modifications have been made to the returned timeline, you must save them
//     at the end; otherwise, the timeline will not be able to maintain data consistency
//     (only need to save at the last modification).
//
//   - Timeline of one sub chunk can't be using by multiple threads. Therefore, you will
//     get blocking when a thread calling NewSubChunkTimeline but there is still some
//     threads are using target sub chunk.
//
//   - Calling SubChunkTimeline.Save to release the timeline.
//
//   - Returned SubChunkTimeline can't shared with multiple threads, and it's your responsibility
//     to ensure this thing.
func (t *TimelineDB) NewSubChunkTimeline(pos define.DimSubChunk) (result *SubChunkTimeline, err error) {
	var success bool

	releaseFunc, succ := t.sessions.Require(pos)
	if !succ {
		return nil, fmt.Errorf("NewSubChunkTimeline: Underlying database is closed")
	}

	defer func() {
		if !success {
			releaseFunc()
		}
	}()

	result = &SubChunkTimeline{
		db:           t.DB,
		pos:          pos,
		releaseFunc:  releaseFunc,
		blockPalette: define.NewBlockPalette(),
		maxLimit:     DefaultMaxLimit,
	}

	payload := t.Get(
		define.Sum(pos, define.KeySubChunkExistStates),
	)
	if len(payload) == 0 {
		result.isEmpty = true
		success = true
		return result, nil
	}

	// Timeline Unix Time
	{
		payload := t.Get(
			define.Sum(pos, define.KeyTimelineUnixTime),
		)
		for len(payload) > 0 {
			result.timelineUnixTime = append(result.timelineUnixTime, int64(binary.LittleEndian.Uint64(payload)))
			payload = payload[8:]
		}
	}

	// Block Palette
	{
		blockPaletteBytes := t.Get(
			define.Sum(pos, []byte(define.KeyBlockPalette)...),
		)
		buf := bytes.NewBuffer(blockPaletteBytes)

		for buf.Len() > 0 {
			var m map[string]any
			if err := nbt.NewDecoderWithEncoding(buf, nbt.LittleEndian).Decode(&m); err != nil {
				return nil, fmt.Errorf("NewSubChunkTimeline: error decoding block palette entry: %w", err)
			}

			blockRuntimeID, err := chunk.BlockPaletteEncoding.DecodeBlockState(m)
			if err != nil {
				return nil, fmt.Errorf("NewSubChunkTimeline: %v", err)
			}

			result.blockPalette.AddBlock(blockRuntimeID)
		}
	}

	// Barrier and Max limit
	{
		payload := t.Get(
			define.Sum(pos, []byte(define.KeyBarrierAndLimit)...),
		)
		if len(payload) < 12 {
			return nil, fmt.Errorf("NewSubChunkTimeline: Barrier and limit is broken (only get %d bytes but expected 12)", len(payload))
		}
		result.barrierLeft = uint(binary.LittleEndian.Uint32(payload))
		result.ptr = result.barrierLeft
		result.barrierRight = uint(binary.LittleEndian.Uint32(payload[4:]))
		result.maxLimit = uint(binary.LittleEndian.Uint32(payload[8:]))
	}

	// Latest Sub Chunk
	{
		latestSubChunkBytes := t.Get(
			define.Sum(pos, define.KeyLatestSubChunk),
		)

		blockMatrix, err := marshal.BytesToLayers(latestSubChunkBytes)
		if err != nil {
			return nil, fmt.Errorf("NewSubChunkTimeline: %v", err)
		}

		result.latestSubChunk = blockMatrix
	}

	// Latest NBT
	{
		latestNBTBytes := t.Get(
			define.Sum(pos, []byte(define.KeyLatestNBT)...),
		)

		latestNBT, err := marshal.BytesToBlockNBT(latestNBTBytes)
		if err != nil {
			return nil, fmt.Errorf("NewSubChunkTimeline: %v", err)
		}

		result.latestNBT = latestNBT
	}

	success = true
	return result, nil
}

// DeleteSubChunkTimeline delete the timeline of sub chunk who at pos.
// If timeline is not exist, then do no operation.
//
// Time complexity: O(n).
// n is the time point that this sub chunk have.
func (t *TimelineDB) DeleteSubChunkTimeline(pos define.DimSubChunk) error {
	var success bool

	timeline, err := t.NewSubChunkTimeline(pos)
	if err != nil {
		return fmt.Errorf("DeleteSubChunkTimeline: %v", err)
	}
	defer func() {
		timeline.releaseFunc()
	}()

	if timeline.isEmpty {
		return nil
	}

	transaction, err := t.OpenTransaction()
	if err != nil {
		return fmt.Errorf("DeleteSubChunkTimeline: %v", err)
	}
	defer func() {
		if !success {
			_ = transaction.Discard()
			return
		}
		_ = transaction.Commit()
	}()

	// Exist states
	err = transaction.Delete(define.Sum(pos, define.KeySubChunkExistStates))
	if err != nil {
		return fmt.Errorf("DeleteSubChunkTimeline: %v", err)
	}

	// Timeline Unix Time
	err = transaction.Delete(define.Sum(pos, define.KeyTimelineUnixTime))
	if err != nil {
		return fmt.Errorf("DeleteSubChunkTimeline: %v", err)
	}

	// Block Palette
	err = transaction.Delete(define.Sum(pos, []byte(define.KeyBlockPalette)...))
	if err != nil {
		return fmt.Errorf("DeleteSubChunkTimeline: %v", err)
	}

	// Barrier and Max limit
	err = transaction.Delete(define.Sum(pos, []byte(define.KeyBarrierAndLimit)...))
	if err != nil {
		return fmt.Errorf("DeleteSubChunkTimeline: %v", err)
	}

	// Latest Sub Chunk
	err = transaction.Delete(define.Sum(pos, define.KeyLatestSubChunk))
	if err != nil {
		return fmt.Errorf("DeleteSubChunkTimeline: %v", err)
	}

	// Latest NBT
	err = transaction.Delete(define.Sum(pos, []byte(define.KeyLatestNBT)...))
	if err != nil {
		return fmt.Errorf("DeleteSubChunkTimeline: %v", err)
	}

	// Each delta update
	for i := timeline.barrierLeft; i <= timeline.barrierRight; i++ {
		err = transaction.Delete(define.IndexBlockDu(pos, i))
		if err != nil {
			return fmt.Errorf("DeleteSubChunkTimeline: %v", err)
		}
		err = transaction.Delete(define.IndexNBTDu(pos, i))
		if err != nil {
			return fmt.Errorf("DeleteSubChunkTimeline: %v", err)
		}
	}

	success = true
	return nil
}
