package dump

import (
	manba "github.com/fagongzi/gateway/pkg/client"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/pkg/errors"
)

// Get raw data from manba api-server
func Get(client manba.Client) (*ManbaRawState, error) {
	var res ManbaRawState

	err := client.GetClusterList(func(c *metapb.Cluster) bool {
		res.Clusters = append(res.Clusters, &Cluster{
			Cluster: c,
		})
		return true
	})
	if err != nil {
		return nil, errors.Wrap(err, "list clusters")
	}

	err = client.GetServerList(func(s *metapb.Server) bool {
		res.Servers = append(res.Servers, &Server{
			Server: s,
		})
		return true
	})
	if err != nil {
		return nil, errors.Wrap(err, "list servers")
	}

	for _, c := range res.Clusters {
		ids, err := client.GetBindServers(c.GetID())
		if err != nil {
			return nil, errors.Wrap(err, "list binds")
		}
		for _, id := range ids {
			res.Binds = append(res.Binds, &Bind{
				Bind: &metapb.Bind{
					ClusterID: c.GetID(),
					ServerID:  id,
				},
			})
		}
	}

	err = client.GetAPIList(func(a *metapb.API) bool {
		res.APIs = append(res.APIs, &API{
			API: a,
		})
		return true
	})
	if err != nil {
		return nil, errors.Wrap(err, "list apis")
	}

	err = client.GetRoutingList(func(r *metapb.Routing) bool {
		res.Routings = append(res.Routings, &Routing{
			Routing: r,
		})
		return true
	})
	if err != nil {
		return nil, errors.Wrap(err, "list routings")
	}

	return &res, nil
}
