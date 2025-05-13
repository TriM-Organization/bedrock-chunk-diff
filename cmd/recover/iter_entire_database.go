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

func IterEntireDatabase(
	db timeline.TimelineDatabase,
	w world.World,
	providedUnixTime *int64,
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

		err := bucket.ForEach(func(k, v []byte) error {
			pos := define.IndexInv(k)
			waiter.Add(1)
			counter++
			go SingleChunkRunner(db, w, providedUnixTime, waiter, pos)
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
