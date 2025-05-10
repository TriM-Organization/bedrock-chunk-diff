package timeline

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/marshal"
	"github.com/TriM-Organization/bedrock-world-operator/block"
	block_general "github.com/TriM-Organization/bedrock-world-operator/block/general"
	"github.com/TriM-Organization/bedrock-world-operator/chunk"
	operator_define "github.com/TriM-Organization/bedrock-world-operator/define"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

const DefaultMaxLimit = 7

// SubChunkTimeline records the timeline of a sub chunk,
// and it contains the change logs about this sub chunk
// on this timeline.
// In other words, the SubChunkTimeline holds the history
// of this sub chunk.
type SubChunkTimeline struct {
	db      LevelDB
	isEmpty bool

	dm            operator_define.Dimension
	position      operator_define.ChunkPos
	subChunkIndex uint8

	timelineUnixTime    []int64
	blockPalette        []block_general.IndexBlockState
	blockPaletteMapping map[uint32]uint

	ptr          uint
	barrierLeft  uint
	barrierRight uint
	maxLimit     uint

	latestSubChunk define.Layers
	latestNBT      []define.NBTWithIndex
}

// NewSubChunkTimeline gets the timeline of a sub chunk who is in dm, position and subChunkIndex.
//
// Note that if timeline of current sub chunk is not exist, then we will not create a timeline
// but return an empty one so you can modify it. The time to create the timeline is only when you
// save a timeline that not empty to the database.
//
// Important: Once any modifications have been made to the returned timeline, you must save them
// at the end; otherwise, the timeline will not be able to maintain data consistency (only need to
// save at the last modification).
//
// subChunkIndex is an integer that bigger then -1.
// For example, if a block is at (x,23,z) and is in Overworld, then it is in a sub chunk
// whose Y position is 23>>4 = 1. However, this is not the index of this sub chunk,
// we need use (23>>4) - (dm.Range()[0]>>4) to get the index, which is 1-4=-3.
func (t *TimelineDB) NewSubChunkTimeline(
	dm operator_define.Dimension,
	position operator_define.ChunkPos,
	subChunkIndex uint8,
) (result *SubChunkTimeline, err error) {
	result = &SubChunkTimeline{
		db:                  t.db,
		dm:                  dm,
		position:            position,
		subChunkIndex:       subChunkIndex,
		blockPaletteMapping: make(map[uint32]uint),
		maxLimit:            DefaultMaxLimit,
	}

	payload, err := t.db.Get(
		define.Sum(dm, position, subChunkIndex, define.KeySubChunkExistStates),
	)
	if err != nil {
		return nil, fmt.Errorf("NewSubChunkTimeline: %v", err)
	}

	if len(payload) == 0 {
		result.isEmpty = true
		return result, nil
	}

	// Timeline Unix Time
	{
		payload, err := t.db.Get(
			define.Sum(dm, position, subChunkIndex, define.KeyTimelineUnixTime),
		)
		if err != nil {
			return nil, fmt.Errorf("NewSubChunkTimeline: %v", err)
		}
		for len(payload) > 0 {
			result.timelineUnixTime = append(result.timelineUnixTime, int64(binary.LittleEndian.Uint64(payload)))
			payload = payload[8:]
		}
	}

	// Block Palette
	{
		blockPaletteBytes, err := t.db.Get(
			define.Sum(dm, position, subChunkIndex, []byte(define.KeyBlockPalette)...),
		)
		if err != nil {
			return nil, fmt.Errorf("NewSubChunkTimeline: %v", err)
		}
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
			if _, ok := result.blockPaletteMapping[blockRuntimeID]; ok {
				continue
			}

			block, _ := block.RuntimeIDToIndexState(blockRuntimeID)
			result.blockPalette = append(result.blockPalette, block)
			result.blockPaletteMapping[blockRuntimeID] = uint(len(result.blockPaletteMapping) + 1)
		}
	}

	// Barrier and Max limit
	{
		payload, err := t.db.Get(
			define.Sum(dm, position, subChunkIndex, []byte(define.KeyBarrierAndLimit)...),
		)
		if err != nil {
			return nil, fmt.Errorf("NewSubChunkTimeline: %v", err)
		}
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
		latestSubChunkBytes, err := t.db.Get(
			define.Sum(dm, position, subChunkIndex, define.KeyLatestSubChunk),
		)
		if err != nil {
			return nil, fmt.Errorf("NewSubChunkTimeline: %v", err)
		}

		blockMatrix, err := marshal.BytesToLayers(latestSubChunkBytes)
		if err != nil {
			return nil, fmt.Errorf("NewSubChunkTimeline: %v", err)
		}

		result.latestSubChunk = blockMatrix
	}

	// Latest NBT
	{
		latestNBTBytes, err := t.db.Get(
			define.Sum(dm, position, subChunkIndex, []byte(define.KeyLatestNBT)...),
		)
		if err != nil {
			return nil, fmt.Errorf("NewSubChunkTimeline: %v", err)
		}

		latestNBT, err := marshal.BytesToBlockNBT(latestNBTBytes)
		if err != nil {
			return nil, fmt.Errorf("NewSubChunkTimeline: %v", err)
		}

		result.latestNBT = latestNBT
	}

	return result, nil
}
