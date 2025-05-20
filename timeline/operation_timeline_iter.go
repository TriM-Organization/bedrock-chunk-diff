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
		return nil, nil, 0, false, nil
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
		s.ptr = s.barrierLeft
		s.currentChunk = make(define.ChunkMatrix, s.pos.Dimension.Height()>>4)
		s.currentNBT = nil
	}

	return oriChunk, oriNBTs, updateUnixTime, isLastElement, nil
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
func (s *ChunkTimeline) Next() (
	c *chunk.Chunk, nbts []map[string]any, updateUnixTime int64,
	isLastElement bool, err error,
) {
	var oriChunk define.ChunkMatrix
	var oriNBTs []define.NBTWithIndex

	oriChunk, oriNBTs, updateUnixTime, isLastElement, err = s.next()
	if err != nil {
		return nil, nil, 0, false, fmt.Errorf("(s *ChunkTimeline) Next: %v", err)
	}

	// Blocks
	c = chunk.NewChunk(block.AirRuntimeID, s.pos.Dimension.Range())
	sub := c.Sub()
	for ChunkIndex, layers := range oriChunk {
		sub := sub[ChunkIndex]
		for index, value := range layers {
			layer := sub.Layer(uint8(index))

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

// Last gets the latest time point of current chunk and the NBT blocks in it.
// Time complexity: O(1).
func (s *ChunkTimeline) Last() (
	c *chunk.Chunk,
	nbts []map[string]any,
	updateUnixTime int64,
	err error,
) {
	var oriChunk define.ChunkMatrix = s.latestChunk
	var oriNBTs []define.NBTWithIndex = s.latestNBT

	// Blocks
	c = chunk.NewChunk(block.AirRuntimeID, s.pos.Dimension.Range())
	sub := c.Sub()
	for ChunkIndex, layers := range oriChunk {
		sub := sub[ChunkIndex]
		for index, value := range layers {
			layer := sub.Layer(uint8(index))

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
	oriNBTsCopyOne, err := define.NBTDeepCopy(oriNBTs)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("Last: %v", err)
	}
	for _, value := range oriNBTsCopyOne {
		nbts = append(nbts, value.NBT)
	}

	return c, nbts, s.timelineUnixTime[len(s.timelineUnixTime)-1], nil
}
