package timeline

// DefaultMaxLimit is the default maxim time points that a single chunk timeline could have.
const DefaultMaxLimit = 7

// CompressMethod is the way to compress data that record in each chunk timeline.
// To be conveniently, we use the same compress method for all timelines.
const (
	CompressMethodGzip uint8 = iota // legacy
	CompressMethodZlib              // default
)

// DatabaseCurrentVersion represents the version of current database.
// It is not equal to the package version.
var DatabaseCurrentVersion = []byte{0, 0, 1}

// These are the key name of the timeline database that we used.
var (
	DatabaseKeyRoot = []byte("root")

	DatabaseKeyMetadata          = []byte("meta-data")
	DatabaseSubKeyVersion        = []byte("version")
	DatabaseSubKeyCompressMethod = []byte("compress-method")

	DatabaseKeyChunkIndex    = []byte("chunk-index")
	DatabaseSubKeyChunkCount = []byte("chunk-count")
)
