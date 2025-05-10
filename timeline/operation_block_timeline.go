package timeline

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/marshal"
	"github.com/TriM-Organization/bedrock-chunk-diff/utils"
	"github.com/TriM-Organization/bedrock-world-operator/block"
	block_general "github.com/TriM-Organization/bedrock-world-operator/block/general"
	"github.com/TriM-Organization/bedrock-world-operator/chunk"
	operator_define "github.com/TriM-Organization/bedrock-world-operator/define"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

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

// Empty returns whether this timeline is empty or not.
// If is empty, then calling Save will result in no operation.
func (s *SubChunkTimeline) Empty() bool {
	return s.isEmpty
}

// SetMaxLimit sets the timeline could record how many time point.
// maxLimit must bigger than 0. If less, then set the limit to 1.
//
// Note that calling SetMaxLimit will not change the empty states
// of this timeline.
func (s *SubChunkTimeline) SetMaxLimit(maxLimit uint) {
	s.maxLimit = max(maxLimit, 1)
}

// BlockPaletteIndex finds the index of blockRuntimeID in block palette.
// If not exist, then added it the underlying block palette.
//
// Returned index is the real index plus 1.
// If you got 0, then that means this is an air block.
// We don't save air block in block palette, and you should to pay attention to it.
func (s *SubChunkTimeline) BlockPaletteIndex(blockRuntimeID uint32) uint {
	if blockRuntimeID == block.AirRuntimeID {
		return 0
	}

	idx, ok := s.blockPaletteMapping[blockRuntimeID]
	if ok {
		return idx
	}

	name, states, found := block.RuntimeIDToState(blockRuntimeID)
	if !found {
		name = "minecraft:unknown"
		states = make(map[string]any)
	}

	blockRuntimeID, _ = block.StateToRuntimeID(name, states)
	indexState, _ := block.RuntimeIDToIndexState(blockRuntimeID)

	s.blockPalette = append(s.blockPalette, indexState)
	idx = uint(len(s.blockPaletteMapping) + 1)
	s.blockPaletteMapping[blockRuntimeID] = idx

	return idx
}

// Pop tries to delete the first time point from this timeline.
// If current timeline is empty of there is only one time point,
// then we will do no operation.
func (s *SubChunkTimeline) Pop() error {
	var success bool

	if s.isEmpty || s.barrierLeft == s.barrierRight {
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

	// Blocks
	for range 1 {
		var ori define.Layers
		var dst define.Layers
		var newDiff define.LayersDiff

		// Step 1: Get element 1 from timeline
		{
			payload, err := transaction.Get(
				define.IndexBlockDu(s.dm, s.position, s.subChunkIndex, s.barrierLeft),
			)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}

			diff, err := marshal.BytesToLayersDiff(payload)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}

			for index, value := range diff {
				_ = ori.Layer(index)
				ori[index] = define.Restore(define.BlockMatrix{}, value)
			}
		}

		// Setp 2: Get element 2 from timeline
		{
			payload, err := transaction.Get(
				define.IndexBlockDu(s.dm, s.position, s.subChunkIndex, s.barrierLeft+1),
			)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}

			if len(payload) == 0 {
				err = transaction.Delete(define.IndexBlockDu(s.dm, s.position, s.subChunkIndex, s.barrierLeft))
				if err != nil {
					return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
				}
				break
			}

			diff, err := marshal.BytesToLayersDiff(payload)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}

			for index, value := range diff {
				_ = dst.Layer(index)
				dst[index] = define.Restore(ori.Layer(index), value)
			}

			for index, value := range dst {
				_ = newDiff.Layer(index)
				newDiff[index] = define.Difference(define.BlockMatrix{}, value)
			}
		}

		// Setp 3: Pop
		{
			err := transaction.Delete(define.IndexBlockDu(s.dm, s.position, s.subChunkIndex, s.barrierLeft))
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}

			payload := marshal.LayersDiffToBytes(newDiff)
			err = transaction.Put(
				define.IndexBlockDu(s.dm, s.position, s.subChunkIndex, s.barrierLeft+1),
				payload,
			)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}
		}

		// Setp 4: Sync data
		s.latestSubChunk = dst
	}

	// NBTs
	for range 1 {
		var ori []define.NBTWithIndex
		var dst []define.NBTWithIndex
		var newDiff *define.MultipleDiffNBT

		// Setp 1: Get element 1 from timeline
		{
			payload, err := transaction.Get(
				define.IndexNBTDu(s.dm, s.position, s.subChunkIndex, s.barrierLeft),
			)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}

			diff, err := marshal.BytesToMultipleDiffNBT(payload)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}

			ori, err = define.NBTRestore(nil, diff)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}
		}

		// Setp 2: Get element 2 from timeline
		{
			payload, err := transaction.Get(
				define.IndexNBTDu(s.dm, s.position, s.subChunkIndex, s.barrierLeft+1),
			)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}

			if len(payload) == 0 {
				err = transaction.Delete(define.IndexNBTDu(s.dm, s.position, s.subChunkIndex, s.barrierLeft))
				if err != nil {
					return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
				}
				break
			}

			diff, err := marshal.BytesToMultipleDiffNBT(payload)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}

			dst, err = define.NBTRestore(ori, diff)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}

			newDiff, err = define.NBTDifference(nil, dst)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}
		}

		// Setp 3: Pop
		{
			err := transaction.Delete(define.IndexNBTDu(s.dm, s.position, s.subChunkIndex, s.barrierLeft))
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}

			payload := marshal.MultipleDiffNBTBytes(*newDiff)
			err = transaction.Put(
				define.IndexNBTDu(s.dm, s.position, s.subChunkIndex, s.barrierLeft+1),
				payload,
			)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}
		}

		// Setp 4: Sync data
		s.latestNBT = dst
	}

	s.barrierLeft++
	s.ptr = max(s.ptr, s.barrierLeft)
	s.timelineUnixTime = s.timelineUnixTime[1:]
	success = true

	return nil
}

