package state

import (
	"fmt"
	"reflect"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
)

type pb interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

// Metadata contains additional information for an entity
type Metadata struct {
	meta map[string]interface{}
}

func (m *Metadata) initMeta() {
	if m.meta == nil {
		m.meta = make(map[string]interface{})
	}
}

// AddMeta adds key->obj metadata.
// It will override the old obj in key is already present.
func (m *Metadata) AddMeta(key string, obj interface{}) {
	m.initMeta()
	m.meta[key] = obj
}

// GetMeta returns the obj previously added using AddMeta().
// It returns nil if key is not present.
func (m *Metadata) GetMeta(key string) interface{} {
	m.initMeta()
	return m.meta[key]
}

// Cluster represents a cluster in Manba.
// It adds some helper methods along with Metadata to the original Cluster object.
type Cluster struct {
	metapb.Cluster
	Metadata
}

// Identifier returns cluster name or id
func (c *Cluster) Identifier() string {
	if c.Name != "" {
		return c.Name
	}
	return fmt.Sprintf("%d", c.ID)
}

// Equal returns true if c and c2 are equal.
func (c *Cluster) Equal(c2 *Cluster) bool {
	return reflect.DeepEqual(c, c2)
}

// Routing represents a routing in Manba.
// It adds some helper methods along with Metadata to the original Routing object.
type Routing struct {
	metapb.Routing
	Metadata
}

// Identifier returns routing name or id
func (c *Routing) Identifier() string {
	if c.Name != "" {
		return c.Name
	}
	return fmt.Sprintf("%d", c.ID)
}

// Equal returns true if c and c2 are equal.
func (c *Routing) Equal(c2 *Routing) bool {
	return reflect.DeepEqual(c, c2)
}

// Server represents a server in Manba.
// It adds some helper methods along with Metadata to the original Server object.
type Server struct {
	metapb.Server
	Metadata
}

// Identifier returns server addr or id
func (c *Server) Identifier() string {
	if c.Addr != "" {
		return c.Addr
	}
	return fmt.Sprintf("%d", c.ID)
}

// Equal returns true if c and c2 are equal.
func (c *Server) Equal(c2 *Server) bool {
	return reflect.DeepEqual(c, c2)
}

// API represents a api in Manba.
// It adds some helper methods along with Metadata to the original API object.
type API struct {
	metapb.API
	Metadata
}

// Identifier returns api name or id
func (c *API) Identifier() string {
	if c.Name != "" {
		return c.Name
	}
	return fmt.Sprintf("%d", c.ID)
}

// Equal returns true if c and c2 are equal.
func (c *API) Equal(c2 *API) bool {
	return reflect.DeepEqual(c, c2)
}

// Bind represents a bind in Manba.
// It adds some helper methods along with Metadata to the original API object.
type Bind struct {
	ID string
	metapb.Bind
	Metadata
}

// Identifier returns cluster_id-server_id
func (c *Bind) Identifier() string {
	if c.ID != "" {
		return c.ID
	}
	return fmt.Sprintf("%d-%d", c.ClusterID, c.ServerID)
}

// Equal returns true if c and c2 are equal.
func (c *Bind) Equal(c2 *Bind) bool {
	return reflect.DeepEqual(c, c2)
}

func deepCopyManbaStruct(src, dist pb) error {
	data, err := src.Marshal()
	if err != nil {
		return err
	}
	return dist.Unmarshal(data)
}
