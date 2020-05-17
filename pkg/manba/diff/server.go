package diff

import (
	"github.com/domgoer/manba-ingress/pkg/manba/crud"
	"github.com/domgoer/manba-ingress/pkg/manba/state"
	"github.com/pkg/errors"
)

const (
	serverKind = "server"
)

func (sc *Syncer) deleteServers() error {
	servers, err := sc.currentState.Servers.GetAll()
	if err != nil {
		return errors.Wrap(err, "fetching servers from state")
	}

	for _, server := range servers {
		event, err := sc.deleteServer(server)
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

func (sc *Syncer) deleteServer(server *state.Server) (*crud.Event, error) {
	_, err := sc.targetState.Servers.Get(server.Identifier())
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: serverKind,
			Obj:  server,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up server '%s'", server.Identifier())
	}
	return nil, nil
}

func (sc *Syncer) createUpdateServers() error {
	servers, err := sc.targetState.Servers.GetAll()
	if err != nil {
		return errors.Wrap(err, "fetching servers from state")
	}

	for _, server := range servers {
		event, err := sc.createUpdateServer(server)
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

func (sc *Syncer) createUpdateServer(server *state.Server) (*crud.Event, error) {
	newServer := server.DeepCopy()

	current, err := sc.currentState.Servers.Get(newServer.Identifier())
	if err == state.ErrNotFound {
		// server not present, create it
		return &crud.Event{
			Op:   crud.Create,
			Kind: serverKind,
			Obj:  newServer,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up server '%s'", newServer.Identifier())
	}

	// found server, check equal

	if !state.CompareServer(current, newServer) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   serverKind,
			Obj:    newServer,
			OldObj: current,
		}, nil
	}
	return nil, nil
}
