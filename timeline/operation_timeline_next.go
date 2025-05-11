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
func (s *SubChunkTimeline) next() (
	oriLayers define.Layers, oriNBTs []define.NBTWithIndex, updateUnixTime int64,
	isLastElement bool, err error,
) {
	if s.isEmpty {
		return nil, nil, 0, false, nil
	}
	isLastElement = (s.ptr == s.barrierRight)

	// Blocks
	{
		payload, err := s.db.Get(
			define.IndexBlockDu(s.pos, s.ptr),
		)
		if err != nil {
			return nil, nil, 0, false, fmt.Errorf("next: %v", err)
		}

		diff, err := marshal.BytesToLayersDiff(payload)
		if err != nil {
			return nil, nil, 0, false, fmt.Errorf("next: %v", err)
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
	}

	// NBTs
	{
		payload, err := s.db.Get(
			define.IndexNBTDu(s.pos, s.ptr),
		)
		if err != nil {
			return nil, nil, 0, false, fmt.Errorf("next: %v", err)
		}

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

	s.currentSubChunk = oriLayers
	s.currentNBT = oriNBTs
	s.ptr++

	if s.ptr > s.barrierRight {
		s.ptr = s.barrierLeft
		s.currentSubChunk = define.Layers{}
		s.currentNBT = nil
	}

	return oriLayers, oriNBTs, updateUnixTime, isLastElement, nil
}

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

	oriLayers, oriNBTs, updateUnixTime, isLastElement, err = s.next()
	if err != nil {
		return nil, nil, 0, false, fmt.Errorf("(s *SubChunkTimeline) Next: %v", err)
	}

	// Blocks
	subChunk = chunk.NewSubChunk(block.AirRuntimeID)
	for index, value := range oriLayers {
		layer := subChunk.Layer(uint8(index))

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

	// NBTs
	nbts = make([]map[string]any, 0)
	for _, value := range oriNBTs {
		nbts = append(nbts, value.NBT)
	}

	return
}
