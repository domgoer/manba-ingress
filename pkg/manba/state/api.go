package state

import (
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	memdb "github.com/hashicorp/go-memdb"
)

// APICollection stores and indexes api information
type APICollection collection

const (
	apiTableName = "api"
)

var apiTableSchema = &memdb.TableSchema{
	Name: apiTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		"name": {
			Name:         "name",
			Unique:       true,
			Indexer:      &memdb.StringFieldIndex{Field: "Name"},
			AllowMissing: true,
		},
		all: allIndex,
	},
}

// Add adds a api to the collection
// An error is thrown if api.ID is empty.
func (c *APICollection) Add(api *metapb.API) error {
	id := id2Str(api.ID)
	if id == "" {
		return errIDRequired
	}
	txn := c.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, id)

	_, err := getAPI(txn, searchBy...)
	if err == nil {
		return ErrAlreadyExists
	} else if err != ErrNotFound {
		return err
	}

	err = txn.Insert(apiTableName, &api)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getAPI(txn *memdb.Txn, searches ...string) (*API, error) {
	for _, search := range searches {
		res, err := multiIndexLookupUsingTxn(txn, apiTableName,
			[]string{"name", "id"}, search)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}
		api, ok := res.(*API)
		if !ok {
			panic(unexpectedType)
		}
		return &API{API: *deepCopyManbaAPI(*api)}, nil
	}
	return nil, ErrNotFound
}

// Get gets a api by name or ID.
func (c *APICollection) Get(nameOrID string) (*API, error) {
	if nameOrID == "" {
		return nil, errIDRequired
	}

	txn := c.db.Txn(false)
	defer txn.Abort()
	return getAPI(txn, nameOrID)
}

// Update updates an existing api.
// It returns an error if the api is not already present.
func (c *APICollection) Update(api API) error {
	id := id2Str(api.ID)
	if id == "" {
		return errIDRequired
	}

	txn := c.db.Txn(true)
	defer txn.Abort()

	err := deleteAPI(txn, id)
	if err != nil {
		return err
	}

	err = txn.Insert(apiTableName, &api)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteAPI(txn *memdb.Txn, nameOrID string) error {
	api, err := getAPI(txn, nameOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(apiTableName, api)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a api by name or ID.
func (c *APICollection) Delete(nameOrID string) error {
	if nameOrID == "" {
		return errIDRequired
	}

	txn := c.db.Txn(true)
	defer txn.Abort()

	err := deleteAPI(txn, nameOrID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets a api by name or ID.
func (c *APICollection) GetAll() ([]*API, error) {
	txn := c.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(apiTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*API
	for el := iter.Next(); el != nil; el = iter.Next() {
		s, ok := el.(*API)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &API{API: *deepCopyManbaAPI(*s)})
	}
	txn.Commit()
	return res, nil
}

func deepCopyManbaAPI(s API) *metapb.API {
	d, _ := s.Marshal()
	res := new(metapb.API)
	res.Unmarshal(d)
	return res
}
