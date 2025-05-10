package time_line

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/marshal"
	"github.com/TriM-Organization/bedrock-world-operator/block"
	"github.com/TriM-Organization/bedrock-world-operator/chunk"
	operator_define "github.com/TriM-Organization/bedrock-world-operator/define"

	block_general "github.com/TriM-Organization/bedrock-world-operator/block/general"
	"github.com/df-mc/goleveldb/leveldb"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const DefaultMaxLimit = 7

// TimeLineDB implements chunk timeline and
// history record provider based on LevelDB.
type TimeLineDB struct {
	db LevelDB
}

// NewTimeLineDB open a level database that used for
// chunk delta update whose at path.
// If not exist, then create a new database.
func NewTimeLineDB(path string) (result *TimeLineDB, err error) {
	result = new(TimeLineDB)

	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, fmt.Errorf("NewTimeLineDB: %v", err)
	}

	result.db = &database{ldb: db}
	return result, nil
}

// SubChunkTimeLine records the timeline of a sub chunk,
// and it contains the change logs about this sub chunk
// on this timeline.
// In other words, the SubChunkTimeLine holds the history
// of this sub chunk.
type SubChunkTimeLine struct {
	db      LevelDB
	isEmpty bool

	subChunkPos  protocol.SubChunkPos
	blockPalette []block_general.IndexBlockState

	barrierLeft  uint
	barrierRight uint
	maxLimit     uint

	latestSubChunk *define.BlockMatrix
	latestNBT      []define.NBTWithIndex
}

// NewSubChunkTimeLine get the timeline of a sub chunk who is in dm, position and subChunkIndex.
//
// Note that if timeline of current sub chunk is not exist, then we will not create a timeline
// but return an empty one so you can modify it. The time to create the timeline is only when you
// save this timeline to the database.
//
// subChunkIndex is an integer that bigger then -1.
// For example, if a block is at (x,23,z) and is in Overworld, then it is in a sub chunk
// whose Y position is 23>>4 = 1. However, this is not the index of this sub chunk,
// we need use (23>>4) - (dm.Range()[0]>>4) to get the index, which is 1-4=-3.
func (t *TimeLineDB) NewSubChunkTimeLine(
	dm operator_define.Dimension,
	position operator_define.ChunkPos,
	subChunkIndex uint8,
) (result *SubChunkTimeLine, err error) {
	subChunkPos := protocol.SubChunkPos{
		position[0],
		int32(subChunkIndex) + int32(dm.Range()[0]>>4),
		position[1],
	}
	result = &SubChunkTimeLine{
		db:             t.db,
		isEmpty:        false,
		subChunkPos:    subChunkPos,
		blockPalette:   make([]block_general.IndexBlockState, 0),
		barrierLeft:    0,
		barrierRight:   0,
		maxLimit:       DefaultMaxLimit,
		latestSubChunk: nil,
		latestNBT:      make([]define.NBTWithIndex, 0),
	}

	payload, err := t.db.Get(
		define.Sum(dm, position, subChunkIndex, define.KeySubChunkExistStates),
	)
	if err != nil {
		return nil, fmt.Errorf("NewSubChunkTimeLine: %v", err)
	}

	if len(payload) == 0 {
		result.isEmpty = true
		return result, nil
	}

	// Block Palette
	{
		blockPaletteBytes, err := t.db.Get(
			define.Sum(dm, position, subChunkIndex, []byte(define.KeyBlockPalette)...),
		)
		if err != nil {
			return nil, fmt.Errorf("NewSubChunkTimeLine: %v", err)
		}
		buf := bytes.NewBuffer(blockPaletteBytes)

		for buf.Len() > 0 {
			var m map[string]any
			if err := nbt.NewDecoderWithEncoding(buf, nbt.LittleEndian).Decode(&m); err != nil {
				return nil, fmt.Errorf("NewSubChunkTimeLine: error decoding block palette entry: %w", err)
			}
			blockRuntimeID, err := chunk.BlockPaletteEncoding.DecodeBlockState(m)
			if err != nil {
				return nil, fmt.Errorf("NewSubChunkTimeLine: %v", err)
			}
			block, _ := block.RuntimeIDToIndexState(blockRuntimeID)
			result.blockPalette = append(result.blockPalette, block)
		}
	}

	// Barrier and Max limit
	{
		barrierLeftBytes, err := t.db.Get(
			define.Sum(dm, position, subChunkIndex, define.KeyBarrierLeft),
		)
		if err != nil {
			return nil, fmt.Errorf("NewSubChunkTimeLine: %v", err)
		}
		result.barrierLeft = uint(binary.LittleEndian.Uint32(barrierLeftBytes))

		barrierRightBytes, err := t.db.Get(
			define.Sum(dm, position, subChunkIndex, define.KeyBarrierRight),
		)
		if err != nil {
			return nil, fmt.Errorf("NewSubChunkTimeLine: %v", err)
		}
		result.barrierRight = uint(binary.LittleEndian.Uint32(barrierRightBytes))

		maxLimitBytes, err := t.db.Get(
			define.Sum(dm, position, subChunkIndex, define.KeyMaxLimit),
		)
		if err != nil {
			return nil, fmt.Errorf("NewSubChunkTimeLine: %v", err)
		}
		result.maxLimit = uint(binary.LittleEndian.Uint32(maxLimitBytes))
	}

	// Latest Sub Chunk
	{
		latestSubChunkBytes, err := t.db.Get(
			define.Sum(dm, position, subChunkIndex, define.KeyLatestSubChunk),
		)
		if err != nil {
			return nil, fmt.Errorf("NewSubChunkTimeLine: %v", err)
		}

		blockMatrix, err := marshal.BytesToBlockMatrix(latestSubChunkBytes)
		if err != nil {
			return nil, fmt.Errorf("NewSubChunkTimeLine: %v", err)
		}

		result.latestSubChunk = &blockMatrix
	}

	// Latest NBT
	{
		latestNBTBytes, err := t.db.Get(
			define.Sum(dm, position, subChunkIndex, []byte(define.KeyLatestNBT)...),
		)
		if err != nil {
			return nil, fmt.Errorf("NewSubChunkTimeLine: %v", err)
		}

		multipleNBTs, err := marshal.BytesToBlockNBT(latestNBTBytes)
		if err != nil {
			return nil, fmt.Errorf("NewSubChunkTimeLine: %v", err)
		}

		result.latestNBT = multipleNBTs
	}

	return result, nil
}
