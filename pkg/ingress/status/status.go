/*
Copyright 2015 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package status

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/leaderelection"

	"github.com/pkg/errors"

	"gopkg.in/go-playground/pool.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"

	"github.com/domgoer/manba-ingress/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/domgoer/manba-ingress/pkg/ingress/k8s"
	"github.com/domgoer/manba-ingress/pkg/ingress/task"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"

	networkingv1beta1 "k8s.io/api/networking/v1beta1"
)

const (
	updateInterval = 60 * time.Second
)

// Syncer ...
type Syncer interface {
	Run(chan struct{})

	Shutdown(bool)

	Callbacks() leaderelection.LeaderCallbacks
}

type ingressLister interface {
	// ListIngresses returns the list of Ingresses
	ListIngresses() []*networkingv1beta1.Ingress
}

// Config ...
type Config struct {
	Client kubernetes.Interface

	PublishStatusAddress string
	PublishService       string

	ElectionID string

	UpdateStatusOnShutdown bool
	Concurrency            int

	IngressLister ingressLister

	OnStartedLeading func()
}

// statusSync keeps the status IP in each Ingress rule updated executing a periodic check
// in all the defined rules. To simplify the process leader election is used so the update
// is executed only in one node (Ingress controllers can be scaled to more than one)
// If the controller is running with the flag --publish-service (with a valid service)
// the IP address behind the service is used, if it is running with the flag
// --publish-status-address, the address specified in the flag is used, if neither of the
// two flags are set, the source is the IP/s of the node/s
type statusSync struct {
	Config

	// pod contains runtime information about this pod
	pod *k8s.PodInfo

	// workqueue used to keep in sync the status IP/s
	// in the Ingress rules
	syncQueue *task.Queue

	callbacks leaderelection.LeaderCallbacks
}

func (s *statusSync) Shutdown(isLeader bool) {
	go s.syncQueue.Shutdown()
	// remove IP from Ingress
	if !isLeader {
		return
	}

	// on shutdown we remove information about the leader election to
	// avoid up to 30 seconds of delay in start the synchronization process
	c, err := s.Client.CoreV1().ConfigMaps(s.pod.Namespace).Get(s.ElectionID, metav1.GetOptions{})
	if err == nil {
		c.Annotations = map[string]string{}
		s.Client.CoreV1().ConfigMaps(s.pod.Namespace).Update(c)
	}

	if !s.UpdateStatusOnShutdown {
		glog.Warningf("skipping update of status of Ingress rules")
		return
	}

	glog.Infof("updating status of Ingress rules (remove)")

	addrs, err := s.runningAddresses()
	if err != nil {
		glog.Errorf("error obtaining running IPs: %v", addrs)
		return
	}

	if len(addrs) > 1 {
		// leave the job to the next leader
		glog.Infof("leaving status update for next leader (%v)", len(addrs))
		return
	}

	if s.isRunningMultiplePods() {
		glog.V(2).Infof("skipping Ingress status update (multiple pods running - another one will be elected as master)")
		return
	}

	glog.Infof("removing address from ingress status (%v)", addrs)
	s.updateStatus([]corev1.LoadBalancerIngress{})
}

// updateStatus changes the status information of Ingress rules
func (s *statusSync) updateStatus(newIngressPoint []corev1.LoadBalancerIngress) {
	ings := s.IngressLister.ListIngresses()

	p := pool.NewLimited(10)
	defer p.Close()

	batch := p.Batch()
	sort.SliceStable(newIngressPoint, lessLoadBalancerIngress(newIngressPoint))

	for _, ing := range ings {
		curIPs := ing.Status.LoadBalancer.Ingress
		sort.SliceStable(curIPs, lessLoadBalancerIngress(curIPs))
		if ingressSliceEqual(curIPs, newIngressPoint) {
			klog.V(3).Infof("skipping update of Ingress %v/%v (no change)", ing.Namespace, ing.Name)
			continue
		}

		batch.Queue(s.runUpdate(ing, newIngressPoint, s.Client))
	}

	batch.QueueComplete()
	batch.WaitAll()
}

// Start starts the loop to keep the status in sync
func (s *statusSync) Run(stopCh chan struct{}) {
}

func (s *statusSync) sync(key interface{}) error {
	if s.syncQueue.IsShuttingDown() {
		glog.V(2).Infof("skipping Ingress status update (shutting down in progress)")
		return nil
	}

	addrs, err := s.runningAddresses()
	if err != nil {
		return err
	}
	s.updateStatus(sliceToStatus(addrs))

	return nil
}

func (s *statusSync) Callbacks() leaderelection.LeaderCallbacks {
	return s.callbacks
}

// runningAddresses returns a list of IP addresses and/or FQDN where the
// ingress controller is currently running
func (s *statusSync) runningAddresses() ([]string, error) {
	addrs := []string{}
	if s.PublishStatusAddress != "" {
		addrs = append(addrs, s.PublishStatusAddress)
		return addrs, nil
	}
	ns, name, _ := utils.ParseNameNS(s.PublishStatusAddress)
	svc, err := s.Client.CoreV1().Services(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	switch svc.Spec.Type {
	case corev1.ServiceTypeLoadBalancer:
		for _, ip := range svc.Status.LoadBalancer.Ingress {
			if ip.IP == "" {
				addrs = append(addrs, ip.Hostname)
			} else {
				addrs = append(addrs, ip.IP)
			}
		}

		addrs = append(addrs, svc.Spec.ExternalIPs...)
		return addrs, nil
	default:
		// get information about all the pods running the ingress controller
		pods, err := s.Client.CoreV1().Pods(s.pod.Namespace).List(metav1.ListOptions{
			LabelSelector: labels.SelectorFromSet(s.pod.Labels).String(),
		})
		if err != nil {
			return nil, err
		}

		for _, pod := range pods.Items {
			// only Running pods are valid
			if pod.Status.Phase != corev1.PodRunning {
				continue
			}

			name := utils.GetNodeIPOrName(s.Client, pod.Spec.NodeName)
			if !inSlice(name, addrs) {
				addrs = append(addrs, name)
			}
		}

		return addrs, nil
	}
}

func (s statusSync) keyfunc(input interface{}) (interface{}, error) {
	return input, nil
}

func (s *statusSync) isRunningMultiplePods() bool {
	pods, err := s.Client.CoreV1().Pods(s.pod.Namespace).List(metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(s.pod.Labels).String(),
	})
	if err != nil {
		return false
	}

	return len(pods.Items) > 1
}

func (s *statusSync) runUpdate(ing *networkingv1beta1.Ingress, status []corev1.LoadBalancerIngress,
	client kubernetes.Interface) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		if wu.IsCancelled() {
			return nil, nil
		}

		sort.SliceStable(status, lessLoadBalancerIngress(status))

		curIPs := ing.Status.LoadBalancer.Ingress
		sort.SliceStable(curIPs, lessLoadBalancerIngress(curIPs))

		if ingressSliceEqual(status, curIPs) {
			glog.V(3).Infof("skipping update of Ingress %v/%v (no change)", ing.Namespace, ing.Name)
			return true, nil
		}

		ingClient := client.NetworkingV1beta1().Ingresses(ing.Namespace)

		currIng, err := ingClient.Get(ing.Name, metav1.GetOptions{})
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("unexpected error searching Ingress %v/%v", ing.Namespace, ing.Name))
		}

		glog.Infof("updating Ingress %v/%v status to %v", currIng.Namespace, currIng.Name, status)
		currIng.Status.LoadBalancer.Ingress = status
		_, err = ingClient.UpdateStatus(currIng)
		if err != nil {
			glog.Warningf("error updating ingress rule: %v", err)
		}

		return true, nil
	}
}

// NewStatusSyncer returns a new Syncer instance
func NewStatusSyncer(podInfo *k8s.PodInfo, config Config) Syncer {
	st := statusSync{
		pod: podInfo,

		Config: config,
	}
	st.syncQueue = task.NewCustomTaskQueue(st.sync, st.keyfunc)

	st.callbacks = leaderelection.LeaderCallbacks{
		OnStartedLeading: func(ctx context.Context) {
			glog.V(2).Infof("I am the new status update leader")
			if st.Config.OnStartedLeading != nil {
				st.Config.OnStartedLeading()
			}
			go st.syncQueue.Run(time.Second, ctx.Done())
			wait.PollUntil(updateInterval, func() (bool, error) {
				// send a dummy object to the queue to force a sync
				st.syncQueue.Enqueue("sync status")
				return false, nil
			}, ctx.Done())
		},
		OnStoppedLeading: func() {
			glog.V(2).Infof("I am not status update leader anymore")
		},
		OnNewLeader: func(identity string) {
			glog.Infof("new leader elected: %v", identity)
		},
	}

	return &st
}

func inSlice(val string, list []string) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}

	return false
}

// sliceToStatus converts a slice of IP and/or hostnames to LoadBalancerIngress
func sliceToStatus(endpoints []string) []corev1.LoadBalancerIngress {
	lbi := []corev1.LoadBalancerIngress{}
	for _, ep := range endpoints {
		if net.ParseIP(ep) == nil {
			lbi = append(lbi, corev1.LoadBalancerIngress{Hostname: ep})
		} else {
			lbi = append(lbi, corev1.LoadBalancerIngress{IP: ep})
		}
	}

	sort.SliceStable(lbi, func(a, b int) bool {
		return lbi[a].IP < lbi[b].IP
	})

	return lbi
}

func lessLoadBalancerIngress(addrs []corev1.LoadBalancerIngress) func(int, int) bool {
	return func(a, b int) bool {
		switch strings.Compare(addrs[a].Hostname, addrs[b].Hostname) {
		case -1:
			return true
		case 1:
			return false
		}
		return addrs[a].IP < addrs[b].IP
	}
}

func ingressSliceEqual(lhs, rhs []corev1.LoadBalancerIngress) bool {
	if len(lhs) != len(rhs) {
		return false
	}

	for i := range lhs {
		if lhs[i].IP != rhs[i].IP {
			return false
		}
		if lhs[i].Hostname != rhs[i].Hostname {
			return false
		}
	}
	return true
}
