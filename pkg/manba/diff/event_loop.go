package diff

import (
	"sync/atomic"

	"github.com/domgoer/manba-ingress/pkg/manba/crud"
	"github.com/pkg/errors"
)

// Do is the worker function to sync the diff
// TODO remove crud.Arg
type Do func(a crud.Arg) (crud.Arg, error)

func (sc *Syncer) eventLoop(d Do, a int) error {
	for event := range sc.eventChan {
		err := sc.handleEvent(d, event, a)
		sc.eventCompleted(event)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sc *Syncer) handleEvent(d Do, event crud.Event, a int) error {
	res, err := d(event)
	if err != nil {
		return errors.Wrapf(err, "while processing event")
	}
	if res == nil {
		return errors.New("result of event is nil")
	}
	_, err = sc.postProcess.Do(event.Kind, event.Op, res)
	if err != nil {
		return errors.Wrap(err, "while post processing event")
	}
	return nil
}

func (sc *Syncer) eventCompleted(e crud.Event) {
	atomic.AddInt32(&sc.InFlightOps, -1)
}
