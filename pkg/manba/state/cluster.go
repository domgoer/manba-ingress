package state

import (
	"reflect"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	memdb "github.com/hashicorp/go-memdb"
)

// ClusterCollection stores and indexes cluster information
type ClusterCollection collection

const (
	clusterTableName = "cluster"
)

var clusterTableSchema = &memdb.TableSchema{
	Name: clusterTableName,
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

// Add adds a cluster to the collection
// An error is thrown if cluster.ID is empty.
func (c *ClusterCollection) Add(cluster Cluster) error {
	id := id2Str(cluster.ID)
	if id == "" {
		return errIDRequired
	}
	cluster.idStr = id
	txn := c.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, id)

	_, err := getCluster(txn, searchBy...)
	if err == nil {
		return ErrAlreadyExists
	} else if err != ErrNotFound {
		return err
	}

	err = txn.Insert(clusterTableName, &cluster)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getCluster(txn *memdb.Txn, searches ...string) (*Cluster, error) {
	for _, search := range searches {
		res, err := multiIndexLookupUsingTxn(txn, clusterTableName,
			[]string{"name", "id"}, search)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}
		cluster, ok := res.(*Cluster)
		if !ok {
			panic(unexpectedType)
		}
		return &Cluster{Cluster: *DeepCopyManbaCluster(cluster)}, nil
	}
	return nil, ErrNotFound
}

// Get gets a cluster by name or ID.
func (c *ClusterCollection) Get(nameOrID string) (*Cluster, error) {
	if nameOrID == "" {
		return nil, errIDRequired
	}

	txn := c.db.Txn(false)
	defer txn.Abort()
	return getCluster(txn, nameOrID)
}

// Update updates an existing cluster.
// It returns an error if the cluster is not already present.
func (c *ClusterCollection) Update(cluster Cluster) error {
	id := id2Str(cluster.ID)
	if id == "" {
		return errIDRequired
	}

	txn := c.db.Txn(true)
	defer txn.Abort()

	err := deleteCluster(txn, id)
	if err != nil {
		return err
	}

	err = txn.Insert(clusterTableName, &cluster)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteCluster(txn *memdb.Txn, nameOrID string) error {
	cluster, err := getCluster(txn, nameOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(clusterTableName, cluster)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a cluster by name or ID.
func (c *ClusterCollection) Delete(nameOrID string) error {
	if nameOrID == "" {
		return errIDRequired
	}

	txn := c.db.Txn(true)
	defer txn.Abort()

	err := deleteCluster(txn, nameOrID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets a cluster by name or ID.
func (c *ClusterCollection) GetAll() ([]*Cluster, error) {
	txn := c.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(clusterTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*Cluster
	for el := iter.Next(); el != nil; el = iter.Next() {
		s, ok := el.(*Cluster)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &Cluster{Cluster: *DeepCopyManbaCluster(s)})
	}
	txn.Commit()
	return res, nil
}

// DeepCopyManbaCluster returns new cluster deep cloned by this function
func DeepCopyManbaCluster(s *Cluster) *metapb.Cluster {
	res := new(metapb.Cluster)
	deepCopyManbaStruct(s, res)
	return res
}

// CompareCluster checks two manba clusters whether deep equal
func CompareCluster(r1, r2 *Cluster) bool {
	d1 := DeepCopyManbaCluster(r1)
	d2 := DeepCopyManbaCluster(r2)

	d1.XXX_unrecognized = nil
	d2.XXX_unrecognized = nil
	return reflect.DeepEqual(&r1.Cluster, &r2.Cluster)
}
