package diff

import (
	"github.com/domgoer/manba-ingress/pkg/manba/crud"
	"github.com/domgoer/manba-ingress/pkg/manba/state"
	"github.com/pkg/errors"
)

const (
	bindKind = "bind"
)

func (sc *Syncer) deleteBinds() error {
	binds, err := sc.currentState.Binds.GetAll()
	if err != nil {
		errors.Wrap(err, "fetching binds from state")
	}

	for _, bind := range binds {
		event, err := sc.deleteBind(bind)
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

func (sc *Syncer) deleteBind(bind *state.Bind) (*crud.Event, error) {
	_, err := sc.targetState.Binds.Get(bind.Identifier())
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: bindKind,
			Obj:  bind,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up bind '%s'", bind.Identifier())
	}
	return nil, nil
}

func (sc *Syncer) createUpdateBinds() error {
	binds, err := sc.targetState.Binds.GetAll()
	if err != nil {
		return errors.Wrap(err, "fetching binds from state")
	}

	for _, bind := range binds {
		event, err := sc.createUpdateBind(bind)
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

func (sc *Syncer) createUpdateBind(bind *state.Bind) (*crud.Event, error) {
	manbaBind := bind.DeepCopy().Bind
	newBind := &state.Bind{Bind: manbaBind}

	current, err := sc.currentState.Binds.Get(newBind.Identifier())
	if err == state.ErrNotFound {
		// bind not present, create it
		return &crud.Event{
			Op:   crud.Create,
			Kind: bindKind,
			Obj:  newBind,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up bind '%s'", newBind.Identifier())
	}

	// found bind, check equal

	if !state.CompareBind(current, newBind) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   bindKind,
			Obj:    newBind,
			OldObj: current,
		}, nil
	}
	return nil, nil
}
