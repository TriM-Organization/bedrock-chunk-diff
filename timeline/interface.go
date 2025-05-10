package timeline

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
