package state

import (
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	memdb "github.com/hashicorp/go-memdb"
)

// ServerCollection stores and indexes server information
type ServerCollection collection

const (
	serverTableName = "server"
)

var serverTableSchema = &memdb.TableSchema{
	Name: serverTableName,
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

// Add adds a server to the collection
// An error is thrown if server.ID is empty.
func (c *ServerCollection) Add(server Server) error {
	id := id2Str(server.ID)
	if id == "" {
		return errIDRequired
	}
	txn := c.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, id)

	_, err := getServer(txn, searchBy...)
	if err == nil {
		return ErrAlreadyExists
	} else if err != ErrNotFound {
		return err
	}

	err = txn.Insert(serverTableName, &server)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getServer(txn *memdb.Txn, searches ...string) (*Server, error) {
	for _, search := range searches {
		res, err := multiIndexLookupUsingTxn(txn, serverTableName,
			[]string{"name", "id"}, search)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}
		server, ok := res.(*Server)
		if !ok {
			panic(unexpectedType)
		}
		return &Server{Server: *deepCopyManbaServer(*server)}, nil
	}
	return nil, ErrNotFound
}

// Get gets a server by name or ID.
func (c *ServerCollection) Get(nameOrID string) (*Server, error) {
	if nameOrID == "" {
		return nil, errIDRequired
	}

	txn := c.db.Txn(false)
	defer txn.Abort()
	return getServer(txn, nameOrID)
}

// Update updates an existing server.
// It returns an error if the server is not already present.
func (c *ServerCollection) Update(server Server) error {
	id := id2Str(server.ID)
	if id == "" {
		return errIDRequired
	}

	txn := c.db.Txn(true)
	defer txn.Abort()

	err := deleteServer(txn, id)
	if err != nil {
		return err
	}

	err = txn.Insert(serverTableName, &server)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteServer(txn *memdb.Txn, nameOrID string) error {
	server, err := getServer(txn, nameOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(serverTableName, server)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a server by name or ID.
func (c *ServerCollection) Delete(nameOrID string) error {
	if nameOrID == "" {
		return errIDRequired
	}

	txn := c.db.Txn(true)
	defer txn.Abort()

	err := deleteServer(txn, nameOrID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets a server by name or ID.
func (c *ServerCollection) GetAll() ([]*Server, error) {
	txn := c.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(serverTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*Server
	for el := iter.Next(); el != nil; el = iter.Next() {
		s, ok := el.(*Server)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &Server{Server: *deepCopyManbaServer(*s)})
	}
	txn.Commit()
	return res, nil
}

func deepCopyManbaServer(s Server) *metapb.Server {
	d, _ := s.Marshal()
	res := new(metapb.Server)
	res.Unmarshal(d)
	return res
}
