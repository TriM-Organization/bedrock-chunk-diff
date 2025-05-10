package timeline

import (
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/marshal"
	"github.com/TriM-Organization/bedrock-world-operator/block"
	"github.com/TriM-Organization/bedrock-world-operator/chunk"
)

// Next gets the next time point of current sub chunk and the NBT blocks in it.
//
// With the call to Next, ensure that the returned time keeps increasing until
// the entire time series is traversed.
//
// isLastElement can inform whether the element obtained after the current call to
// Next is at the end of the time series.
//
// When it is already at the end of the timeline, calling Next again will back to
// the earliest time point. In other words, Next is self-loop and can be called continuously.
func (s *SubChunkTimeline) Next() (
	subChunk *chunk.SubChunk, nbts []map[string]any, updateUnixTime int64,
	isLastElement bool, err error,
) {
	var oriLayers define.Layers
	var oriNBTs []define.NBTWithIndex

	if s.isEmpty {
		return nil, nil, 0, false, nil
	}

	if s.ptr > s.barrierRight {
		s.ptr = s.barrierLeft
		s.currentSubChunk = define.Layers{}
		s.currentNBT = nil
	}
	isLastElement = (s.ptr == s.barrierRight)

	// Blocks
	{
		payload, err := s.db.Get(
			define.IndexBlockDu(s.dm, s.position, s.subChunkIndex, s.ptr),
		)
		if err != nil {
			return nil, nil, 0, false, fmt.Errorf("(s *SubChunkTimeline) Next: %v", err)
		}

		diff, err := marshal.BytesToLayersDiff(payload)
		if err != nil {
			return nil, nil, 0, false, fmt.Errorf("(s *SubChunkTimeline) Next: %v", err)
		}

		for index := range s.currentSubChunk {
			_ = oriLayers.Layer(index)
		}
		for index := range diff {
			_ = oriLayers.Layer(index)
		}

		for index := range oriLayers {
			oriLayers[index] = define.Restore(s.currentSubChunk.Layer(index), diff.Layer(index))
		}

		subChunk = chunk.NewSubChunk(block.AirRuntimeID)
		for index, value := range oriLayers {
			layer := subChunk.Layer(uint8(index))

			ptr := 0
			for x := range uint8(16) {
				for y := range uint8(16) {
					for z := range uint8(16) {
						layer.Set(x, y, z, s.BlockRuntimeID(value[ptr]))
						ptr++
					}
				}
			}
		}
	}

	// NBTs
	{
		payload, err := s.db.Get(
			define.IndexNBTDu(s.dm, s.position, s.subChunkIndex, s.ptr),
		)
		if err != nil {
			return nil, nil, 0, false, fmt.Errorf("(s *SubChunkTimeline) Next: %v", err)
		}

		diff, err := marshal.BytesToMultipleDiffNBT(payload)
		if err != nil {
			return nil, nil, 0, false, fmt.Errorf("(s *SubChunkTimeline) Next: %v", err)
		}

		oriNBTs, err = define.NBTRestore(s.currentNBT, diff)
		if err != nil {
			return nil, nil, 0, false, fmt.Errorf("(s *SubChunkTimeline) Next: %v", err)
		}

		nbts = make([]map[string]any, 0)
		for _, value := range oriNBTs {
			nbts = append(nbts, value.NBT)
		}
	}

	s.currentSubChunk = oriLayers
	s.currentNBT = oriNBTs
	s.ptr++

	return subChunk, nbts, s.timelineUnixTime[s.ptr-s.barrierLeft], isLastElement, nil
}
