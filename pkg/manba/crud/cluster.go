package crud

import (
	"github.com/domgoer/manba-ingress/pkg/manba/state"
	manba "github.com/fagongzi/gateway/pkg/client"
)

// clusterPostAction crud cluster in mem-db
type clusterPostAction struct {
	currentState *state.ManbaState
}

func (c *clusterPostAction) Create(arg Arg) (Arg, error) {
	return nil, c.currentState.Clusters.Add(*arg.(*state.Cluster))
}

func (c *clusterPostAction) Delete(arg Arg) (Arg, error) {
	return nil, c.currentState.Clusters.Delete(arg.(*state.Cluster).Identifier())
}

func (c *clusterPostAction) Update(arg Arg) (Arg, error) {
	return nil, c.currentState.Clusters.Update(*arg.(*state.Cluster))
}

// clusterRawAction crud cluster in manba
type clusterRawAction struct {
	client manba.Client
}

func clusterFromObj(obj interface{}) *state.Cluster {
	cluster, ok := obj.(*state.Cluster)
	if !ok {
		panic("unexpected type, expected *state.Cluster")
	}
	return cluster
}

func (c *clusterRawAction) Create(arg Arg) (Arg, error) {
	event := eventFromArg(arg)
	cluster := clusterFromObj(event.Obj)
	cb := c.client.NewClusterBuilder()
	id, err := cb.Use(cluster.Cluster).Commit()
	if err != nil {
		return nil, err
	}
	cluster.ID = id
	return &state.Cluster{Cluster: cluster.Cluster}, nil

}

func (c *clusterRawAction) Delete(arg Arg) (Arg, error) {
	event := eventFromArg(arg)
	cluster := clusterFromObj(event.Obj)
	err := c.client.RemoveCluster(cluster.ID)
	return cluster, err
}

func (c *clusterRawAction) Update(arg Arg) (Arg, error) {
	event := eventFromArg(arg)
	cluster := clusterFromObj(event.Obj)
	cb := c.client.NewClusterBuilder()
	_, err := cb.Use(cluster.Cluster).Commit()
	return cluster, err
}
