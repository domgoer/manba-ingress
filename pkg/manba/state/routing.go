package state

import (
	"reflect"

	memdb "github.com/hashicorp/go-memdb"
)

// RoutingCollection stores and indexes routing information
type RoutingCollection collection

const (
	routingTableName = "routing"
)

var routingTableSchema = &memdb.TableSchema{
	Name: routingTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "idStr"},
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

// Add adds a routing to the collection
// An error is thrown if routing.ID is empty.
func (c *RoutingCollection) Add(routing Routing) error {
	id := id2Str(routing.ID)
	if id == "" {
		return errIDRequired
	}
	routing.idStr = id
	txn := c.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, id)

	_, err := getRouting(txn, searchBy...)
	if err == nil {
		return ErrAlreadyExists
	} else if err != ErrNotFound {
		return err
	}

	err = txn.Insert(routingTableName, &routing)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getRouting(txn *memdb.Txn, searches ...string) (*Routing, error) {
	for _, search := range searches {
		res, err := multiIndexLookupUsingTxn(txn, routingTableName,
			[]string{"name", "id"}, search)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}
		routing, ok := res.(*Routing)
		if !ok {
			panic(unexpectedType)
		}
		dp := routing.DeepCopy()
		return &Routing{Routing: dp.Routing, idStr: dp.idStr}, nil
	}
	return nil, ErrNotFound
}

// Get gets a routing by name or ID.
func (c *RoutingCollection) Get(nameOrID string) (*Routing, error) {
	if nameOrID == "" {
		return nil, errIDRequired
	}

	txn := c.db.Txn(false)
	defer txn.Abort()
	return getRouting(txn, nameOrID)
}

// Update updates an existing routing.
// It returns an error if the routing is not already present.
func (c *RoutingCollection) Update(routing Routing) error {
	id := id2Str(routing.ID)
	if id == "" {
		return errIDRequired
	}

	txn := c.db.Txn(true)
	defer txn.Abort()

	err := deleteRouting(txn, id)
	if err != nil {
		return err
	}

	routing.idStr = id

	err = txn.Insert(routingTableName, &routing)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteRouting(txn *memdb.Txn, nameOrID string) error {
	routing, err := getRouting(txn, nameOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(routingTableName, routing)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a routing by name or ID.
func (c *RoutingCollection) Delete(nameOrID string) error {
	if nameOrID == "" {
		return errIDRequired
	}

	txn := c.db.Txn(true)
	defer txn.Abort()

	err := deleteRouting(txn, nameOrID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets a routing by name or ID.
func (c *RoutingCollection) GetAll() ([]*Routing, error) {
	txn := c.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(routingTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*Routing
	for el := iter.Next(); el != nil; el = iter.Next() {
		s, ok := el.(*Routing)
		if !ok {
			panic(unexpectedType)
		}
		dp := s.DeepCopy()
		res = append(res, &Routing{Routing: dp.Routing, idStr: dp.idStr})
	}
	txn.Commit()
	return res, nil
}

// CompareRouting checks two manba routings whether deep equal
func CompareRouting(r1, r2 *Routing) bool {
	d1 := r1.DeepCopy().Routing
	d2 := r2.DeepCopy().Routing

	d1.XXX_unrecognized = nil
	d2.XXX_unrecognized = nil
	return reflect.DeepEqual(&r1.Routing, &r2.Routing)
}
