package timeline

import "github.com/TriM-Organization/bedrock-chunk-diff/define"

// DatabaseOperation represents some basic
// operation that a leveldb database should
// be implement.
type DatabaseOperation interface {
	Delete(key []byte) error
	Get(key []byte) (value []byte, err error)
	Has(key []byte) (has bool, err error)
	Put(key []byte, value []byte) error
}

// Transaction represents a transaction in leveldb.
type Transaction interface {
	DatabaseOperation
	Commit() error
	Discard()
}

// LevelDB represent to a level database
// that implements some basic funtions.
type LevelDB interface {
	DatabaseOperation
	OpenTransaction() (Transaction, error)
	Close() error
}

// Timeline is the function that timeline database should to implement.
type Timeline interface {
	DeleteChunkTimeline(pos define.DimChunk) error
	DeleteSubChunkTimeline(pos define.DimSubChunk) error
	LoadLatestTimePointUnixTime(pos define.DimSubChunk) (timeStamp int64)
	NewChunkTimeline(pos define.DimChunk) (result *ChunkTimeline, err error)
	NewSubChunkTimeline(pos define.DimSubChunk) (result *SubChunkTimeline, err error)
	SaveLatestTimePointUnixTime(pos define.DimSubChunk, timeStamp int64) error
}

// TimelineDatabase wrapper and implements all features from Timeline,
// and as a provider to provide timeline of chunk/sub chunk related functions.
type TimelineDatabase interface {
	LevelDB
	Timeline
	CloseTimelineDB() error
}
