package timeline

import (
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
)

// ChunkTimeline holds the timeline for a whole chunk.
type ChunkTimeline struct {
	pos               define.DimChunk
	subChunkTimelines []*SubChunkTimeline
}

// NewChunkTimeline gets the timeline of a chunk who is at pos.
//
// Note that if timeline of a sub chunk that who is in this chunk is not exist, then we will
// not create a timeline but set an empty one so you can modify it. The time to create the
// timeline is only when you save a timeline that not empty to the database.
//
// Important:
//
//   - Once any modifications have been made to the returned timeline, you must save them
//     at the end; otherwise, the timeline will not be able to maintain data consistency
//     (only need to save at the last modification).
//
//   - A chunk timeline is actually holding each timeline of their sub chunks.
//     And due to we don't allow multiple threads use the same sub chunk timeline,
//     then it's possible to get blocking after alling NewChunkTimeline when a sub chunk
//     of this chunk is still using on.
//
//   - Calling ChunkTimeline.Save to release the timeline.
//
//   - Returned ChunkTimeline can't shared with multiple threads, and it's your responsibility
//     to ensure this thing.
func (t *TimelineDB) NewChunkTimeline(pos define.DimChunk) (result *ChunkTimeline, err error) {
	result = &ChunkTimeline{
		pos: pos,
	}

	n := pos.Dimension.Height() >> 4
	for i := range n {
		s, err := t.NewSubChunkTimeline(define.DimSubChunk{
			Dimension:     pos.Dimension,
			ChunkPos:      pos.ChunkPos,
			SubChunkIndex: uint8(i),
		})
		if err != nil {
			return nil, fmt.Errorf("NewChunkTimeline: %v", err)
		}
		result.subChunkTimelines = append(result.subChunkTimelines, s)
	}

	return
}

// DeleteChunkTimeline delete the timeline of a chunk who at pos.
//
// Time complexity: O(n×L).
// n represents the average number of time points that each sub
// chunk has, and L is the count of sub chunks.
func (t *TimelineDB) DeleteChunkTimeline(pos define.DimChunk) error {
	n := pos.Dimension.Height() >> 4
	for i := range n {
		err := t.DeleteSubChunkTimeline(define.DimSubChunk{
			Dimension:     pos.Dimension,
			ChunkPos:      pos.ChunkPos,
			SubChunkIndex: uint8(i),
		})
		if err != nil {
			return fmt.Errorf("DeleteChunkTimeline: %v", err)
		}
	}
	return nil
}
