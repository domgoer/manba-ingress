package diff

import (
	"github.com/domgoer/manba-ingress/pkg/manba/crud"
	"github.com/domgoer/manba-ingress/pkg/manba/state"
	"github.com/pkg/errors"
)

const (
	routingKind = "routing"
)

func (sc *Syncer) deleteRoutings() error {
	routings, err := sc.currentState.Routings.GetAll()
	if err != nil {
		errors.Wrap(err, "fetching routings from state")
	}

	for _, routing := range routings {
		event, err := sc.deleteRouting(routing)
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

func (sc *Syncer) deleteRouting(routing *state.Routing) (*crud.Event, error) {
	_, err := sc.targetState.Routings.Get(routing.Identifier())
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: routingKind,
			Obj:  routing,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up routing '%s'", routing.Identifier())
	}
	return nil, nil
}

func (sc *Syncer) createUpdateRoutings() error {
	routings, err := sc.targetState.Routings.GetAll()
	if err != nil {
		return errors.Wrap(err, "fetching routings from state")
	}

	for _, routing := range routings {
		event, err := sc.createUpdateRouting(routing)
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

func (sc *Syncer) createUpdateRouting(routing *state.Routing) (*crud.Event, error) {
	manbaRouting := routing.DeepCopy().Routing
	newRouting := &state.Routing{Routing: manbaRouting}

	current, err := sc.currentState.Routings.Get(newRouting.Identifier())
	if err == state.ErrNotFound {
		// routing not present, create it
		return &crud.Event{
			Op:   crud.Create,
			Kind: routingKind,
			Obj:  newRouting,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up routing '%s'", newRouting.Identifier())
	}

	// found routing, check equal

	if !state.CompareRouting(current, newRouting) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   routingKind,
			Obj:    newRouting,
			OldObj: current,
		}, nil
	}
	return nil, nil
}
