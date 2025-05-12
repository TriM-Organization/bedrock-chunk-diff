package timeline

import (
	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"go.etcd.io/bbolt"
)

// DatabaseOperation represents some basic
// operation that a database should be implement.
type DatabaseOperation interface {
	Delete(key []byte) error
	Get(key []byte) (value []byte)
	Has(key []byte) (has bool)
	Put(key []byte, value []byte) (err error)
}

// Transaction represents a transaction in database.
type Transaction interface {
	DatabaseOperation
	Commit() error
	Discard() error
}

// DB represent to a database that implements some basic funtions.
type DB interface {
	DatabaseOperation
	OpenTransaction() (Transaction, error)
	Close() error
}

// Timeline is the function that timeline database should to implement.
type Timeline interface {
	DeleteChunkTimeline(pos define.DimChunk) error
	LoadLatestTimePointUnixTime(pos define.DimChunk) (timeStamp int64)
	NewChunkTimeline(pos define.DimChunk, readOnly bool) (result *ChunkTimeline, err error)
	SaveLatestTimePointUnixTime(pos define.DimChunk, timeStamp int64) error
}

// TimelineDatabase wrapper and implements all features from Timeline,
// and as a provider to provide timeline of chunk related functions.
type TimelineDatabase interface {
	DB
	Timeline
	UnderlyingDatabase() *bbolt.DB
	CloseTimelineDB() error
}
