package state

import (
	"fmt"
	"reflect"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
)

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
