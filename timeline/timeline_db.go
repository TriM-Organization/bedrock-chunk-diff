package timeline

import (
	"context"
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/utils"
	"go.etcd.io/bbolt"
)

// TimelineDB implements chunk timeline and
// history record provider based on bbolt.
type TimelineDB struct {
	DB
	sessions   *InProgressSession
	compresser *utils.Compresser
}

// Open open a level database that used for
// chunk delta update whose at path.
// If not exist, then create a new database.
//
// When noGrowSync is true, skips the truncate call when growing the database.
// Setting this to true is only safe on non-ext3/ext4 systems.
// Skipping truncation avoids preallocation of hard drive space and
// bypasses a truncate() and fsync() syscall on remapping.
//   - See also: https://github.com/boltdb/bolt/issues/284
//
// Setting the NoSync flag will cause the database to skip fsync()
// calls after each commit. This can be useful when bulk loading data
// into a database and you can restart the bulk load in the event of
// a system failure or database corruption. Do not set this flag for
// normal use.
//
// If the package global IgnoreNoSync constant is true, this value is
// ignored.  See the comment on that constant for more details.
//
// THIS IS UNSAFE. PLEASE USE WITH CAUTION.
func Open(path string, noGrowSync bool, noSync bool) (result TimelineDatabase, err error) {
	timelineDB := &TimelineDB{
		sessions: NewInProgressSession(),
	}

	db, err := bbolt.Open(path, 0600, &bbolt.Options{
		FreelistType: bbolt.FreelistMapType,
		NoGrowSync:   noGrowSync,
		NoSync:       noSync,
	})
	if err != nil {
		return nil, fmt.Errorf("Open: %v", err)
	}

	err = db.Update(func(tx *bbolt.Tx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", err)
			}
		}()

		// DatabaseKeyRoot
		var databaseIsInit bool
		_, err = tx.CreateBucket(DatabaseKeyRoot)
		databaseIsInit = (err != nil)

		// DatabaseKeyChunkIndex/DatabaseSubKeyChunkCount
		bucket, err := tx.CreateBucketIfNotExists(DatabaseKeyChunkIndex)
		if err != nil {
			return err
		}
		if len(bucket.Get(DatabaseSubKeyChunkCount)) < 4 {
			err = bucket.Put(DatabaseSubKeyChunkCount, make([]byte, 4))
			if err != nil {
				return err
			}
		}

		// DatabaseKeyMetadata
		{
			bucket, err := tx.CreateBucketIfNotExists(DatabaseKeyMetadata)
			if err != nil {
				return err
			}

			// DatabaseSubKeyVersion
			if len(bucket.Get(DatabaseSubKeyVersion)) == 0 {
				if databaseIsInit {
					err = bucket.Put(DatabaseSubKeyCompressMethod, CompressMethodBytesByID(CompressMethodGzip))
					if err != nil {
						return err
					}
				}
				if err = bucket.Put(DatabaseSubKeyVersion, DatabaseCurrentVersion); err != nil {
					return err
				}
			}

			// DatabaseSubKeyCompressMethod
			compressMethodBytes := bucket.Get(DatabaseSubKeyCompressMethod)
			if len(compressMethodBytes) == 0 {
				bucket.Put(DatabaseSubKeyCompressMethod, CompressMethodBytesByID(CompressMethodZlib))
				timelineDB.compresser = CompresserFuncByID(CompressMethodZlib)
			} else {
				timelineDB.compresser = CompresserFuncByID(CompressMethodByBytes(compressMethodBytes))
			}
		}

		return nil
	})
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("Open: %v", err)
	}

	timelineDB.DB = &database{bdb: db}
	return timelineDB, nil
}

// CloseTimelineDB closes the timeline database.
// It will wait until all the timelines in use
// are released before closing the database.
func (t *TimelineDB) CloseTimelineDB() error {
	allPendingCtx := make([]context.Context, 0)

	t.sessions.mu.Lock()
	for _, value := range t.sessions.session {
		allPendingCtx = append(allPendingCtx, value)
	}
	t.sessions.closed = true
	t.sessions.mu.Unlock()

	for _, value := range allPendingCtx {
		<-value.Done()
	}

	err := t.Close()
	if err != nil {
		return fmt.Errorf("CloseTimelineDB: %v", err)
	}
	return nil
}

// UnderlyingDatabase returns the underlying database of this timeline database.
// Only should be calling when need iter all timelines for all chunks.
func (t *TimelineDB) UnderlyingDatabase() *bbolt.DB {
	return t.DB.(*database).bdb
}
