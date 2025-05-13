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
	providedUnixTime int64,
) {
	startTime := time.Now()
	counter := 0
	defer func() {
		fmt.Println("Time used:", time.Since(startTime))
		fmt.Println("Find chunks:", counter)
	}()

	err := db.UnderlyingDatabase().View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(timeline.DatabaseKeyChunkIndex)
		waiter := new(sync.WaitGroup)

		for _, pos := range enumChunks {
			if bucket.Get(define.Index(pos)) == nil {
				continue
			}
			waiter.Add(1)
			counter++
			go SingleChunkRunner(db, w, providedUnixTime, waiter, pos)
		}

		waiter.Wait()
		return nil
	})
	if err != nil {
		log.Fatalln(err)
	}
}
