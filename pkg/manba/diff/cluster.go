package diff

import (
	"github.com/domgoer/manba-ingress/pkg/manba/crud"
	"github.com/domgoer/manba-ingress/pkg/manba/state"
	"github.com/pkg/errors"
)

const (
	clusterKind = "cluster"
)

func (sc *Syncer) deleteClusters() error {
	clusters, err := sc.currentState.Clusters.GetAll()
	if err != nil {
		errors.Wrap(err, "fetching clusters from state")
	}

	for _, cluster := range clusters {
		event, err := sc.deleteCluster(cluster)
		if err != nil {
			return err
		}
		if event != nil {
			err = sc.queueEvent(*event)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sc *Syncer) deleteCluster(cluster *state.Cluster) (*crud.Event, error) {
	_, err := sc.targetState.Clusters.Get(cluster.Identifier())
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: clusterKind,
			Obj:  cluster,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up cluster '%s'", cluster.Identifier())
	}
	return nil, nil
}

func (sc *Syncer) createUpdateClusters() error {
	clusters, err := sc.targetState.Clusters.GetAll()
	if err != nil {
		return errors.Wrap(err, "fetching clusters from state")
	}

	for _, cluster := range clusters {
		event, err := sc.createUpdateCluster(cluster)
		if err != nil {
			return err
		}
		if event != nil {
			err = sc.queueEvent(*event)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sc *Syncer) createUpdateCluster(cluster *state.Cluster) (*crud.Event, error) {
	manbaCluster := state.DeepCopyManbaCluster(cluster)
	newCluster := &state.Cluster{Cluster: *manbaCluster}

	current, err := sc.currentState.Clusters.Get(newCluster.Identifier())
	if err == state.ErrNotFound {
		// cluster not present, create it
		return &crud.Event{
			Op:   crud.Create,
			Kind: clusterKind,
			Obj:  newCluster,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up cluster '%s'", newCluster.Identifier())
	}

	// found cluster, check equal

	if !state.CompareCluster(current, newCluster) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   clusterKind,
			Obj:    newCluster,
			OldObj: current,
		}, nil
	}
	return nil, nil
}
