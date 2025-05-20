package timeline

import (
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
)

// Compact compacts the underlying block palette as much as possible, try to delete all
// unused blocks from it.
//
// If current timeline is empty or read only, then calling Compact will do no operation.
//
// Compact is very expensive due to its time complexity is O(C×k×4096×N×L).
//
//   - k is the count of sub chunks that this chunk have.
//   - N is the count of time point that this timeline have.
//   - L is the average count of layers for each sub chunks in this timeline.
//   - C is a little big (bigger than 2) due to there are multiple difference/prefix-sum
//     operations need to do.
func (s *ChunkTimeline) Compact() error {
	var err error
	var success bool

	if s.isEmpty || s.isReadOnly {
		return nil
	}

	for s.barrierRight-s.barrierLeft+1 > s.maxLimit {
		if err := s.Pop(); err != nil {
			return fmt.Errorf("(s *ChunkTimeline) Compact: %v", err)
		}
	}

	originPtr := s.ptr
	length := s.barrierRight - s.barrierLeft + 1
	allTimePoint := make([]define.ChunkMatrix, length)
	for index := range allTimePoint {
		allTimePoint[index] = make(define.ChunkMatrix, s.pos.Dimension.Height()>>4)
	}

	for {
		index := s.ptr - s.barrierLeft

		allTimePoint[index], _, _, _, err = s.next()
		if err != nil {
			s.ptr = originPtr
			return fmt.Errorf("(s *ChunkTimeline) Compact: %v", err)
		}

		if s.ptr == originPtr {
			break
		}
	}

	newBlockPalette := define.NewBlockPalette()
	newAllTimePoint := make([]define.ChunkMatrix, length)
	for index := range newAllTimePoint {
		newAllTimePoint[index] = make(define.ChunkMatrix, s.pos.Dimension.Height()>>4)
	}

	for _, timePoint := range allTimePoint {
		for _, Chunk := range timePoint {
			for _, layer := range Chunk {
				if define.BlockMatrixIsEmpty(layer) {
					continue
				}
				for _, index := range layer {
					newBlockPalette.AddBlock(s.blockPalette.BlockRuntimeID(index))
				}
			}
		}
	}

	for whichTimePoint, timePoint := range allTimePoint {
		for whichSubChunk, Chunk := range timePoint {
			for whichLayer, layer := range Chunk {
				l := define.Layers{}
				_ = l.Layer(whichLayer)
				l[whichLayer] = define.NewBlockMatrix()

				if !define.BlockMatrixIsEmpty(layer) {
					for index, blockPaletteIndex := range layer {
						l[whichLayer][index] = newBlockPalette.BlockPaletteIndex(
							s.blockPalette.BlockRuntimeID(blockPaletteIndex),
						)
					}
				}

				newAllTimePoint[whichTimePoint][whichSubChunk] = l
			}
		}
	}

	transaction, err := s.db.OpenTransaction()
	if err != nil {
		return fmt.Errorf("(s *ChunkTimeline) Compact: %v", err)
	}

	originBarrierLeft := s.barrierLeft
	originBarrierRight := s.barrierRight
	originLatestChunk := s.latestChunk
	originCurrentChunk := s.currentChunk
	originCurrentNBT := s.currentNBT

	defer func() {
		s.ptr = originPtr
		s.barrierLeft = originBarrierLeft
		s.barrierRight = originBarrierRight
		s.latestChunk = originLatestChunk
		s.currentChunk = originCurrentChunk
		s.currentNBT = originCurrentNBT
		if !success {
			_ = transaction.Discard()
			return
		}
		_ = transaction.Commit()
	}()

	s.barrierRight = s.barrierLeft - 1
	s.latestChunk = make(define.ChunkMatrix, s.pos.Dimension.Height()>>4)

	// Update each time point
	for _, value := range newAllTimePoint {
		err = s.appendBlocks(value, transaction)
		if err != nil {
			return fmt.Errorf("(s *ChunkTimeline) Compact: %v", err)
		}
		s.latestChunk = value
		s.barrierRight++
	}

	s.blockPalette = newBlockPalette
	success = true

	return nil
}
