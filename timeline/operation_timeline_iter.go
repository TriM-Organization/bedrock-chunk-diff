package timeline

import (
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/marshal"
	"github.com/TriM-Organization/bedrock-world-operator/block"
	"github.com/TriM-Organization/bedrock-world-operator/chunk"
)

// "next" is an internal implement detail.
// The feature of "next" is like one progress of prefix sum.
func (s *ChunkTimeline) next() (
	oriChunk define.ChunkMatrix, oriNBTs []define.NBTWithIndex, updateUnixTime int64,
	isLastElement bool, err error,
) {
	if s.isEmpty {
		return nil, nil, 0, false, fmt.Errorf("next: Current chunk timeline is empty")
	}
	isLastElement = (s.ptr == s.barrierRight)

	// Blocks
	{
		payload := s.db.Get(
			define.IndexBlockDu(s.pos, s.ptr),
		)

		diff, err := marshal.BytesToChunkDiffMatrix(payload, s.pos.Dimension.Range())
		if err != nil {
			return nil, nil, 0, false, fmt.Errorf("next: %v", err)
		}

		oriChunk = define.ChunkRestore(s.currentChunk, diff)
	}

	// NBTs
	{
		payload := s.db.Get(
			define.IndexNBTDu(s.pos, s.ptr),
		)

		diff, err := marshal.BytesToMultipleDiffNBT(payload)
		if err != nil {
			return nil, nil, 0, false, fmt.Errorf("next: %v", err)
		}

		oriNBTs, err = define.NBTRestore(s.currentNBT, diff)
		if err != nil {
			return nil, nil, 0, false, fmt.Errorf("next: %v", err)
		}
	}

	// Timeline Unix Time
	updateUnixTime = s.timelineUnixTime[s.ptr-s.barrierLeft]

	s.currentChunk = oriChunk
	s.currentNBT = oriNBTs
	s.ptr++

	if s.ptr > s.barrierRight {
		s.ResetPointer()
	}

	return oriChunk, oriNBTs, updateUnixTime, isLastElement, nil
}

// "convert" is an internal implement detail.
func (s *ChunkTimeline) convert(
	oriChunk define.ChunkMatrix,
	oriNBTs []define.NBTWithIndex,
) (c *chunk.Chunk, nbts []map[string]any) {
	// Blocks
	c = chunk.NewChunk(block.AirRuntimeID, s.pos.Dimension.Range())
	sub := c.Sub()
	for ChunkIndex, layers := range oriChunk {
		subChunk := sub[ChunkIndex]
		for index, value := range layers {
			layer := subChunk.Layer(uint8(index))

			if define.BlockMatrixIsEmpty(value) {
				continue
			}

			ptr := 0
			for x := range uint8(16) {
				for y := range uint8(16) {
					for z := range uint8(16) {
						layer.Set(x, y, z, s.blockPalette.BlockRuntimeID(value[ptr]))
						ptr++
					}
				}
			}
		}
	}

	// NBTs
	nbts = make([]map[string]any, 0)
	for _, value := range oriNBTs {
		nbts = append(nbts, value.NBT)
	}

	return
}

// Next gets the next time point of current chunk and the NBT blocks in it.
//
// With the call to Next, we granted that the returned time keeps increasing until
// the entire time series is traversed.
//
// isLastElement can inform whether the element obtained after the current call to
// Next is at the end of the time series.
//
// When it is already at the end of the timeline, calling Next again will back to
// the earliest time point. In other words, Next is self-loop and can be called continuously.
//
// Note that if Next returned non-nil error, then the underlying pointer will back to the
// firest time point due to when an error occurs, some of the underlying data maybe is inconsistent.
//
// Time complexity: O(4096×n + C).
// n is the sub chunk count of this chunk.
// C is relevant to the average changes between last time point and the next one.
func (s *ChunkTimeline) Next() (
	c *chunk.Chunk, nbts []map[string]any, updateUnixTime int64,
	isLastElement bool, err error,
) {
	var oriChunk define.ChunkMatrix
	var oriNBTs []define.NBTWithIndex

	if s.isEmpty {
		return nil, nil, 0, false, fmt.Errorf("(s *ChunkTimeline) Next: Current chunk timeline is empty")
	}

	oriChunk, oriNBTs, updateUnixTime, isLastElement, err = s.next()
	if err != nil {
		s.ResetPointer()
		return nil, nil, 0, false, fmt.Errorf("(s *ChunkTimeline) Next: %v", err)
	}
	c, nbts = s.convert(oriChunk, oriNBTs)

	return
}

// JumpTo moves to a specific time point of this timeline who is in index.
//
// JumpTo is a very useful replacement of Next when you are trying to jump
// to a specific time point and no need to get the information of other time point.
//
// Note that if JumpTo returned non-nil error, then the underlying pointer will back
// to the firest time point due to when an error occurs, some of the underlying data
// maybe is inconsistent.
//
// Time complexity: O(4096×n + C×(d+1)).
//   - n is the sub chunk count of this chunk.
//   - d is the distance between index and current pointer.
//   - C is relevant to the average changes of all these time point.
func (s *ChunkTimeline) JumpTo(index uint) (c *chunk.Chunk, nbts []map[string]any, updateUnixTime int64, err error) {
	var oriChunk define.ChunkMatrix
	var oriNBTs []define.NBTWithIndex

	if s.isEmpty {
		return nil, nil, 0, fmt.Errorf("(s *ChunkTimeline) JumpTo: Current chunk timeline is empty")
	}

	idx := s.barrierLeft + index
	if idx > s.barrierRight {
		return nil, nil, 0, fmt.Errorf("(s *ChunkTimeline) JumpTo: index %d is out of index %d", index, s.barrierRight-s.barrierLeft)
	}

	for {
		couldBreak := (s.ptr == idx)

		oriChunk, oriNBTs, updateUnixTime, _, err = s.next()
		if err != nil {
			s.ResetPointer()
			return nil, nil, 0, fmt.Errorf("(s *ChunkTimeline) JumpTo: %v", err)
		}

		if couldBreak {
			break
		}
	}
	c, nbts = s.convert(oriChunk, oriNBTs)

	return
}

// Last gets the latest time point of current chunk and the NBT blocks in it.
// Time complexity: O(4096×n).
// n is the sub chunk count of this chunk.
func (s *ChunkTimeline) Last() (
	c *chunk.Chunk,
	nbts []map[string]any,
	updateUnixTime int64,
	err error,
) {
	if s.isEmpty {
		return nil, nil, 0, fmt.Errorf("(s *ChunkTimeline) Last: Current chunk timeline is empty")
	}

	oriNBTsCopyOne, err := define.NBTDeepCopy(s.latestNBT)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("Last: %v", err)
	}
	c, nbts = s.convert(s.latestChunk, oriNBTsCopyOne)

	return c, nbts, s.timelineUnixTime[len(s.timelineUnixTime)-1], nil
}