// Append tries append a new sub chunk with block NBT data to the timeline of current sub chunk.
// updateUnixTime is the arrive time of the sub chunk data.
//
// If the size of timeline will overflow max limit, then we will firstly pop some time point from
// the underlying timeline. Note the poped time points must be the most earliest one.
func (s *SubChunkTimeline) Append(subChunk *chunk.SubChunk, nbt []map[string]any, updateUnixTime int64) error {
	var success bool

	for s.barrierRight-s.barrierLeft+1 >= s.maxLimit {
		if err := s.Pop(); err != nil {
			return fmt.Errorf("(s *SubChunkTimeline) Append: %v", err)
		}
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

	// Blocks
	{
		var newer define.Layers

		diff := define.LayersDiff{}
		for i := range s.latestSubChunk {
			diff.Layer(i)
		}
		for i := range subChunk.Layers() {
			diff.Layer(i)
		}

		for index := range len(diff) {
			newerBlockMartrix := define.BlockMatrix{}
			layer := subChunk.Layer(uint8(index))

			ptr := 0
			for x := range uint8(16) {
				for y := range uint8(16) {
					for z := range uint8(16) {
						newerBlockMartrix[ptr] = uint16(s.BlockPaletteIndex(layer.At(x, y, z)))
						ptr++
					}
				}
			}

			_ = newer.Layer(index)
			diff[index] = define.Difference(s.latestSubChunk.Layer(index), newerBlockMartrix)
			newer[index] = newerBlockMartrix
		}

		err := transaction.Put(
			define.IndexBlockDu(s.dm, s.position, s.subChunkIndex, s.barrierRight+1),
			marshal.LayersDiffToBytes(diff),
		)
		if err != nil {
			return fmt.Errorf("(s *SubChunkTimeline) Append: %v", err)
		}

		err = transaction.Put(
			define.Sum(s.dm, s.position, s.subChunkIndex, define.KeyLatestSubChunk),
			marshal.LayersToBytes(newer),
		)
		if err != nil {
			return fmt.Errorf("(s *SubChunkTimeline) Append: %v", err)
		}
	}

	// NBTs
	{
		newer := make([]define.NBTWithIndex, 0)

		for _, value := range nbt {
			x, ok1 := value["x"].(int32)
			y, ok2 := value["y"].(int32)
			z, ok3 := value["z"].(int32)

			if !ok1 || !ok2 || !ok3 {
				return fmt.Errorf("(s *SubChunkTimeline) Append: Broken NBT data %#v", value)
			}

			nbtWithIndex := define.NBTWithIndex{}

			xBlock, zBlock := s.position[0]<<4, s.position[0]<<4
			yBlock := (int32(s.subChunkIndex) - int32(s.dm.Range()[0])) << 4

			deltaY := y - yBlock
			if deltaY < 0 || deltaY > 15 {
				continue
			}

			nbtWithIndex.Index.UpdateIndex(uint8(x-xBlock), uint8(deltaY), uint8(z-zBlock))
			nbtWithIndex.NBT = value
			newer = append(newer, nbtWithIndex)
		}

		diff, err := define.NBTDifference(s.latestNBT, newer)
		if err != nil {
			return fmt.Errorf("(s *SubChunkTimeline) Append: %v", err)
		}

		err = transaction.Put(
			define.IndexNBTDu(s.dm, s.position, s.subChunkIndex, s.barrierRight+1),
			marshal.MultipleDiffNBTBytes(*diff),
		)
		if err != nil {
			return fmt.Errorf("(s *SubChunkTimeline) Append: %v", err)
		}

		err = transaction.Put(
			define.Sum(s.dm, s.position, s.subChunkIndex, []byte(define.KeyLatestNBT)...),
			marshal.BlockNBTBytes(newer),
		)
		if err != nil {
			return fmt.Errorf("(s *SubChunkTimeline) Append: %v", err)
		}
	}

	s.barrierRight++
	s.timelineUnixTime = append(s.timelineUnixTime, updateUnixTime)
	success = true
	s.isEmpty = false

	return nil
}
