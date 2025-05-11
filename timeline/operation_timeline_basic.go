package timeline

import "fmt"

// Empty returns whether this timeline is empty or not.
// If is empty, then calling Save will result in no operation.
func (s *SubChunkTimeline) Empty() bool {
	return s.isEmpty
}

// SetMaxLimit sets the timeline could record how many time point.
// maxLimit must bigger than 0. If less, then set the limit to 1.
//
// After calling SetMaxLimit if overflow immediately, then we will
// pop some time point from the underlying timeline.
// Poped time points must be the most earliest one.
//
// Note that calling SetMaxLimit will not change the empty states
// of this timeline.
func (s *SubChunkTimeline) SetMaxLimit(maxLimit uint) error {
	s.maxLimit = max(maxLimit, 1)

	for s.barrierRight-s.barrierLeft+1 >= s.maxLimit {
		if err := s.Pop(); err != nil {
			return fmt.Errorf("(s *SubChunkTimeline) SetMaxLimit: %v", err)
		}
	}

	return nil
}
