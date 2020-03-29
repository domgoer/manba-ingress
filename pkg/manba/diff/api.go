package diff

import (
	"github.com/domgoer/manba-ingress/pkg/manba/crud"
	"github.com/domgoer/manba-ingress/pkg/manba/state"
	"github.com/pkg/errors"
)

const (
	apiKind = "api"
)

func (sc *Syncer) deleteAPIs() error {
	apis, err := sc.currentState.APIs.GetAll()
	if err != nil {
		errors.Wrap(err, "fetching apis from state")
	}

	for _, api := range apis {
		event, err := sc.deleteAPI(api)
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

func (sc *Syncer) deleteAPI(api *state.API) (*crud.Event, error) {
	_, err := sc.targetState.APIs.Get(api.Identifier())
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: apiKind,
			Obj:  api,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up api '%s'", api.Identifier())
	}
	return nil, nil
}

func (sc *Syncer) createUpdateAPIs() error {
	apis, err := sc.targetState.APIs.GetAll()
	if err != nil {
		return errors.Wrap(err, "fetching apis from state")
	}

	for _, api := range apis {
		event, err := sc.createUpdateAPI(api)
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

func (sc *Syncer) createUpdateAPI(api *state.API) (*crud.Event, error) {
	manbaAPI := state.DeepCopyManbaAPI(api)
	newAPI := &state.API{API: *manbaAPI}

	current, err := sc.currentState.APIs.Get(newAPI.Identifier())
	if err == state.ErrNotFound {
		// api not present, create it
		return &crud.Event{
			Op:   crud.Create,
			Kind: apiKind,
			Obj:  newAPI,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up api '%s'", newAPI.Identifier())
	}

	// found api, check equal

	if !state.CompareAPI(current, newAPI) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   apiKind,
			Obj:    newAPI,
			OldObj: current,
		}, nil
	}
	return nil, nil
}
