package state

import memdb "github.com/hashicorp/go-memdb"

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
