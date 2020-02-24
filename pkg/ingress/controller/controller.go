package controller

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/domgoer/manba-ingress/pkg/ingress/k8s"
	"github.com/domgoer/manba-ingress/pkg/ingress/status"

	"github.com/domgoer/manba-ingress/pkg/ingress/election"
	"github.com/domgoer/manba-ingress/pkg/ingress/store"
	"github.com/domgoer/manba-ingress/pkg/ingress/task"
	"github.com/eapache/channels"
	manbaClient "github.com/fagongzi/gateway/pkg/client"
	"github.com/golang/glog"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/flowcontrol"
)

// Manba Represents a Manba client and connection information
type Manba struct {
	Client manbaClient.Client
}

// Config contains all the settings required by an Ingress controller
type Config struct {
	Manba

	ElectionID string

	KubeClient   kubernetes.Interface
	IngressClass string

	ResyncPeriod  time.Duration
	SyncRateLimit float32

	Namespace string

	UpdateStatus         bool
	PublishService       string
	PublishStatusAddress string
}

type ManbaController struct {
	cfg     Config
	elector election.Elector
	store   store.Store

	syncQueue       *task.Queue
	syncRateLimiter flowcontrol.RateLimiter

	stopCh   chan struct{}
	updateCh *channels.RingChannel

	stopLock       sync.Mutex
	isShuttingDown bool

	syncStatus status.Syncer
}

func NewManbaController(cfg Config, store store.Store) (*ManbaController, error) {
	m := &ManbaController{
		cfg:             cfg,
		store:           store,
		updateCh:        channels.NewRingChannel(1024),
		syncRateLimiter: flowcontrol.NewTokenBucketRateLimiter(cfg.SyncRateLimit, 1),
		stopCh:          make(chan struct{}),
	}
	m.syncQueue = task.NewTaskQueue(m.syncIngress)

	pod, err := k8s.GetPodDetails(cfg.KubeClient)
	if err != nil {
		glog.Fatalf("unexpected error obtaining pod information: %v", err)
	}

	// init leader election
	resourceName := fmt.Sprintf("%s-ingress-controller", cfg.ElectionID)

	ec := election.Config{
		ResourceName:      resourceName,
		ResourceNamespace: cfg.Namespace,
		ElectionID:        cfg.ElectionID,
	}

	if cfg.UpdateStatus {
		m.syncStatus = status.NewStatusSyncer(pod, status.Config{
			Client:               cfg.KubeClient,
			PublishStatusAddress: cfg.PublishStatusAddress,
			PublishService:       cfg.PublishService,
			ElectionID:           cfg.ElectionID,
			// TODO: update status on shutdown
			UpdateStatusOnShutdown: false,
			IngressLister:          store,
			OnStartedLeading: func() {
				// force a sync
				m.syncQueue.Enqueue(&networkingv1beta1.Ingress{})
			},
		})
		ec.Callbacks = m.syncStatus.Callbacks()
	} else {
		glog.Warning("Update of ingress status is disabled (flag --update-status=false was specified)")
	}

	elector, err := election.NewElection(ec, cfg.KubeClient)
	if err != nil {
		return nil, err
	}

	m.elector = elector

	return m, nil
}

// sync collects all the pieces required to assemble the configuration file and
// then sends the content to the backend (OnUpdate) receiving the populated
// template as response reloading the backend if is required.
func (m *ManbaController) syncIngress(interface{}) error {
	m.syncRateLimiter.Accept()

	if m.syncQueue.IsShuttingDown() {
		return nil
	}

	// Sort ingress rules using the ResourceVersion field
	ings := m.store.ListIngresses()
	sort.SliceStable(ings, func(i, j int) bool {
		ir := ings[i].ResourceVersion
		jr := ings[j].ResourceVersion
		return ir < jr
	})

	glog.V(2).Infof("syncing Ingress configuration...")

	return nil
}

func (m *ManbaController) Start() {
	glog.Infof("starting Ingress controller")

	go m.elector.Run(context.Background())

	if m.syncStatus != nil {
		m.syncStatus.Run(m.stopCh)
	}

	go m.syncQueue.Run(time.Second, m.stopCh)
	// force initial sync
	m.syncQueue.Enqueue(&networkingv1beta1.Ingress{})

	for {
		select {
		case event := <-m.updateCh.Out():
			if m.isShuttingDown {
				break
			}
			if evt, ok := event.(Event); ok {
				glog.V(3).Infof("Event %v received - object %v", evt.Type, evt.Obj)
				m.syncQueue.Enqueue(evt.Obj)
			} else {
				glog.Warningf("unexpected event type received %T", event)
			}

		case <-m.stopCh:
			break
		}
	}
}

// Stop gracefully stops the controller
func (m *ManbaController) Stop() error {
	m.isShuttingDown = true
	m.stopLock.Lock()
	defer m.stopLock.Unlock()

	if m.syncQueue.IsShuttingDown() {
		return fmt.Errorf("shutdown already in progress")
	}

	glog.Info("Shutting down controller queues")
	close(m.stopCh)
	go m.syncQueue.Shutdown()

	if m.syncStatus != nil {
		m.syncStatus.Shutdown(m.elector.IsLeader())
	}
	return nil
}
