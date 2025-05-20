package timeline

import (
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/marshal"
)

// Pop tries to delete the first time point from this timeline.
// If current timeline is empty, or it is read only, or there is
// only one time point, then we will do no operation.
func (s *ChunkTimeline) Pop() error {
	var success bool

	if s.isEmpty || s.isReadOnly || s.barrierLeft == s.barrierRight {
		return nil
	}

	transaction, err := s.db.OpenTransaction()
	if err != nil {
		return fmt.Errorf("(s *ChunkTimeline) Pop: %v", err)
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
		var dst define.ChunkMatrix
		var newDiff define.ChunkDiffMatrix

		// Step 1: Get element 1 from timeline
		{
			payload := transaction.Get(
				define.IndexBlockDu(s.pos, s.barrierLeft),
			)

			diff, err := marshal.BytesToChunkDiffMatrix(payload, s.pos.Dimension.Range())
			if err != nil {
				return fmt.Errorf("(s *ChunkTimeline) Pop: %v", err)
			}

			dst = define.ChunkRestore(make(define.ChunkMatrix, len(diff)), diff)
		}

		// Setp 2: Get element 2 from timeline
		{
			payload := transaction.Get(
				define.IndexBlockDu(s.pos, s.barrierLeft+1),
			)
			if len(payload) == 0 {
				err = transaction.Delete(define.IndexBlockDu(s.pos, s.barrierLeft))
				if err != nil {
					return fmt.Errorf("(s *ChunkTimeline) Pop: %v", err)
				}
				break
			}

			diff, err := marshal.BytesToChunkDiffMatrix(payload, s.pos.Dimension.Range())
			if err != nil {
				return fmt.Errorf("(s *ChunkTimeline) Pop: %v", err)
			}

			_ = define.ChunkRestore(dst, diff)
			newDiff = define.ChunkDifference(make(define.ChunkMatrix, len(dst)), dst)
		}

		// Setp 3: Pop
		{
			err := transaction.Delete(define.IndexBlockDu(s.pos, s.barrierLeft))
			if err != nil {
				return fmt.Errorf("(s *ChunkTimeline) Pop: %v", err)
			}

			payload, err := marshal.ChunkDiffMatrixToBytes(newDiff)
			if err != nil {
				return fmt.Errorf("(s *ChunkTimeline) Pop: %v", err)
			}
			err = transaction.Put(
				define.IndexBlockDu(s.pos, s.barrierLeft+1),
				payload,
			)
			if err != nil {
				return fmt.Errorf("(s *ChunkTimeline) Pop: %v", err)
			}
		}
	}

	// NBTs
	for range 1 {
		var dst []define.NBTWithIndex
		var newDiff *define.MultipleDiffNBT

		// Setp 1: Get element 1 from timeline
		{
			payload := transaction.Get(
				define.IndexNBTDu(s.pos, s.barrierLeft),
			)

			diff, err := marshal.BytesToMultipleDiffNBT(payload)
			if err != nil {
				return fmt.Errorf("(s *ChunkTimeline) Pop: %v", err)
			}

			dst, err = define.NBTRestore(nil, diff)
			if err != nil {
				return fmt.Errorf("(s *ChunkTimeline) Pop: %v", err)
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
					return fmt.Errorf("(s *ChunkTimeline) Pop: %v", err)
				}
				break
			}

			diff, err := marshal.BytesToMultipleDiffNBT(payload)
			if err != nil {
				return fmt.Errorf("(s *ChunkTimeline) Pop: %v", err)
			}

			dst, err = define.NBTRestore(dst, diff)
			if err != nil {
				return fmt.Errorf("(s *ChunkTimeline) Pop: %v", err)
			}

			newDiff, err = define.NBTDifference(nil, dst)
			if err != nil {
				return fmt.Errorf("(s *ChunkTimeline) Pop: %v", err)
			}
		}

		// Setp 3: Pop
		{
			err := transaction.Delete(define.IndexNBTDu(s.pos, s.barrierLeft))
			if err != nil {
				return fmt.Errorf("(s *ChunkTimeline) Pop: %v", err)
			}

			payload, err := marshal.MultipleDiffNBTBytes(*newDiff)
			if err != nil {
				return fmt.Errorf("(s *ChunkTimeline) Pop: %v", err)
			}
			err = transaction.Put(
				define.IndexNBTDu(s.pos, s.barrierLeft+1),
				payload,
			)
			if err != nil {
				return fmt.Errorf("(s *ChunkTimeline) Pop: %v", err)
			}
		}
	}

	s.barrierLeft++
	s.timelineUnixTime = s.timelineUnixTime[1:]

	if s.ptr < s.barrierLeft {
		s.ptr = s.barrierLeft
		s.currentChunk = make(define.ChunkMatrix, s.pos.Dimension.Height()>>4)
		s.currentNBT = nil
	}

	success = true

	return nil
}
