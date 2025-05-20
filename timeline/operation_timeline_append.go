package timeline

import (
	"fmt"
	"time"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/marshal"
	"github.com/TriM-Organization/bedrock-world-operator/chunk"
)

// "appendBlocks" is an internal implement detail.
func (s *ChunkTimeline) appendBlocks(newerChunk define.ChunkMatrix, transaction Transaction) error {
	for s.barrierRight-s.barrierLeft+1 >= s.maxLimit {
		if err := s.Pop(); err != nil {
			return fmt.Errorf("appendBlocks: %v", err)
		}
	}

	// Put delta update
	diff := define.ChunkDifference(s.latestChunk, newerChunk)
	payload, err := marshal.ChunkDiffMatrixToBytes(diff)
	if err != nil {
		return fmt.Errorf("appendBlocks: %v", err)
	}
	err = transaction.Put(
		define.IndexBlockDu(s.pos, s.barrierRight+1),
		payload,
	)
	if err != nil {
		return fmt.Errorf("appendBlocks: %v", err)
	}

	// Update Latest Chunk
	payload, err = marshal.ChunkMatrixToBytes(newerChunk)
	if err != nil {
		return fmt.Errorf("appendBlocks: %v", err)
	}
	err = transaction.Put(
		define.Sum(s.pos, define.KeyLatestChunk),
		payload,
	)
	if err != nil {
		return fmt.Errorf("appendBlocks: %v", err)
	}

	return nil
}

// "appendNBTs" is an internal implement detail.
func (s *ChunkTimeline) appendNBTs(newerNBTs []define.NBTWithIndex, transaction Transaction) error {
	for s.barrierRight-s.barrierLeft+1 >= s.maxLimit {
		if err := s.Pop(); err != nil {
			return fmt.Errorf("appendNBTs: %v", err)
		}
	}

	// Compute diff
	diff, err := define.NBTDifference(s.latestNBT, newerNBTs)
	if err != nil {
		return fmt.Errorf("appendNBTs: %v", err)
	}

	// Put delta update
	payload, err := marshal.MultipleDiffNBTBytes(*diff)
	if err != nil {
		return fmt.Errorf("appendNBTs: %v", err)
	}
	err = transaction.Put(
		define.IndexNBTDu(s.pos, s.barrierRight+1),
		payload,
	)
	if err != nil {
		return fmt.Errorf("appendNBTs: %v", err)
	}

	// Update Latest NBT
	payload, err = marshal.BlockNBTBytes(newerNBTs)
	if err != nil {
		return fmt.Errorf("appendNBTs: %v", err)
	}
	err = transaction.Put(
		define.Sum(s.pos, []byte(define.KeyLatestNBT)...),
		payload,
	)
	if err != nil {
		return fmt.Errorf("appendNBTs: %v", err)
	}

	return nil
}

// Append tries append a new chunk with block
// NBT data to the timeline of current chunk.
//
// If the size of timeline will overflow max
// limit, then we will firstly pop some time
// point from the underlying timeline.
//
// Note the poped time points must be the most
// earliest one.
//
// If current timeline is read only, then calling
// Append will do no operation.
func (s *ChunkTimeline) Append(Chunk *chunk.Chunk, nbt []map[string]any) error {
	var success bool
	var newerChunk define.ChunkMatrix
	var newerNBTs []define.NBTWithIndex

	if s.isReadOnly {
		return nil
	}

	for s.barrierRight-s.barrierLeft+1 >= s.maxLimit {
		if err := s.Pop(); err != nil {
			return fmt.Errorf("(s *ChunkTimeline) Append: %v", err)
		}
	}

	transaction, err := s.db.OpenTransaction()
	if err != nil {
		return fmt.Errorf("(s *ChunkTimeline) Append: %v", err)
	}
	defer func() {
		if !success {
			_ = transaction.Discard()
			return
		}
		_ = transaction.Commit()
	}()

	// Blocks
	for _, value := range Chunk.Sub() {
		l := define.Layers{}

		if value.Empty() {
			if len(value.Layers()) == 0 {
				newerChunk = append(newerChunk, l)
				continue
			}
			_ = l.Layer(0)
			newerChunk = append(newerChunk, l)
			continue
		}

		for index, layer := range value.Layers() {
			newerBlockMartrix := define.NewBlockMatrix()

			ptr := 0
			for x := range uint8(16) {
				for y := range uint8(16) {
					for z := range uint8(16) {
						newerBlockMartrix[ptr] = s.blockPalette.BlockPaletteIndex(layer.At(x, y, z))
						ptr++
					}
				}
			}

			_ = l.Layer(index)
			l[index] = newerBlockMartrix
		}

		newerChunk = append(newerChunk, l)
	}
	err = s.appendBlocks(newerChunk, transaction)
	if err != nil {
		return fmt.Errorf("(s *ChunkTimeline) Append: %v", err)
	}

	// NBTs
	{
		for _, value := range nbt {
			x, ok1 := value["x"].(int32)
			y, ok2 := value["y"].(int32)
			z, ok3 := value["z"].(int32)

			if !ok1 || !ok2 || !ok3 {
				return fmt.Errorf("(s *ChunkTimeline) Append: Broken NBT data %#v", value)
			}

			nbtWithIndex := define.NBTWithIndex{}

			xBlock, zBlock := s.pos.ChunkPos[0]<<4, s.pos.ChunkPos[1]<<4

			deltaX := x - xBlock
			deltaZ := z - zBlock
			if deltaX < 0 || deltaX > 15 || deltaZ < 0 || deltaZ > 15 {
				continue
			}

			nbtWithIndex.Index.UpdateIndex(uint8(x-xBlock), int16(y), uint8(z-zBlock))
			nbtWithIndex.NBT = value
			newerNBTs = append(newerNBTs, nbtWithIndex)
		}

		err = s.appendNBTs(newerNBTs, transaction)
		if err != nil {
			return fmt.Errorf("(s *ChunkTimeline) Append: %v", err)
		}
	}

	s.latestChunk = newerChunk
	s.latestNBT = newerNBTs
	s.barrierRight++
	s.timelineUnixTime = append(s.timelineUnixTime, time.Now().Unix())
	success = true

	if s.isEmpty {
		s.barrierLeft = s.barrierRight
		s.ptr = s.barrierLeft
		s.isEmpty = false
	}

	return nil
}
