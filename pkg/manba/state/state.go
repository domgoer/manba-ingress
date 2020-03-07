package state

// store the state of all components

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/pkg/errors"
)

type collection struct {
	db *memdb.MemDB
}

// ManbaState is an in-memory database representation
// of Manba's configuration.
type ManbaState struct {
	common   collection
	APIs     *APICollection
	Servers  *ServerCollection
	Clusters *ClusterCollection
	Routings *RoutingCollection
}

// NewManbaState creates a new in-memory ManbaState.
func NewManbaState() (*ManbaState, error) {
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			clusterTableName: clusterTableSchema,
			apiTableName:     apiTableSchema,
			routingTableName: routingTableSchema,
			serverTableName:  serverTableSchema,
		},
	}
	memDB, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, errors.Wrap(err, "creating new ServiceCollection")
	}
	var state ManbaState
	state.common = collection{
		db: memDB,
	}
	state.Clusters = (*ClusterCollection)(&state.common)
	state.APIs = (*APICollection)(&state.common)
	state.Servers = (*ServerCollection)(&state.common)
	state.Routings = (*RoutingCollection)(&state.common)
	return &state, nil
}
