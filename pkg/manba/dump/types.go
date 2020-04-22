package dump

import (
	"fmt"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
)

// ManbaRawState contains all of Manba data
type ManbaRawState struct {
	APIs     []*API
	Servers  []*Server
	Clusters []*Cluster
	Binds    []*Bind
	Routings []*Routing
}

// Bind ...
type Bind struct {
	ClusterName string
	ServerAddr  string
	*metapb.Bind
}

// Server ...
type Server struct {
	*metapb.Server
}

// Cluster ...
type Cluster struct {
	*metapb.Cluster
}

// Routing ...
type Routing struct {
	*metapb.Routing
}

// API ...
type API struct {
	*metapb.API

	Proxies []Proxy
}

// Proxy ...
type Proxy struct {
	*metapb.DispatchNode

	ServiceNamespace string
	ServiceName      string
	ServiceSubSet    string
	ServicePort      string
}

// GetClusterName returns cluster name of dispatch node
func (p *Proxy) GetClusterName() string {
	return fmt.Sprintf("%s.%s.%s.%d.svc", p.ServiceNamespace, p.ServiceName, p.ServiceSubSet, p.ServicePort)
}
