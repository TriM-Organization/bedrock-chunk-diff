package timeline

import (
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
)

// Compact compacts the underlying block palette as much as possible, try to delete all unused blocks from it.
// Compact is very expensive due to its time complexity is O(C×4096×N×L).
// N is the count of time point that this timeline have, and L is the average count of layers in all timelines.
// C is a little big (bigger than 2) due to there are multiple difference/prefix-sum operations need to do.
func (s *SubChunkTimeline) Compact() error {
	var err error
	var success bool

	if s.isEmpty {
		return nil
	}

	originPtr := s.ptr
	length := s.barrierRight - s.barrierLeft + 1
	allTimePoint := make([]define.Layers, length)

	for {
		index := s.ptr - s.barrierLeft

		allTimePoint[index], _, _, _, err = s.next()
		if err != nil {
			s.ptr = originPtr
			return fmt.Errorf("(s *SubChunkTimeline) Compact: %v", err)
		}

		if s.ptr == originPtr {
			break
		}
	}

	newBlockPalette := define.NewBlockPalette()
	newAllTimePoint := make([]define.Layers, length)

	for _, timePoint := range allTimePoint {
		for _, layer := range timePoint {
			for _, index := range layer {
				newBlockPalette.AddBlock(s.blockPalette.BlockRuntimeID(index))
			}
		}
	}

	for whichTimePoint, timePoint := range allTimePoint {
		for whichLayer, layer := range timePoint {
			_ = newAllTimePoint[whichTimePoint].Layer(whichLayer)
			for index, blockPaletteIndex := range layer {
				newAllTimePoint[whichTimePoint][whichLayer][index] = newBlockPalette.BlockPaletteIndex(
					s.blockPalette.BlockRuntimeID(blockPaletteIndex),
				)
			}
		}
	}

	transaction, err := s.db.OpenTransaction()
	if err != nil {
		return fmt.Errorf("(s *SubChunkTimeline) Compact: %v", err)
	}

	originLatestSubChunk := s.latestSubChunk
	originBarrierRight := s.barrierRight
	defer func() {
		if success {
			_ = transaction.Commit()
			return
		}
		transaction.Discard()
		s.isEmpty = false
		s.barrierRight = originBarrierRight
		s.latestSubChunk = originLatestSubChunk
	}()

	s.isEmpty = true
	s.barrierRight = s.barrierLeft - 1

	// Update each time point
	for _, value := range newAllTimePoint {
		err = s.appendBlocks(value, transaction)
		if err != nil {
			return fmt.Errorf("(s *SubChunkTimeline) Compact: %v", err)
		}
		s.latestSubChunk = value
	}

	s.isEmpty = false
	s.blockPalette = newBlockPalette
	success = true

	return nil
}
