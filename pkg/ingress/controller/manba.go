/*
Copyright 2015 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"crypto/sha256"
	"encoding/json"
	"reflect"
	"sort"

	"github.com/domgoer/manba-ingress/pkg/ingress/controller/parser"
	"github.com/domgoer/manba-ingress/pkg/manba/diff"
	"github.com/domgoer/manba-ingress/pkg/manba/dump"
	"github.com/domgoer/manba-ingress/pkg/manba/solver"
	"github.com/domgoer/manba-ingress/pkg/manba/state"
	"github.com/domgoer/manba-ingress/pkg/utils"
	"github.com/fagongzi/gateway/pkg/pb"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// OnUpdate is called periodically by syncQueue to keep the configuration in sync.
// returning nil implies the synchronization finished correctly.
// Returning an error means requeue the update.
func (m *ManbaController) OnUpdate(state *parser.ManbaState) error {
	target := m.toStable(state)

	jsonConfig, err := json.Marshal(target)
	if err != nil {
		return errors.Wrap(err, "marshaling Manba declarative configuration to JSON")
	}
	shaSum := sha256.Sum256(jsonConfig)
	if reflect.DeepEqual(m.runningConfigHash, shaSum) {
		glog.Info("no configuration change, skipping sync to Manba")
		return nil
	}

	err = m.onUpdate(state)
	if err == nil {
		glog.Info("successfully synced configuration to Manba")
		m.runningConfigHash = shaSum
	}
	return err
}

func (m *ManbaController) onUpdate(p *parser.ManbaState) error {
	targetRaw := m.toStable(p)
	client := m.cfg.Client

	raw, err := dump.Get(client)
	if err != nil {
		return errors.Wrap(err, "loading configuration from manba")
	}

	currentState, err := state.Get(raw)
	if err != nil {
		return errors.Wrap(err, "get current state")
	}

	err = m.setTargetsIDs(targetRaw, currentState)
	if err != nil {
		return errors.Wrap(err, "set target IDs")
	}

	targetRaw = m.filterInvalidations(targetRaw)

	targetState, err := state.Get(targetRaw)
	if err != nil {
		return errors.Wrap(err, "get target state")
	}

	syncer, err := diff.NewSyncer(currentState, targetState)
	if err != nil {
		return errors.Wrap(err, "new syncer")
	}

	syncer.SilenceWarnings = true
	_, err = solver.Solve(nil, syncer, client, m.cfg.Concurrency)

	return err
}

func (m *ManbaController) toStable(s *parser.ManbaState) *dump.ManbaRawState {
	var ms dump.ManbaRawState
	for _, api := range s.APIs {
		a := api.API

		var proxies []parser.Proxy
		for _, a := range api.Proxies {
			proxies = append(proxies, a)
		}
		ms.APIs = append(ms.APIs, &dump.API{
			API:     &a,
			Proxies: proxies,
		})

	}

	sort.SliceStable(ms.APIs, func(i, j int) bool {
		return ms.APIs[i].Name < ms.APIs[j].Name
	})

	for _, server := range s.Servers {
		svr := server.Server
		ms.Servers = append(ms.Servers, &dump.Server{
			Server: &svr,
		})
	}

	sort.SliceStable(ms.Servers, func(i, j int) bool {
		return ms.Servers[i].Addr < ms.Servers[j].Addr
	})

	for _, routing := range s.Routings {
		r := routing.Routing
		ms.Routings = append(ms.Routings, &dump.Routing{
			Routing: &r,
		})
	}
	sort.SliceStable(ms.Routings, func(i, j int) bool {
		return ms.Routings[i].Name < ms.Routings[j].Name
	})

	for _, cls := range s.Clusters {
		c := cls.Cluster

		ms.Clusters = append(ms.Clusters, &dump.Cluster{
			Cluster: &c,
		})

		for _, svr := range cls.Servers {
			// Add binds
			ms.Binds = append(ms.Binds, &dump.Bind{
				ClusterName: c.GetName(),
				ServerAddr:  svr.GetAddr(),
			})
		}
	}

	sort.SliceStable(ms.Clusters, func(i, j int) bool {
		return ms.Clusters[i].Name < ms.Clusters[j].Name
	})

	sort.SliceStable(ms.Binds, func(i, j int) bool {
		return ms.Binds[i].ClusterName < ms.Binds[j].ClusterName && ms.Binds[i].ServerAddr < ms.Binds[j].ServerAddr
	})

	return &ms
}

// setTargetsIDs gets their id from the existing state in manba and fill it into k8s state
// p: Used to obtain the relationship between various resources
func (m *ManbaController) setTargetsIDs(target *dump.ManbaRawState, current *state.ManbaState) error {
	serverAddrIDsMap := make(map[string]uint64, len(target.Servers))
	clusterNameIDsMap := make(map[string]uint64, len(target.Clusters))

	for _, server := range target.Servers {
		if server.GetID() == 0 {
			s, err := current.Servers.Get(server.GetAddr())

			if err == state.ErrNotFound {
				server.ID = utils.SnowID()
			} else if err != nil {
				return err
			} else {
				server.ID = s.GetID()
			}
		}
		serverAddrIDsMap[server.GetAddr()] = server.GetID()
	}

	for _, cluster := range target.Clusters {
		if cluster.GetID() == 0 {
			c, err := current.Clusters.Get(cluster.GetName())
			if err == state.ErrNotFound {
				cluster.ID = utils.SnowID()
			} else if err != nil {
				return err
			} else {
				cluster.ID = c.GetID()
			}
		}

		clusterNameIDsMap[cluster.GetName()] = cluster.GetID()
	}

	for _, bind := range target.Binds {
		clusterID, ok := clusterNameIDsMap[bind.ClusterName]
		if !ok {
			glog.Warningf("not found cluster <%s> in bind", bind.ClusterName)
			continue
		}

		serverID, ok := serverAddrIDsMap[bind.ServerAddr]
		if !ok {
			glog.Warningf("not found server <%s> in bind", bind.ServerAddr)
			continue
		}

		bind.Bind = &metapb.Bind{
			ClusterID: clusterID,
			ServerID:  serverID,
		}
	}

	for _, api := range target.APIs {
		if api.GetID() == 0 {
			a, err := current.APIs.Get(api.Name)
			if err == state.ErrNotFound {
				api.ID = utils.SnowID()
			} else if err != nil {
				return err
			} else {
				api.ID = a.GetID()
			}
		}
		var nodes []*metapb.DispatchNode
		for _, proxy := range api.Proxies {
			n := proxy.DispatchNode
			n.ClusterID = clusterNameIDsMap[proxy.ClusterName]
			nodes = append(nodes, &n)
		}

		api.Nodes = nodes
	}

	for _, routing := range target.Routings {
		if routing.ID == 0 {
			r, err := current.Routings.Get(routing.Name)
			if err == state.ErrNotFound {
				routing.ID = utils.SnowID()
			} else if err != nil {
				return err
			} else {
				routing.ID = r.GetID()
			}
		}
	}

	return nil
}

func (m *ManbaController) filterInvalidations(raw *dump.ManbaRawState) *dump.ManbaRawState {
	res := new(dump.ManbaRawState)
	validClusters := make(map[uint64]bool, len(raw.Clusters))
	validServers := make(map[uint64]bool, len(raw.Servers))

	for _, cluster := range raw.Clusters {
		if err := pb.ValidateCluster(cluster.Cluster); err != nil {
			glog.Warningf("cluster <%v> is invalid: %v", cluster, err)
			continue
		}
		validClusters[cluster.GetID()] = true
		res.Clusters = append(res.Clusters, cluster)
	}

	for _, server := range raw.Servers {
		if err := pb.ValidateServer(server.Server); err != nil {
			glog.Warningf("server <%v> is invalid: %v", server, err)
			continue
		}
		validServers[server.GetID()] = true
		res.Servers = append(res.Servers, server)
	}

	for _, bind := range raw.Binds {
		if validServers[bind.GetServerID()] && validClusters[bind.GetClusterID()] {
			res.Binds = append(res.Binds, bind)
		} else {
			glog.Warningf("cluster: %d, server: %d", bind.GetClusterID(), bind.GetServerID())
		}
	}

	for _, api := range raw.APIs {
		if err := pb.ValidateAPI(api.API); err != nil {
			glog.Warningf("api <%v> is invalid: %v", api, err)
			continue
		}
		res.APIs = append(res.APIs, api)
	}

	for _, routing := range raw.Routings {
		if err := pb.ValidateRouting(routing.Routing); err != nil {
			glog.Warningf("routing <%v> is invalid: %v", routing, err)
			continue
		}
		res.Routings = append(res.Routings, routing)
	}
	return res
}
