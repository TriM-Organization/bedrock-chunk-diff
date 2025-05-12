package timeline

import (
	"context"
	"fmt"

	"go.etcd.io/bbolt"
)

// TimelineDB implements chunk timeline and
// history record provider based on LevelDB.
type TimelineDB struct {
	DB
	sessions *InProgressSession
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

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists(DatabaseRootKey)
		return err
	})
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("Open: %v", err)
	}

	timelineDB.DB = &database{bdb: db}
	return timelineDB, nil
}

// CloseTimelineDB closed the timeline database.
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
