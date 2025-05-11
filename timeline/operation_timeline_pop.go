package timeline

import (
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/marshal"
)

// Pop tries to delete the first time point from this timeline.
// If current timeline is empty of there is only one time point,
// then we will do no operation.
func (s *SubChunkTimeline) Pop() error {
	var success bool
	var blockDst define.Layers
	var nbtDst []define.NBTWithIndex

	if s.isEmpty || s.barrierLeft == s.barrierRight {
		return nil
	}

	transaction, err := s.db.OpenTransaction()
	if err != nil {
		return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
	}
	defer func() {
		if !success {
			_ = transaction.Discard()
			return
		}
		_ = transaction.Commit()
	}()

	// Blocks
	for range 1 {
		var ori define.Layers
		var newDiff define.LayersDiff

		// Step 1: Get element 1 from timeline
		{
			payload := transaction.Get(
				define.IndexBlockDu(s.pos, s.barrierLeft),
			)

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
			payload := transaction.Get(
				define.IndexBlockDu(s.pos, s.barrierLeft+1),
			)
			if len(payload) == 0 {
				err = transaction.Delete(define.IndexBlockDu(s.pos, s.barrierLeft))
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
				_ = blockDst.Layer(index)
				blockDst[index] = define.Restore(ori.Layer(index), value)
			}

			for index, value := range blockDst {
				_ = newDiff.Layer(index)
				newDiff[index] = define.Difference(define.BlockMatrix{}, value)
			}
		}

		// Setp 3: Pop
		{
			err := transaction.Delete(define.IndexBlockDu(s.pos, s.barrierLeft))
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}

			payload, err := marshal.LayersDiffToBytes(newDiff)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}
			if len(payload) == 0 {
				err = transaction.Delete(define.IndexBlockDu(s.pos, s.barrierLeft+1))
				if err != nil {
					return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
				}
			} else {
				err = transaction.Put(
					define.IndexBlockDu(s.pos, s.barrierLeft+1),
					payload,
				)
				if err != nil {
					return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
				}
			}
		}
	}

	// NBTs
	for range 1 {
		var ori []define.NBTWithIndex
		var newDiff *define.MultipleDiffNBT

		// Setp 1: Get element 1 from timeline
		{
			payload := transaction.Get(
				define.IndexNBTDu(s.pos, s.barrierLeft),
			)

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
			payload := transaction.Get(
				define.IndexNBTDu(s.pos, s.barrierLeft+1),
			)
			if len(payload) == 0 {
				err = transaction.Delete(define.IndexNBTDu(s.pos, s.barrierLeft))
				if err != nil {
					return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
				}
				break
			}

			diff, err := marshal.BytesToMultipleDiffNBT(payload)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}

			nbtDst, err = define.NBTRestore(ori, diff)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}

			newDiff, err = define.NBTDifference(nil, nbtDst)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}
		}

		// Setp 3: Pop
		{
			err := transaction.Delete(define.IndexNBTDu(s.pos, s.barrierLeft))
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}

			payload, err := marshal.MultipleDiffNBTBytes(*newDiff)
			if err != nil {
				return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
			}
			if len(payload) == 0 {
				err = transaction.Delete(define.IndexNBTDu(s.pos, s.barrierLeft+1))
				if err != nil {
					return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
				}
			} else {
				err = transaction.Put(
					define.IndexNBTDu(s.pos, s.barrierLeft+1),
					payload,
				)
				if err != nil {
					return fmt.Errorf("(s *SubChunkTimeline) Pop: %v", err)
				}
			}
		}
	}

	s.barrierLeft++
	s.timelineUnixTime = s.timelineUnixTime[1:]

	if s.ptr < s.barrierLeft {
		s.ptr = s.barrierLeft
		s.currentSubChunk = define.Layers{}
		s.currentNBT = nil
	}

	success = true

	return nil
}
