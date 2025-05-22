package main

import (
	"log"
	"slices"
	"sync"
	"time"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/timeline"
	"github.com/TriM-Organization/bedrock-world-operator/world"
	"github.com/pterm/pterm"
	"go.etcd.io/bbolt"
)

func IterEntireDatabase(
	db timeline.TimelineDatabase,
	w world.World,
	doCompact bool,
	maxConcurrent int,
	providedUnixTime int64,
	ensureExistOne bool,
) {
	startTime := time.Now()
	counter := 0
	defer func() {
		pterm.Success.Println("Time used:", time.Since(startTime))
		pterm.Success.Println("Found chunks:", counter)
	}()

	err := db.UnderlyingDatabase().View(func(tx *bbolt.Tx) error {
		var foundKeyChunkCount bool
		var startGoRoutines = 0
		bucket := tx.Bucket(timeline.DatabaseKeyChunkIndex)
		waiter := new(sync.WaitGroup)

		err := bucket.ForEach(func(k, v []byte) error {
			if !foundKeyChunkCount && slices.Equal(k, timeline.DatabaseKeyChunkCount) {
				foundKeyChunkCount = true
				return nil
			}

			pos := define.IndexInv(k)
			counter++

			if maxConcurrent == 0 {
				waiter.Add(1)
				SingleChunkRunner(db, w, doCompact, providedUnixTime, ensureExistOne, waiter, pos)
			} else {
				if startGoRoutines > maxConcurrent {
					waiter.Wait()
					startGoRoutines = 0
				}
				startGoRoutines++
				waiter.Add(1)
				go SingleChunkRunner(db, w, doCompact, providedUnixTime, ensureExistOne, waiter, pos)
			}

			return nil
		})
		if err != nil {
			return err
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
