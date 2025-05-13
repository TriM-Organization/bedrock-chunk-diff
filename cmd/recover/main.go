package main

import (
	"flag"
	"fmt"
	"log"
	"slices"
	"sync"
	"time"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/timeline"
	"github.com/TriM-Organization/bedrock-world-operator/chunk"
	"github.com/TriM-Organization/bedrock-world-operator/world"
	"go.etcd.io/bbolt"
)

func main() {
	path := flag.String("path", "", "The path of your timeline database. e.g.")
	output := flag.String("output", "", "The path to output your Minecraft world")
	providedUnixTime := flag.Int64(
		"provided_unix_time",
		time.Now().Unix(),
		"Restore to the world closest to this time (earlier than or equal to the given time)",
	)
	noGrowSync := flag.Bool("no_grow_sync", false, "Database settings: No grow sync")
	noSync := flag.Bool("no_sync", false, "Database settings:No Sync")
	flag.Parse()

	if len(*path) == 0 {
		log.Fatalln("Please provide the path of your timeline database.\n\te.g. -path \"test\"")
	}
	if len(*output) == 0 {
		log.Fatalln("Please provide the path to output your Minecraft world.\n\te.g. -output \"mcworld\"")
	}

	db, err := timeline.Open(*path, *noGrowSync, *noSync)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.CloseTimelineDB()

	w, err := world.Open(*output)
	if err != nil {
		log.Fatalln(err)
	}
	defer w.CloseWorld()

	startTime := time.Now()
	counter := 0
	defer func() {
		fmt.Println("Time used:", time.Since(startTime))
		fmt.Println("Find chunks:", counter)
	}()

	err = db.UnderlyingDatabase().View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(timeline.DatabaseKeyChunkIndex)
		waiter := new(sync.WaitGroup)

		err = bucket.ForEach(func(k, v []byte) error {
			pos := define.IndexInv(k)
			waiter.Add(1)
			counter++
			go func() {
				defer func() {
					waiter.Done()
					fmt.Printf("Chunk (%d, %d) in dim %d is down.\n", pos.ChunkPos[0], pos.ChunkPos[1], pos.Dimension)
				}()

				tl, err := db.NewChunkTimeline(pos, true)
				if err != nil {
					return
				}
				defer tl.Save()

				if tl.Empty() {
					return
				}

				index, hit := slices.BinarySearch(tl.AllTimePoint(), *providedUnixTime)
				if hit {
					index++
				}

				if index <= 0 {
					return
				}

				var c *chunk.Chunk
				var nbts []map[string]any

				if index >= tl.AllTimePointLen() {
					c, nbts, _, err = tl.Last()
					if err != nil {
						return
					}
				} else {
					for range index {
						c, nbts, _, _, err = tl.Next()
						if err != nil {
							return
						}
					}
				}

				err = w.SaveChunk(pos.Dimension, pos.ChunkPos, c)
				if err != nil {
					return
				}
				err = w.SaveNBT(pos.Dimension, pos.ChunkPos, nbts)
				if err != nil {
					return
				}
			}()
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

	fmt.Println("ALL DOWN :)")
}
