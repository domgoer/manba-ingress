package dump

import "github.com/fagongzi/gateway/pkg/pb/metapb"

// ManbaRawState contains all of Manba data
type ManbaRawState struct {
	APIs     []*metapb.API
	Servers  []*metapb.Server
	Clusters []*metapb.Cluster
	Binds    []*metapb.Bind
	Routings []*metapb.Routing
}

