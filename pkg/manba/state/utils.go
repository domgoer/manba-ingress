package state

import (
	"errors"
	"strconv"

	memdb "github.com/hashicorp/go-memdb"
)

var (
	errIDRequired = errors.New("ID is required")
	// ErrNotFound is an error type that is
	// returned when an entity is not found in the state.
	ErrNotFound = errors.New("entity not found")

	// ErrAlreadyExists represents an entity is already present in the state.
	ErrAlreadyExists = errors.New("entity already exists")
)

const (
	all = "all"

	unexpectedType = "unexpected type found"
)

var allIndex = &memdb.IndexSchema{
	Name: all,
	Indexer: &memdb.ConditionalIndex{
		Conditional: func(v interface{}) (bool, error) {
			return true, nil
		},
	},
}

// multiIndexLookup can be used to search for an entity
// based on search on multiple indexes with same key.
func multiIndexLookup(memdb *memdb.MemDB, tableName string,
	indices []string,
	args ...interface{}) (interface{}, error) {

	txn := memdb.Txn(false)
	defer txn.Abort()

	for _, indexName := range indices {
		res, err := txn.First(tableName, indexName, args...)
		if res == nil && err == nil {
			continue
		}
		if err != nil {
			return nil, err
		}
		if res != nil {
			return res, nil
		}
	}
	return nil, ErrNotFound
}

// multiIndexLookupUsingTxn can be used to search for an entity
// based on search on multiple indexes with same key.
func multiIndexLookupUsingTxn(txn *memdb.Txn, tableName string,
	indices []string,
	args ...interface{}) (interface{}, error) {

	for _, indexName := range indices {
		res, err := txn.First(tableName, indexName, args...)
		if res == nil && err == nil {
			continue
		}
		if err != nil {
			return nil, err
		}
		if res != nil {
			return res, nil
		}
	}
	return nil, ErrNotFound
}

func id2Str(id uint64) string {
	return strconv.Itoa(int(id))
}