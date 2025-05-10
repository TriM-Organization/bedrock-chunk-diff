package time_line

// LevelDB represent to a level database
// that implements some basic funtions.
type LevelDB interface {
	Close() error
	Delete(key []byte) error
	Get(key []byte) (value []byte, err error)
	Has(key []byte) (has bool, err error)
	Put(key []byte, value []byte) error
}
