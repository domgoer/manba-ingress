package dump

import "github.com/fagongzi/gateway/pkg/pb/metapb"

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
	APIName     string
	ClusterName string
}

// Proxy ...
type Proxy struct {
	APIName     string
	ClusterName string
	*metapb.DispatchNode
}

// API ...
type API struct {
	*metapb.API
	Proxies []Proxy
}
