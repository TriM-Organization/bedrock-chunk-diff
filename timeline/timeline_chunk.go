package timeline

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/marshal"
	"github.com/TriM-Organization/bedrock-chunk-diff/utils"
	"github.com/TriM-Organization/bedrock-world-operator/chunk"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"go.etcd.io/bbolt"
)

const DefaultMaxLimit = 7

// ChunkTimeline records the timeline of a chunk,
// and it contains the change logs about this chunk
// on this timeline.
//
// In other words, the ChunkTimeline holds the history
// of this chunk.
//
// Note that it's unsafe for multiple thread to access this
// struct due to we don't use mutex to ensure the operation
// is atomic.
//
// So, it's your responsibility to make ensure there is only
// one thread is using this object.
type ChunkTimeline struct {
	db          DB
	pos         define.DimChunk
	releaseFunc func()

	isReadOnly bool
	isEmpty    bool

	timelineUnixTime []int64
	blockPalette     *define.BlockPalette

	ptr          uint
	currentChunk define.ChunkMatrix
	currentNBT   []define.NBTWithIndex

	barrierLeft  uint
	barrierRight uint
	maxLimit     uint

	latestChunk define.ChunkMatrix
	latestNBT   []define.NBTWithIndex
}

// NewChunkTimeline gets the timeline of a chunk who is at pos.
//
// Note that if timeline of current chunk is not exist, then we will not create a timeline
// but return an empty one so you can modify it. The time to create the timeline is only when you
// save a timeline that not empty to the database.
//
// If readOnly is true, then returned a timeline but only can read.
// For a read only timeline, you also need use ChunkTimeline.Save to release it.
//
// Important:
//
//   - Once any modifications have been made to the returned timeline, you must save them
//     at the end; otherwise, the timeline will not be able to maintain data consistency
//     (only need to save at the last modification).
//
//   - Timeline of one chunk can't be using by multiple threads. Therefore, you will
//     get blocking when a thread calling NewChunkTimeline but there is still some
//     threads are using target chunk.
//
//   - Calling ChunkTimeline.Save to release the timeline.
//
//   - Returned ChunkTimeline can't shared with multiple threads, and it's your responsibility
//     to ensure this thing.
func (t *TimelineDB) NewChunkTimeline(pos define.DimChunk, readOnly bool) (result *ChunkTimeline, err error) {
	var exist bool
	var success bool

	releaseFunc, succ := t.sessions.Require(pos)
	if !succ {
		return nil, fmt.Errorf("NewChunkTimeline: Underlying database is closed")
	}

	defer func() {
		if !success {
			releaseFunc()
		}
	}()

	result = &ChunkTimeline{
		db:           t.DB,
		pos:          pos,
		isReadOnly:   readOnly,
		releaseFunc:  releaseFunc,
		blockPalette: define.NewBlockPalette(),
		currentChunk: make(define.ChunkMatrix, pos.Dimension.Height()>>4),
		maxLimit:     DefaultMaxLimit,
		latestChunk:  make(define.ChunkMatrix, pos.Dimension.Height()>>4),
	}

	err = t.DB.(*database).bdb.View(func(tx *bbolt.Tx) error {
		exist = (tx.Bucket(DatabaseKeyChunkIndex).Get(define.Index(pos)) != nil)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("NewChunkTimeline: %v", err)
	}
	if !exist {
		result.isEmpty = true
		success = true
		return result, nil
	}

	gzippedGlobalData := t.Get(
		define.Sum(pos, []byte(define.KeyChunkGlobalData)...),
	)
	globalData, err := utils.Ungzip(gzippedGlobalData)
	if err != nil {
		return nil, fmt.Errorf("NewChunkTimeline: %v", err)
	}

	// Timeline Unix Time
	{
		length := binary.LittleEndian.Uint32(globalData)
		payload := globalData[4 : 4+length]
		for len(payload) > 0 {
			result.timelineUnixTime = append(result.timelineUnixTime, int64(binary.LittleEndian.Uint64(payload)))
			payload = payload[8:]
		}
		globalData = globalData[4+length:]
	}

	// Block Palette
	{
		length := binary.LittleEndian.Uint32(globalData)
		payload := globalData[4 : 4+length]
		buf := bytes.NewBuffer(payload)

		for buf.Len() > 0 {
			var m map[string]any
			if err := nbt.NewDecoderWithEncoding(buf, nbt.LittleEndian).Decode(&m); err != nil {
				return nil, fmt.Errorf("NewChunkTimeline: error decoding block palette entry: %w", err)
			}

			blockRuntimeID, err := chunk.BlockPaletteEncoding.DecodeBlockState(m)
			if err != nil {
				return nil, fmt.Errorf("NewChunkTimeline: %v", err)
			}

			result.blockPalette.AddBlock(blockRuntimeID)
		}

		globalData = globalData[4+length:]
	}

	// Barrier and Max limit
	{
		if len(globalData) < 12 {
			return nil, fmt.Errorf("NewChunkTimeline: Barrier and limit is broken (only get %d bytes but expected 12)", len(globalData))
		}
		result.barrierLeft = uint(binary.LittleEndian.Uint32(globalData))
		result.ptr = result.barrierLeft
		result.barrierRight = uint(binary.LittleEndian.Uint32(globalData[4:]))
		result.maxLimit = uint(binary.LittleEndian.Uint32(globalData[8:]))
	}

	// Latest Chunk
	{
		latestChunkBytes := t.Get(
			define.Sum(pos, define.KeyLatestChunk),
		)

		chunkMatrix, err := marshal.BytesToChunkMatrix(latestChunkBytes, pos.Dimension.Range())
		if err != nil {
			return nil, fmt.Errorf("NewChunkTimeline: %v", err)
		}

		result.latestChunk = chunkMatrix
	}

	// Latest NBT
	{
		latestNBTBytes := t.Get(
			define.Sum(pos, []byte(define.KeyLatestNBT)...),
		)

		latestNBT, err := marshal.BytesToBlockNBT(latestNBTBytes)
		if err != nil {
			return nil, fmt.Errorf("NewChunkTimeline: %v", err)
		}

		result.latestNBT = latestNBT
	}

	success = true
	return result, nil
}

// DeleteChunkTimeline delete the timeline of chunk who at pos.
// If timeline is not exist, then do no operation.
//
// Time complexity: O(n).
// n is the time point that this chunk have.
func (t *TimelineDB) DeleteChunkTimeline(pos define.DimChunk) error {
	var success bool

	timeline, err := t.NewChunkTimeline(pos, false)
	if err != nil {
		return fmt.Errorf("DeleteChunkTimeline: %v", err)
	}
	defer func() {
		timeline.releaseFunc()
	}()

	if timeline.isEmpty {
		return nil
	}

	tran, err := t.OpenTransaction()
	if err != nil {
		return fmt.Errorf("DeleteChunkTimeline: %v", err)
	}
	defer func() {
		if !success {
			_ = tran.Discard()
			return
		}
		_ = tran.Commit()
	}()

	// Global data
	err = tran.Delete(define.Sum(pos, []byte(define.KeyChunkGlobalData)...))
	if err != nil {
		return fmt.Errorf("DeleteChunkTimeline: %v", err)
	}

	// Chunk Index
	keyBytes := define.Index(pos)
	bucket := tran.(*transaction).tx.Bucket(DatabaseKeyChunkIndex)
	if bucket.Get(keyBytes) != nil {
		err = bucket.Put(
			DatabaseKeyChunkCount,
			utils.Uint32BinaryAdd(bucket.Get(DatabaseKeyChunkCount), []byte{1, 0, 0, 0}, -1),
		)
		if err != nil {
			return fmt.Errorf("DeleteChunkTimeline: %v", err)
		}
		err = bucket.Delete(keyBytes)
		if err != nil {
			return fmt.Errorf("DeleteChunkTimeline: %v", err)
		}
	}

	if err != nil {
		return fmt.Errorf("DeleteChunkTimeline: %v", err)
	}

	// Latest Chunk
	err = tran.Delete(define.Sum(pos, define.KeyLatestChunk))
	if err != nil {
		return fmt.Errorf("DeleteChunkTimeline: %v", err)
	}

	// Latest NBT
	err = tran.Delete(define.Sum(pos, []byte(define.KeyLatestNBT)...))
	if err != nil {
		return fmt.Errorf("DeleteChunkTimeline: %v", err)
	}

	// Each delta update
	for i := timeline.barrierLeft; i <= timeline.barrierRight; i++ {
		err = tran.Delete(define.IndexBlockDu(pos, i))
		if err != nil {
			return fmt.Errorf("DeleteChunkTimeline: %v", err)
		}
		err = tran.Delete(define.IndexNBTDu(pos, i))
		if err != nil {
			return fmt.Errorf("DeleteChunkTimeline: %v", err)
		}
	}

	success = true
	return nil
}
