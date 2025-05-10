package timeline

import "github.com/df-mc/goleveldb/leveldb"

// database wrapper a level database,
// and expose some useful functions.
type database struct {
	ldb *leveldb.DB
}

// Has returns true if the DB does contains the given key.
//
// It is safe to modify the contents of the argument after Has returns.
func (db *database) Has(key []byte) (has bool, err error) {
	return db.ldb.Has(key, nil)
}

// Get gets the value for the given key. It returns ErrNotFound if the
// DB does not contains the key.
//
// The returned slice is its own copy, it is safe to modify the contents
// of the returned slice.
// It is safe to modify the contents of the argument after Get returns.
//
// Note that if the key is not exist, then return nil value and nil error.
func (db *database) Get(key []byte) (value []byte, err error) {
	value, err = db.ldb.Get(key, nil)
	if err == leveldb.ErrNotFound {
		return nil, nil
	}
	return
}

// Put sets the value for the given key. It overwrites any previous value
// for that key; a DB is not a multi-map. Write merge also applies for Put, see
// Write.
//
// It is safe to modify the contents of the arguments after Put returns but not
// before.
func (db *database) Put(key []byte, value []byte) error {
	return db.ldb.Put(key, value, nil)
}

// Delete deletes the value for the given key. Delete will not returns error if
// key doesn't exist. Write merge also applies for Delete, see Write.
//
// It is safe to modify the contents of the arguments after Delete returns but
// not before.
func (db *database) Delete(key []byte) error {
	return db.ldb.Delete(key, nil)
}

// Close closes the DB. This will also releases any outstanding snapshot,
// abort any in-flight compaction and discard open transaction.
//
// It is not safe to close a DB until all outstanding iterators are released.
// It is valid to call Close multiple times. Other methods should not be
// called after the DB has been closed.
func (db *database) Close() error {
	return db.ldb.Close()
}

// OpenTransaction opens an atomic DB transaction. Only one transaction can be
// opened at a time. Subsequent call to Write and OpenTransaction will be blocked
// until in-flight transaction is committed or discarded.
// The returned transaction handle is safe for concurrent use.
//
// Transaction is very expensive and can overwhelm compaction, especially if
// transaction size is small. Use with caution.
// The rule of thumb is if you need to merge at least same amount of
// `Options.WriteBuffer` worth of data then use transaction, otherwise don't.
//
// The transaction must be closed once done, either by committing or discarding
// the transaction.
// Closing the DB will discard open transaction.
func (db *database) OpenTransaction() (Transaction, error) {
	t, err := db.ldb.OpenTransaction()
	if err != nil {
		return nil, err
	}
	return &transaction{t: t}, nil
}

// transaction wrapper a level transaction,
// and expose some useful functions.
type transaction struct {
	t *leveldb.Transaction
}

// Has returns true if the DB does contains the given key.
//
// It is safe to modify the contents of the argument after Has returns.
func (t *transaction) Has(key []byte) (has bool, err error) {
	return t.t.Has(key, nil)
}

// Get gets the value for the given key. It returns ErrNotFound if the
// DB does not contains the key.
//
// The returned slice is its own copy, it is safe to modify the contents
// of the returned slice.
// It is safe to modify the contents of the argument after Get returns.
//
// Note that if the key is not exist, then return nil value and nil error.
func (t *transaction) Get(key []byte) (value []byte, err error) {
	value, err = t.t.Get(key, nil)
	if err == leveldb.ErrNotFound {
		return nil, nil
	}
	return
}

// Put sets the value for the given key. It overwrites any previous value
// for that key; a DB is not a multi-map. Write merge also applies for Put, see
// Write.
//
// It is safe to modify the contents of the arguments after Put returns but not
// before.
func (t *transaction) Put(key []byte, value []byte) error {
	return t.t.Put(key, value, nil)
}

// Delete deletes the value for the given key. Delete will not returns error if
// key doesn't exist. Write merge also applies for Delete, see Write.
//
// It is safe to modify the contents of the arguments after Delete returns but
// not before.
func (t *transaction) Delete(key []byte) error {
	return t.t.Delete(key, nil)
}

// Commit commits the transaction. If error is not nil, then the transaction is
// not committed, it can then either be retried or discarded.
//
// Other methods should not be called after transaction has been committed.
func (t *transaction) Commit() error {
	return t.t.Commit()
}

// Discard discards the transaction.
// This method is noop if transaction is already closed (either committed or
// discarded)
//
// Other methods should not be called after transaction has been discarded.
func (t *transaction) Discard() {
	t.t.Discard()
}
