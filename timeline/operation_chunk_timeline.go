package timeline

import "fmt"

// SaveNOP releases each sub chunk timeline of this chunk,
// and don't do more things (will not change ths database).
func (c *ChunkTimeline) SaveNOP() {
	for _, value := range c.subChunkTimelines {
		value.SaveNOP()
	}
}

// Save saves timelines for all sub chunks in this chunk into
// the underlying database, and also release each sub chunk timeline.
//
// That means, after calling Save, any timeline from this object
// can't be used again. Also, you can't call Save multiple times.
//
// Note that we will not check whether it has been released,
// nor will we check whether you have called Save multiple times.
//
// Save must calling at the last modification of the timeline;
// otherwise, the timeline will not be able to maintain data consistency.
func (c *ChunkTimeline) Save() error {
	for _, value := range c.subChunkTimelines {
		if err := value.Save(); err != nil {
			return fmt.Errorf("Save: %v", err)
		}
	}
	return nil
}

// Sub returns a list of timeline for all sub chunks present in the chunk.
func (c *ChunkTimeline) Sub() []*SubChunkTimeline {
	return c.subChunkTimelines
}

// SetSub overwrite the timeline for sub chunks of this chunk.
//
// The length of subChunks could less or bigger than the sub chunk counts of this
// whole chunk.
//
// If less, then only the given part will be modified,
// if bigger, then the bigger part will be not used.
func (c *ChunkTimeline) SetSub(subChunkTimelines []*SubChunkTimeline) {
	n := c.pos.Dimension.Height() >> 4
	c.subChunkTimelines = make([]*SubChunkTimeline, n)
	for i := range min(n, len(subChunkTimelines)) {
		c.subChunkTimelines[i] = subChunkTimelines[i]
	}
}

// SubChunk finds the timeline of a sub chunk whose in subChunkIndex.
//
// subChunkIndex is an integer that bigger then -1.
// For example, if a block is at (x,23,z) and is in Overworld, then
// it is in a sub chunk whose Y position is 23>>4 = 1.
//
// However, this is not the index of this sub chunk, we need use
// (23>>4) - (dm.Range()[0]>>4) to get the index, which is 1-4=-3.
func (c *ChunkTimeline) SubChunk(subChunkIndex int16) *SubChunkTimeline {
	return c.subChunkTimelines[subChunkIndex]
}

// SetSubChunk set the timeline of a sub chunk in this chunk.
//
// subChunkIndex is an integer that bigger then -1.
// For example, if a block is at (x,23,z) and is in Overworld,
// then it is in a sub chunk whose Y position is 23>>4 = 1.
//
// However, this is not the index of this sub chunk, we need
// use (23>>4) - (dm.Range()[0]>>4) to get the index, which is
// 1-4=-3.
func (c *ChunkTimeline) SetSubChunk(subChunkTimeline *SubChunkTimeline, subChunkIndex int16) {
	c.subChunkTimelines[subChunkIndex] = subChunkTimeline
}
