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

func IterRangeEntireDatabase(
	db timeline.TimelineDatabase,
	w world.World,
	enumChunks []define.DimChunk,
	rangeDimension int,
	providedUnixTime int64,
	ensureExistOne bool,
) {
	startTime := time.Now()
	counter := 0
	defer func() {
		fmt.Println("Time used:", time.Since(startTime))
		fmt.Println("Find chunks:", counter)
	}()

	mapping := make(map[define.DimChunk]bool)
	for _, value := range enumChunks {
		mapping[value] = true
	}

	err := db.UnderlyingDatabase().View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(timeline.DatabaseKeyChunkIndex)
		waiter := new(sync.WaitGroup)

		err := bucket.ForEach(func(k, v []byte) error {
			pos := define.IndexInv(k)
			if !mapping[pos] {
				return nil
			}
			waiter.Add(1)
			counter++
			go SingleChunkRunner(db, w, providedUnixTime, ensureExistOne, waiter, pos)
			return nil
		})
		if err != nil {
			return err
		}

		waiter.Wait()
		return nil
	})
	if err != nil {
		log.Fatalln(err)
	}
}
