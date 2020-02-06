package election

import (
	"context"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
)

const (
	lockType = "configmaps"

	leaseDuration = time.Second * 30
	retryPeriod   = time.Second * 2
	renewDeadline = time.Second * 10
)

type Config struct {
	ElectionID        string
	ResourceName      string
	ResourceNamespace string

	Callbacks leaderelection.LeaderCallbacks
}

type Elector interface {
	IsLeader() bool
	Run(context.Context)
}

// NewElection returns leaderelection.LeaderElector to start election, should use leaderelection.LeaderElector.Run(ctx)
func NewElection(config Config, client kubernetes.Interface) (*leaderelection.LeaderElector, error) {
	lec, err := getLeaderElectionConfig(config, client)
	if err != nil {
		return nil, err
	}
	return leaderelection.NewLeaderElector(lec)
}

func getLeaderElectionConfig(config Config, client kubernetes.Interface) (lec leaderelection.LeaderElectionConfig, err error) {
	leaderElectionBroadcaster := record.NewBroadcaster()
	host, err := os.Hostname()
	if err != nil {
		return
	}
	recorder := leaderElectionBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: config.ResourceName, Host: host})

	rl, err := resourcelock.New(lockType,
		config.ResourceNamespace,
		config.ResourceName,
		client.CoreV1(),
		client.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity:      config.ElectionID,
			EventRecorder: recorder,
		})
	if err != nil {
		return
	}

	lec = leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: leaseDuration,
		RenewDeadline: renewDeadline,
		RetryPeriod:   retryPeriod,
		Callbacks:     config.Callbacks,
		WatchDog:      leaderelection.NewLeaderHealthzAdaptor(time.Second * 20),
		Name:          config.ResourceName,
	}
	return
}
