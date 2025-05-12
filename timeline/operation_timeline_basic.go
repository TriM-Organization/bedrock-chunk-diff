package timeline

import "fmt"

// Empty returns whether this timeline is empty or not.
// If is empty, then calling Save will result in no operation.
func (s *ChunkTimeline) Empty() bool {
	return s.isEmpty
}

// ReadOnly returns whether this timeline is read only or not.
// If is read only, then calling any function that will modify
// underlying timeline will result in no operation.
func (s *ChunkTimeline) ReadOnly() bool {
	return s.isReadOnly
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
//
// If current timeline is read only, then calling SetMaxLimit will
// do no operation.
func (s *ChunkTimeline) SetMaxLimit(maxLimit uint) error {
	if s.isReadOnly {
		return nil
	}

	s.maxLimit = max(maxLimit, 1)

	for s.barrierRight-s.barrierLeft+1 > s.maxLimit {
		if err := s.Pop(); err != nil {
			return fmt.Errorf("(s *ChunkTimeline) SetMaxLimit: %v", err)
		}
	}

	return nil
}
