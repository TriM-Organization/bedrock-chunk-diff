package timeline

import (
	"go.etcd.io/bbolt"
)

var DatabaseRootKey = []byte("root")

// database wrapper a database,
// and expose some useful functions.
type database struct {
	bdb *bbolt.DB
}

// Has returns true if the DB does contains the given key.
func (db *database) Has(key []byte) (has bool) {
	db.bdb.View(func(tx *bbolt.Tx) error {
		has = (tx.Bucket(DatabaseRootKey).Get(key) != nil)
		return nil
	})
	return
}

// Get retrieves the value for a key in the bucket.
// Returns a nil value if the key does not exist or if the key is a nested bucket.
func (db *database) Get(key []byte) (value []byte) {
	db.bdb.View(func(tx *bbolt.Tx) error {
		result := tx.Bucket(DatabaseRootKey).Get(key)
		value = make([]byte, len(result))
		copy(value, result)
		return nil
	})
	return
}

// Put sets the value for a key in the bucket.
// If the key exist then its previous value will be overwritten.
// Returns an error if the key is blank, if the key is too large, or if the value is too large.
func (db *database) Put(key []byte, value []byte) (err error) {
	return db.bdb.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(DatabaseRootKey).Put(key, value)
	})
}

// Delete removes a key from the bucket.
// If the key does not exist then nothing is done and a nil error is returned.
func (db *database) Delete(key []byte) error {
	return db.bdb.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(DatabaseRootKey).Delete(key)
	})
}

// Close releases all database resources.
// It will block waiting for any open transactions to finish
// before closing the database and returning.
func (db *database) Close() error {
	return db.bdb.Close()
}

// Begin starts a new transaction.
// Multiple read-only transactions can be used concurrently but only one
// write transaction can be used at a time. Starting multiple write transactions
// will cause the calls to block and be serialized until the current write
// transaction finishes.
//
// Transactions should not be dependent on one another. Opening a read
// transaction and a write transaction in the same goroutine can cause the
// writer to deadlock because the database periodically needs to re-mmap itself
// as it grows and it cannot do that while a read transaction is open.
//
// If a long running read transaction (for example, a snapshot transaction) is
// needed, you might want to set DB.InitialMmapSize to a large enough value
// to avoid potential blocking of write transaction.
//
// IMPORTANT: You must close read-only transactions after you are finished or
// else the database will not reclaim old pages.
func (db *database) OpenTransaction() (Transaction, error) {
	tx, err := db.bdb.Begin(true)
	if err != nil {
		return nil, err
	}
	return &transaction{tx: tx}, nil
}

// transaction wrapper a database transaction,
// and expose some useful functions.
type transaction struct {
	tx *bbolt.Tx
}

// Has returns true if the DB does contains the given key.
func (t *transaction) Has(key []byte) (has bool) {
	return (t.tx.Bucket(DatabaseRootKey).Get(key) != nil)
}

// Get retrieves the value for a key in the bucket.
// Returns a nil value if the key does not exist or if the key is a nested bucket.
func (t *transaction) Get(key []byte) (value []byte) {
	return t.tx.Bucket(DatabaseRootKey).Get(key)
}

// Put sets the value for a key in the bucket.
// If the key exist then its previous value will be overwritten.
// Returns an error if the key is blank, if the key is too large, or if the value is too large.
func (t *transaction) Put(key []byte, value []byte) error {
	return t.tx.Bucket(DatabaseRootKey).Put(key, value)
}

// Delete removes a key from the bucket.
// If the key does not exist then nothing is done and a nil error is returned.
func (t *transaction) Delete(key []byte) error {
	return t.tx.Bucket(DatabaseRootKey).Delete(key)
}

// Commit writes all changes to disk, updates the meta page and closes the transaction.
// Returns an error if a disk write error occurs, or if Commit is
// called on a read-only transaction.
func (t *transaction) Commit() error {
	return t.tx.Commit()
}

// Discard closes the transaction and ignores all previous updates.
func (t *transaction) Discard() error {
	return t.tx.Rollback()
}
