package timeline

import (
	"encoding/binary"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
)

// LoadLatestTimePointUnixTime loads the time when latest time point update.
// If not exist, then return 0.
func (t *TimelineDB) LoadLatestTimePointUnixTime(pos define.DimSubChunk) (timeStamp int64) {
	keyBytes := define.Sum(pos, define.KeyLatestTimePointUnixTime)
	data, err := t.db.Get(keyBytes)
	if err != nil || len(data) == 0 {
		return 0
	}
	return int64(binary.LittleEndian.Uint64(data))
}

// SaveLatestTimePointUnixTime saves the time when the latest time point is generated.
// If timeStamp is 0, then delete the time from the database.
func (t *TimelineDB) SaveLatestTimePointUnixTime(pos define.DimSubChunk, timeStamp int64) error {
	keyBytes := define.Sum(pos, define.KeyLatestTimePointUnixTime)
	if timeStamp == 0 {
		return t.db.Delete(keyBytes)
	}
	timeStampBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timeStampBytes, uint64(timeStamp))
	return t.db.Put(keyBytes, timeStampBytes)
}
