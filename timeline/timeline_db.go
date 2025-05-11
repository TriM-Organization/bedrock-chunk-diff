package timeline

import (
	"fmt"

	"github.com/df-mc/goleveldb/leveldb"
)

// TimelineDB implements chunk timeline and
// history record provider based on LevelDB.
type TimelineDB struct {
	LevelDB
	sessions *InProgressSession
}

// NewTimelineDB open a level database that used for
// chunk delta update whose at path.
// If not exist, then create a new database.
func NewTimelineDB(path string) (result TimelineDatabase, err error) {
	timelineDB := &TimelineDB{
		sessions: NewInProgressSession(),
	}

	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, fmt.Errorf("NewTimelineDB: %v", err)
	}

	timelineDB.LevelDB = &database{ldb: db}
	return timelineDB, nil
}
