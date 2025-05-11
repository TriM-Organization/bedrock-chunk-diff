package timeline

// Empty returns whether this timeline is empty or not.
// If is empty, then calling Save will result in no operation.
func (s *SubChunkTimeline) Empty() bool {
	return s.isEmpty
}

// SetMaxLimit sets the timeline could record how many time point.
// maxLimit must bigger than 0. If less, then set the limit to 1.
//
// Note that calling SetMaxLimit will not change the empty states
// of this timeline.
func (s *SubChunkTimeline) SetMaxLimit(maxLimit uint) {
	s.maxLimit = max(maxLimit, 1)
}
