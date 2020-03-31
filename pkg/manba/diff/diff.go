package diff

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/domgoer/manba-ingress/pkg/manba/crud"
	"github.com/domgoer/manba-ingress/pkg/manba/state"
)

var (
	errEnqueueFailed = errors.New("failed to queue event")
)

// TODO get rid of the syncer struct and simply have a func for it

// Syncer takes in a current and target state of Kong,
// diffs them, generating a Graph to get Kong from current
// to target state.
type Syncer struct {
	currentState *state.ManbaState
	targetState  *state.ManbaState
	postProcess  crud.Registry

	eventChan chan crud.Event
	errChan   chan error
	stopChan  chan struct{}

	InFlightOps int32

	SilenceWarnings bool

	once sync.Once
}

// NewSyncer constructs a Syncer.
func NewSyncer(current, target *state.ManbaState) (*Syncer, error) {
	s := &Syncer{}
	s.currentState, s.targetState = current, target

	s.postProcess = crud.NewPostProcess(current)
	return s, nil
}

func (sc *Syncer) wait() {
	for atomic.LoadInt32(&sc.InFlightOps) != 0 {
		// TODO hack?
		time.Sleep(5 * time.Millisecond)
	}
}

// Run starts a diff and invokes d for every diff.
func (sc *Syncer) Run(done <-chan struct{}, parallelism int, d Do) []error {
	if parallelism < 1 {
		return append([]error{}, errors.New("parallelism can not be negative"))
	}

	var wg sync.WaitGroup

	sc.eventChan = make(chan crud.Event, 10)
	sc.stopChan = make(chan struct{})
	sc.errChan = make(chan error)

	// run rabbit run
	// start the consumers
	wg.Add(parallelism)
	for i := 0; i < parallelism; i++ {
		go func(a int) {
			err := sc.eventLoop(d, a)
			if err != nil {
				sc.errChan <- err
			}
			wg.Done()
		}(i)
	}

	// start the producer
	wg.Add(1)
	go func() {
		err := sc.diff()
		if err != nil {
			sc.errChan <- err
		}
		close(sc.eventChan)
		wg.Done()
	}()

	// close the error chan once all done
	go func() {
		wg.Wait()
		close(sc.errChan)
	}()

	var errs []error
	select {
	case <-done:
	case err, ok := <-sc.errChan:
		if ok && err != nil {
			if err != errEnqueueFailed {
				errs = append(errs, err)
			}
		}
	}

	// stop the producer
	close(sc.stopChan)

	// collect errors
	for err := range sc.errChan {
		if err != errEnqueueFailed {

			errs = append(errs, err)
		}
	}

	return errs
}

func (sc *Syncer) diff() error {
	err := sc.createUpdate()
	if err != nil {
		return err
	}

	err = sc.delete()
	if err != nil {
		return err
	}
	return nil
}

func (sc *Syncer) queueEvent(e crud.Event) error {
	atomic.AddInt32(&sc.InFlightOps, 1)
	select {
	case sc.eventChan <- e:
		return nil
	case <-sc.stopChan:
		return errEnqueueFailed
	}
}

func (sc *Syncer) createUpdate() error {
	var err error
	createUpdateList := []func() error{
		sc.createUpdateServers,
		sc.createUpdateClusters,
		sc.createUpdateBinds,
		sc.createUpdateAPIs,
		sc.createUpdateRoutings,
	}

	for _, createUpdate := range createUpdateList {
		err = createUpdate()
		if err != nil {
			return err
		}
		sc.wait()
	}
	return nil
}

func (sc *Syncer) delete() error {
	var err error
	deleteList := []func() error{
		sc.deleteRoutings,
		sc.deleteAPIs,
		sc.deleteBinds,
		sc.deleteClusters,
		sc.deleteServers,
	}

	for _, delete := range deleteList {
		err = delete()
		if err != nil {
			return err
		}
		sc.wait()
	}
	return nil
}
