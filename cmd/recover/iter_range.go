package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/timeline"
	"github.com/TriM-Organization/bedrock-world-operator/world"
	"go.etcd.io/bbolt"
)

func IterRange(
	db timeline.TimelineDatabase,
	w world.World,
	enumChunks []define.DimChunk,
	rangeDimension int,
	maxConcurrent int,
	providedUnixTime int64,
	ensureExistOne bool,
) {
	startTime := time.Now()
	counter := 0
	defer func() {
		fmt.Println("Time used:", time.Since(startTime))
		fmt.Println("Found chunks:", counter)
	}()

	err := db.UnderlyingDatabase().View(func(tx *bbolt.Tx) error {
		var startGoRoutines = 0

		bucket := tx.Bucket(timeline.DatabaseKeyChunkIndex)
		waiter := new(sync.WaitGroup)

		for _, pos := range enumChunks {
			if bucket.Get(define.Index(pos)) == nil {
				continue
			}
			counter++

			if maxConcurrent == 0 {
				SingleChunkRunner(db, w, providedUnixTime, ensureExistOne, waiter, pos)
			} else {
				if startGoRoutines > maxConcurrent {
					waiter.Wait()
					startGoRoutines = 0
				}
				startGoRoutines++
				waiter.Add(1)
				go SingleChunkRunner(db, w, providedUnixTime, ensureExistOne, waiter, pos)
			}
		}

		if maxConcurrent != 0 {
			waiter.Wait()
		}
		return nil
	})
	if err != nil {
		log.Fatalln(err)
	}
}
