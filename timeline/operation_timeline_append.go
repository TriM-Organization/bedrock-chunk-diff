package timeline

import (
	"fmt"
	"time"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/marshal"
	"github.com/TriM-Organization/bedrock-world-operator/chunk"
)

// "appendBlocks" is an internal implement detail.
func (s *SubChunkTimeline) appendBlocks(newerLayers define.Layers, transaction Transaction) error {
	for s.barrierRight-s.barrierLeft+1 >= s.maxLimit {
		if err := s.Pop(); err != nil {
			return fmt.Errorf("appendBlocks: %v", err)
		}
	}

	diff := define.LayersDiff{}
	for i := range s.latestSubChunk {
		diff.Layer(i)
	}
	for i := range newerLayers {
		diff.Layer(i)
	}

	for index := range diff {
		diff[index] = define.Difference(
			s.latestSubChunk.Layer(index),
			newerLayers.Layer(index),
		)
	}

	payload, err := marshal.LayersDiffToBytes(diff)
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

	payload, err = marshal.LayersToBytes(newerLayers)
	if err != nil {
		return fmt.Errorf("appendBlocks: %v", err)
	}
	err = transaction.Put(
		define.Sum(s.pos, define.KeyLatestSubChunk),
		payload,
	)
	if err != nil {
		return fmt.Errorf("appendBlocks: %v", err)
	}

	return nil
}

// "appendNBTs" is an internal implement detail.
func (s *SubChunkTimeline) appendNBTs(newerNBTs []define.NBTWithIndex, transaction Transaction) error {
	for s.barrierRight-s.barrierLeft+1 >= s.maxLimit {
		if err := s.Pop(); err != nil {
			return fmt.Errorf("appendNBTs: %v", err)
		}
	}

	diff, err := define.NBTDifference(s.latestNBT, newerNBTs)
	if err != nil {
		return fmt.Errorf("appendNBTs: %v", err)
	}

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

// Append tries append a new sub chunk with block NBT data to the timeline of current sub chunk.
// If the size of timeline will overflow max limit, then we will firstly pop some time point from
// the underlying timeline. Note the poped time points must be the most earliest one.
func (s *SubChunkTimeline) Append(subChunk *chunk.SubChunk, nbt []map[string]any) error {
	var success bool
	var newerLayers define.Layers
	var newerNBTs []define.NBTWithIndex

	for s.barrierRight-s.barrierLeft+1 >= s.maxLimit {
		if err := s.Pop(); err != nil {
			return fmt.Errorf("(s *SubChunkTimeline) Append: %v", err)
		}
	}

	transaction, err := s.db.OpenTransaction()
	if err != nil {
		return fmt.Errorf("(s *SubChunkTimeline) Append: %v", err)
	}
	defer func() {
		if !success {
			_ = transaction.Discard()
			return
		}
		_ = transaction.Commit()
	}()

	// Blocks
	{
		for index, layer := range subChunk.Layers() {
			newerBlockMartrix := define.BlockMatrix{}

			ptr := 0
			for x := range uint8(16) {
				for y := range uint8(16) {
					for z := range uint8(16) {
						newerBlockMartrix[ptr] = s.blockPalette.BlockPaletteIndex(layer.At(x, y, z))
						ptr++
					}
				}
			}

			_ = newerLayers.Layer(index)
			newerLayers[index] = newerBlockMartrix
		}

		err = s.appendBlocks(newerLayers, transaction)
		if err != nil {
			return fmt.Errorf("(s *SubChunkTimeline) Append: %v", err)
		}
	}

	// NBTs
	{
		for _, value := range nbt {
			x, ok1 := value["x"].(int32)
			y, ok2 := value["y"].(int32)
			z, ok3 := value["z"].(int32)

			if !ok1 || !ok2 || !ok3 {
				return fmt.Errorf("(s *SubChunkTimeline) Append: Broken NBT data %#v", value)
			}

			nbtWithIndex := define.NBTWithIndex{}

			xBlock, zBlock := s.pos.ChunkPos[0]<<4, s.pos.ChunkPos[1]<<4
			yBlock := (int32(s.pos.SubChunkIndex) + (int32(s.pos.Dimension.Range()[0]) >> 4)) << 4

			deltaX := x - xBlock
			deltaY := y - yBlock
			deltaZ := z - zBlock
			if deltaX < 0 || deltaX > 15 || deltaY < 0 || deltaY > 15 || deltaZ < 0 || deltaZ > 15 {
				continue
			}

			nbtWithIndex.Index.UpdateIndex(uint8(x-xBlock), uint8(deltaY), uint8(z-zBlock))
			nbtWithIndex.NBT = value
			newerNBTs = append(newerNBTs, nbtWithIndex)
		}

		err = s.appendNBTs(newerNBTs, transaction)
		if err != nil {
			return fmt.Errorf("(s *SubChunkTimeline) Append: %v", err)
		}
	}

	s.latestSubChunk = newerLayers
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
