package state

// store the state of all components

import (
	"github.com/domgoer/manba-ingress/pkg/manba/dump"
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
	Binds    *BindCollection
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

// Get builds a ManbaState from a raw representation of Manba.
func Get(raw *dump.ManbaRawState) (*ManbaState, error) {
	state, err := NewManbaState()
	if err != nil {
		return nil, errors.Wrap(err, "new manba state")
	}

	for _, a := range raw.APIs {
		err := state.APIs.Add(API{API: *a})
		if err != nil {
			return nil, errors.Wrap(err, "add api to state")
		}
	}

	for _, r := range raw.Routings {
		err := state.Routings.Add(Routing{Routing: *r})
		if err != nil {
			return nil, errors.Wrap(err, "add routing to state")
		}
	}

	for _, c := range raw.Clusters {
		err := state.Clusters.Add(Cluster{Cluster: *c})
		if err != nil {
			return nil, errors.Wrap(err, "add cluster to state")
		}
	}

	for _, s := range raw.Servers {
		err := state.Servers.Add(Server{Server: *s})
		if err != nil {
			return nil, errors.Wrap(err, "add server to state")
		}
	}

	for _, b := range raw.Binds {
		err := state.Binds.Add(Bind{Bind: *b})
		if err != nil {
			return nil, errors.Wrap(err, "add bind to state")
		}
	}
	return state, nil
}