package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/timeline"
	operator_define "github.com/TriM-Organization/bedrock-world-operator/define"
	"github.com/TriM-Organization/bedrock-world-operator/world"
	"go.etcd.io/bbolt"
)

var (
	path             *string
	output           *string
	useRange         *bool
	rangeDimension   *int
	rangeStartX      *int
	rangeStartZ      *int
	rangeEndX        *int
	rangeEndZ        *int
	providedUnixTime *int64
	noGrowSync       *bool
	noSync           *bool
)

func init() {
	path = flag.String("path", "", "The path of your timeline database.")
	output = flag.String("output", "", "The path to output your Minecraft world.")

	useRange = flag.Bool("use-range", false, "If you would like recover the part of the world, but not the entire.")
	rangeDimension = flag.Int("range-dimension", 0, "Where to find these chunks (only for use-range flag)")
	rangeStartX = flag.Int("range-start-x", 0, "The starting point X coordinate to be restored.")
	rangeStartZ = flag.Int("range-start-z", 0, "The starting point Z coordinate to be restored.")
	rangeEndX = flag.Int("range-end-x", 0, "The ending point X coordinate to be restored.")
	rangeEndZ = flag.Int("range-end-z", 0, "The ending point Z coordinate to be restored.")

	providedUnixTime = flag.Int64(
		"provided_unix_time",
		time.Now().Unix(),
		"Restore to the world closest to this time (earlier than or equal to the given time).",
	)

	noGrowSync = flag.Bool("no-grow-sync", false, "Database settings: No grow sync.")
	noSync = flag.Bool("no-sync", false, "Database settings: No Sync.")

	flag.Parse()
	if len(*path) == 0 {
		log.Fatalln("Please provide the path of your timeline database.\n\te.g. -path \"test\"")
	}
	if len(*output) == 0 {
		log.Fatalln("Please provide the path to output your Minecraft world.\n\te.g. -output \"mcworld\"")
	}
}

func main() {
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

	if *useRange {
		var shouldIterEntire bool

		startX := int32(min(*rangeStartX, *rangeEndX))
		startZ := int32(min(*rangeStartZ, *rangeEndZ))
		endX := int32(max(*rangeStartX, *rangeEndX))
		endZ := int32(max(*rangeStartZ, *rangeEndZ))

		enumChunks := make([]define.DimChunk, 0)
		for x := startX; x <= endX; x++ {
			for z := startZ; z <= endZ; z++ {
				enumChunks = append(enumChunks, define.DimChunk{
					Dimension: operator_define.Dimension(*rangeDimension),
					ChunkPos:  operator_define.ChunkPos{x, z},
				})
			}
		}

		err = db.UnderlyingDatabase().View(func(tx *bbolt.Tx) error {
			countBytes := tx.Bucket(timeline.DatabaseKeyChunkIndex).Get(timeline.DatabaseKeyChunkCount)
			if len(countBytes) < 4 {
				countBytes = make([]byte, 4)
			}
			if binary.LittleEndian.Uint32(countBytes) < uint32(len(enumChunks)) {
				shouldIterEntire = true
			}
			return nil
		})
		if err != nil {
			log.Fatalln(err)
		}

		if shouldIterEntire {
			IterRangeEntireDatabase(db, w, enumChunks, *rangeDimension, *providedUnixTime)
		} else {
			IterRange(db, w, enumChunks, *rangeDimension, *providedUnixTime)
		}
	} else {
		IterEntireDatabase(db, w, *providedUnixTime)
	}

	fmt.Println("ALL DOWN :)")
}
