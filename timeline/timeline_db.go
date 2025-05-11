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
func Open(path string) (result TimelineDatabase, err error) {
	timelineDB := &TimelineDB{
		sessions: NewInProgressSession(),
	}

	db, err := bbolt.Open(path, 0600, &bbolt.Options{
		FreelistType: bbolt.FreelistMapType,
		NoSync:       true,
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
