package state

import memdb "github.com/hashicorp/go-memdb"

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
