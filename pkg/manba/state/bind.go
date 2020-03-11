package state

import (
	"fmt"
	"reflect"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	memdb "github.com/hashicorp/go-memdb"
)

// BindCollection stores and indexes bind information
type BindCollection collection

const (
	bindTableName = "bind"
)

var bindTableSchema = &memdb.TableSchema{
	Name: bindTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"name": {
			Name:         "name",
			Unique:       true,
			Indexer:      &memdb.StringFieldIndex{Field: "Name"},
			AllowMissing: true,
		},
		all: allIndex,
	},
}

// Add adds a bind to the collection
// An error is thrown if bind.ID is empty.
func (c *BindCollection) Add(bind Bind) error {
	if bind.ClusterID == 0 || bind.ServerID == 0 {
		return errIDRequired
	}
	txn := c.db.Txn(true)
	defer txn.Abort()

	bind.Name = fmt.Sprintf("%d-%d", bind.ClusterID, bind.ServerID)

	var searchBy []string
	searchBy = append(searchBy, bind.Name)

	_, err := getBind(txn, searchBy...)
	if err == nil {
		return ErrAlreadyExists
	} else if err != ErrNotFound {
		return err
	}

	err = txn.Insert(bindTableName, &bind)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getBind(txn *memdb.Txn, searches ...string) (*Bind, error) {
	for _, search := range searches {
		res, err := multiIndexLookupUsingTxn(txn, bindTableName,
			[]string{"name"}, search)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}
		bind, ok := res.(*Bind)
		if !ok {
			panic(unexpectedType)
		}
		return &Bind{Bind: *DeepCopyManbaBind(*bind)}, nil
	}
	return nil, ErrNotFound
}

// Get gets a bind by name or ID.
func (c *BindCollection) Get(nameOrID string) (*Bind, error) {
	if nameOrID == "" {
		return nil, errIDRequired
	}

	txn := c.db.Txn(false)
	defer txn.Abort()
	return getBind(txn, nameOrID)
}

// Update updates an existing bind.
// It returns an error if the bind is not already present.
func (c *BindCollection) Update(bind Bind) error {
	if bind.Name == "" {
		return errIDRequired
	}

	txn := c.db.Txn(true)
	defer txn.Abort()

	err := deleteBind(txn, bind.Name)
	if err != nil {
		return err
	}

	err = txn.Insert(bindTableName, &bind)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteBind(txn *memdb.Txn, nameOrID string) error {
	bind, err := getBind(txn, nameOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(bindTableName, bind)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a bind by name or ID.
func (c *BindCollection) Delete(nameOrID string) error {
	if nameOrID == "" {
		return errIDRequired
	}

	txn := c.db.Txn(true)
	defer txn.Abort()

	err := deleteBind(txn, nameOrID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets a bind by name or ID.
func (c *BindCollection) GetAll() ([]*Bind, error) {
	txn := c.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(bindTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*Bind
	for el := iter.Next(); el != nil; el = iter.Next() {
		s, ok := el.(*Bind)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &Bind{Bind: *DeepCopyManbaBind(*s)})
	}
	txn.Commit()
	return res, nil
}

// DeepCopyManbaBind returns new bind deep cloned by this function
func DeepCopyManbaBind(s Bind) *metapb.Bind {
	d, _ := s.Marshal()
	res := new(metapb.Bind)
	res.Unmarshal(d)
	return res
}

// CompareBind checks two manba apis whether deep equal
func CompareBind(r1, r2 *Bind) bool {
	return reflect.DeepEqual(&r1.Bind, &r2.Bind)
}
