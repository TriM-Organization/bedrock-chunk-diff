package timeline

import (
	"context"
	"fmt"

	"github.com/df-mc/goleveldb/leveldb"
)

// TimelineDB implements chunk timeline and
// history record provider based on LevelDB.
type TimelineDB struct {
	LevelDB
	sessions *InProgressSession
}

// Open open a level database that used for
// chunk delta update whose at path.
// If not exist, then create a new database.
func Open(path string) (result TimelineDatabase, err error) {
	timelineDB := &TimelineDB{
		sessions: NewInProgressSession(),
	}

	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, fmt.Errorf("Open: %v", err)
	}

	timelineDB.LevelDB = &database{ldb: db}
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
