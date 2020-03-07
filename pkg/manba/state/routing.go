package state

import memdb "github.com/hashicorp/go-memdb"

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
