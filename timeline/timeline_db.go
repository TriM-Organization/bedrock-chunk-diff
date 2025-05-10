package timeline

import (
	"fmt"

	"github.com/df-mc/goleveldb/leveldb"
)

// TimelineDB implements chunk timeline and
// history record provider based on LevelDB.
type TimelineDB struct {
	db LevelDB
}

// NewTimelineDB open a level database that used for
// chunk delta update whose at path.
// If not exist, then create a new database.
func NewTimelineDB(path string) (result *TimelineDB, err error) {
	result = new(TimelineDB)

	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, fmt.Errorf("NewTimelineDB: %v", err)
	}

	result.db = &database{ldb: db}
	return result, nil
}
