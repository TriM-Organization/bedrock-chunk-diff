package timeline

import (
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/marshal"
	"github.com/TriM-Organization/bedrock-world-operator/chunk"
)

// Append tries append a new sub chunk with block NBT data to the timeline of current sub chunk.
// updateUnixTime is the arrive time of the sub chunk data.
//
// If the size of timeline will overflow max limit, then we will firstly pop some time point from
// the underlying timeline. Note the poped time points must be the most earliest one.
func (s *SubChunkTimeline) Append(subChunk *chunk.SubChunk, nbt []map[string]any, updateUnixTime int64) error {
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

			_ = newerLayers.Layer(index)
			diff[index] = define.Difference(s.latestSubChunk.Layer(index), newerBlockMartrix)
			newerLayers[index] = newerBlockMartrix
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
			marshal.LayersToBytes(newerLayers),
		)
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

			xBlock, zBlock := s.position[0]<<4, s.position[0]<<4
			yBlock := (int32(s.subChunkIndex) - int32(s.dm.Range()[0])) << 4

			deltaY := y - yBlock
			if deltaY < 0 || deltaY > 15 {
				continue
			}

			nbtWithIndex.Index.UpdateIndex(uint8(x-xBlock), uint8(deltaY), uint8(z-zBlock))
			nbtWithIndex.NBT = value
			newerNBTs = append(newerNBTs, nbtWithIndex)
		}

		diff, err := define.NBTDifference(s.latestNBT, newerNBTs)
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
			marshal.BlockNBTBytes(newerNBTs),
		)
		if err != nil {
			return fmt.Errorf("(s *SubChunkTimeline) Append: %v", err)
		}
	}

	s.latestSubChunk = newerLayers
	s.latestNBT = newerNBTs
	s.barrierRight++
	s.timelineUnixTime = append(s.timelineUnixTime, updateUnixTime)
	success = true
	s.isEmpty = false

	return nil
}
