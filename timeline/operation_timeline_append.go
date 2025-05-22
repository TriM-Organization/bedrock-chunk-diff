package timeline

import (
	"fmt"
	"time"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/marshal"
	"github.com/TriM-Organization/bedrock-world-operator/chunk"
)

// "appendBlocks" is an internal implement detail.
func (s *ChunkTimeline) appendBlocks(
	newerChunk define.ChunkMatrix,
	chunkDiff define.ChunkDiffMatrix,
	transaction Transaction,
) error {
	for s.barrierRight-s.barrierLeft+1 >= s.maxLimit {
		if err := s.Pop(); err != nil {
			return fmt.Errorf("appendBlocks: %v", err)
		}
	}

	// Put delta update
	payload, err := marshal.ChunkDiffMatrixToBytes(chunkDiff)
	if err != nil {
		return fmt.Errorf("appendBlocks: %v", err)
	}
	err = transaction.Put(
		define.IndexBlockDu(s.pos, s.barrierRight+1),
		payload,
	)
	if err != nil {
		return fmt.Errorf("appendBlocks: %v", err)
	}

	// Update Latest Chunk
	payload, err = marshal.ChunkMatrixToBytes(newerChunk)
	if err != nil {
		return fmt.Errorf("appendBlocks: %v", err)
	}
	err = transaction.Put(
		define.Sum(s.pos, define.KeyLatestChunk),
		payload,
	)
	if err != nil {
		return fmt.Errorf("appendBlocks: %v", err)
	}

	return nil
}

// "appendNBTs" is an internal implement detail.
func (s *ChunkTimeline) appendNBTs(
	newerNBTs []define.NBTWithIndex,
	nbtDiff define.MultipleDiffNBT,
	transaction Transaction,
) error {
	for s.barrierRight-s.barrierLeft+1 >= s.maxLimit {
		if err := s.Pop(); err != nil {
			return fmt.Errorf("appendNBTs: %v", err)
		}
	}

	// Put delta update
	payload, err := marshal.MultipleDiffNBTBytes(nbtDiff)
	if err != nil {
		return fmt.Errorf("appendNBTs: %v", err)
	}
	err = transaction.Put(
		define.IndexNBTDu(s.pos, s.barrierRight+1),
		payload,
	)
	if err != nil {
		return fmt.Errorf("appendNBTs: %v", err)
	}

	// Update Latest NBT
	payload, err = marshal.BlockNBTBytes(newerNBTs)
	if err != nil {
		return fmt.Errorf("appendNBTs: %v", err)
	}
	err = transaction.Put(
		define.Sum(s.pos, []byte(define.KeyLatestNBT)...),
		payload,
	)
	if err != nil {
		return fmt.Errorf("appendNBTs: %v", err)
	}

	return nil
}

// Append tries append a new chunk with block
// NBT data to the timeline of current chunk.
//
// If NOPWhenNoChange is true, then if their
// is no change between the one that want to
// append and the latest one, then finally will
// result in NOP.
//
// Calling Append will make sure there is exist
// at least one empty space to place the new time
// point, whether new time point will be added in
// the end or not.
//
// The way to leave empty space is by calling 
// Pop, and the poped time points must be the 
// most earliest one.
//
// If current timeline is read only, then calling
// Append will do no operation.
func (s *ChunkTimeline) Append(
	c *chunk.Chunk, nbts []map[string]any,
	NOPWhenNoChange bool,
) error {
	var success bool
	var newerChunk define.ChunkMatrix
	var newerNBTs []define.NBTWithIndex

	if s.isReadOnly {
		return nil
	}

	for s.barrierRight-s.barrierLeft+1 >= s.maxLimit {
		if err := s.Pop(); err != nil {
			return fmt.Errorf("(s *ChunkTimeline) Append: %v", err)
		}
	}

	transaction, err := s.db.OpenTransaction()
	if err != nil {
		return fmt.Errorf("(s *ChunkTimeline) Append: %v", err)
	}
	defer func() {
		if !success {
			_ = transaction.Discard()
			return
		}
		_ = transaction.Commit()
	}()

	// Blocks
	newerChunk = define.ChunkToMatrix(c, s.blockPalette)
	chunkDiff := define.ChunkDifference(s.latestChunk, newerChunk)

	// NBTs
	newerNBTs = define.FromChunkNBT(s.pos.ChunkPos, nbts)
	nbtDiff, err := define.NBTDifference(s.latestNBT, newerNBTs)
	if err != nil {
		return fmt.Errorf("(s *ChunkTimeline) Append: %v", err)
	}

	// NOP Check
	if NOPWhenNoChange && define.ChunkNoChange(chunkDiff) && define.NBTNoChange(*nbtDiff) {
		return nil
	}

	// Append
	err = s.appendBlocks(newerChunk, chunkDiff, transaction)
	if err != nil {
		return fmt.Errorf("(s *ChunkTimeline) Append: %v", err)
	}
	err = s.appendNBTs(newerNBTs, *nbtDiff, transaction)
	if err != nil {
		return fmt.Errorf("(s *ChunkTimeline) Append: %v", err)
	}

	s.latestChunk = newerChunk
	s.latestNBT = newerNBTs
	s.barrierRight++
	s.timelineUnixTime = append(s.timelineUnixTime, time.Now().Unix())
	success = true

	if s.isEmpty {
		s.barrierLeft = s.barrierRight
		s.ptr = s.barrierLeft
		s.isEmpty = false
	}

	return nil
}
