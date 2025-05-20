package main

import (
	"fmt"
	"slices"
	"sync"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/timeline"
	"github.com/TriM-Organization/bedrock-world-operator/chunk"
	"github.com/TriM-Organization/bedrock-world-operator/world"
	"github.com/pterm/pterm"
)

func SingleChunkRunner(
	db timeline.TimelineDatabase,
	w world.World,
	providedUnixTime int64,
	ensureExistOne bool,
	waiter *sync.WaitGroup,
	pos define.DimChunk,
) {
	var c *chunk.Chunk
	var nbts []map[string]any

	defer func() {
		waiter.Done()
		fmt.Printf("Chunk (%d, %d) in dim %d is down.\n", pos.ChunkPos[0], pos.ChunkPos[1], pos.Dimension)
	}()

	tl, err := db.NewChunkTimeline(pos, true)
	if err != nil {
		pterm.Warning.Printf("SingleChunkRunner: %v\n", err)
		return
	}
	defer tl.Save()

	if tl.Empty() {
		return
	}

	index, hit := slices.BinarySearch(tl.AllTimePoint(), providedUnixTime)
	if !hit {
		index--
	}

	if index < 0 {
		if !ensureExistOne {
			return
		}
		index = 0
	}

	if index >= tl.AllTimePointLen() {
		c, nbts, _, err = tl.Last()
		if err != nil {
			pterm.Warning.Printf("SingleChunkRunner: %v\n", err)
			return
		}
	} else {
		c, nbts, _, err = tl.JumpTo(uint(index))
		if err != nil {
			pterm.Warning.Printf("SingleChunkRunner: %v\n", err)
			return
		}
	}

	err = w.SaveChunk(pos.Dimension, pos.ChunkPos, c)
	if err != nil {
		pterm.Warning.Printf("SingleChunkRunner: %v\n", err)
		return
	}

	err = w.SaveNBT(pos.Dimension, pos.ChunkPos, nbts)
	if err != nil {
		pterm.Warning.Printf("SingleChunkRunner: %v\n", err)
		return
	}
}
