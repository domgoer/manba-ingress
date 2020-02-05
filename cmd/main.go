package main

import (
	"fmt"
	"math/rand"
	"time"

	cache2 "github.com/domgoer/manba-ingress/pkg/cache"
	manbaClient "github.com/fagongzi/manba/pkg/client"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sVersion "k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultQPS   = 1e6
	defaultBurst = 1e6
)

func main() {
	rand.Seed(time.Now().UnixNano())

	cfg, err := parseFlags()
	if err != nil {
		glog.Fatal(err)
	}

	if cfg.SyncPeriod.Seconds() < 10 {
		glog.Fatalf("resync period (%vs) is too low", cfg.SyncPeriod.Seconds())
	}

	// init k8s client
	restCfg, k8sClient, err := createApiserverClient(cfg.APIServerHost, cfg.KubeConfigFilePath)
	if err != nil {
		glog.Fatalf("create k8s client failed, err: %v", err)
	}

	// init manba client
	manbaCli, err := manbaClient.NewClient(cfg.ManbaAPIServerTimeout, cfg.ManbaAPIServer)
	if err != nil {
		glog.Fatalf("create manba client failed, err: %v", err)
	}

	var synced []cache.InformerSynced
	informers := cache2.CreateInformers(k8sClient, cfg.SyncPeriod, cfg.WatchNamespace)
	stopCh := make(chan struct{})
	for _, informer := range informers {
		go informer.Run(stopCh)
		synced = append(synced, informer.HasSynced)
	}
	if !cache.WaitForCacheSync(stopCh, synced...) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
	}

	fmt.Println(cfg)

}

// createApiserverClient creates new Kubernetes Apiserver client. When kubeconfig or apiserverHost param is empty
// the function assumes that it is running inside a Kubernetes cluster and attempts to
// discover the Apiserver. Otherwise, it connects to the Apiserver specified.
//
// apiserverHost param is in the format of protocol://address:port/pathPrefix, e.g.http://localhost:8001.
// kubeConfig location of kubeconfig file
func createApiserverClient(apiserverHost string, kubeConfig string) (*rest.Config, *kubernetes.Clientset, error) {
	cfg, err := clientcmd.BuildConfigFromFlags(apiserverHost, kubeConfig)
	if err != nil {
		return nil, nil, err
	}

	cfg.QPS = defaultQPS
	cfg.Burst = defaultBurst

	glog.Infof("Creating API client for %s", cfg.Host)

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, err
	}

	var v *k8sVersion.Info

	// In some environments is possible the client cannot connect the API server in the first request
	// https://github.com/kubernetes/ingress-nginx/issues/1968
	defaultRetry := wait.Backoff{
		Steps:    10,
		Duration: 1 * time.Second,
		Factor:   1.5,
		Jitter:   0.1,
	}

	var lastErr error
	retries := 0
	glog.V(2).Info("trying to discover Kubernetes version")
	err = wait.ExponentialBackoff(defaultRetry, func() (bool, error) {
		v, err = client.Discovery().ServerVersion()

		if err == nil {
			return true, nil
		}

		lastErr = err
		glog.V(2).Infof("unexpected error discovering Kubernetes version (attempt %v): %v", err, retries)
		retries++
		return false, nil
	})

	// err is not null only if there was a timeout in the exponential backoff (ErrWaitTimeout)
	if err != nil {
		return nil, nil, lastErr
	}

	// this should not happen, warn the user
	if retries > 0 {
		glog.Warningf("it was required to retry %v times before reaching the API server", retries)
	}

	glog.Infof("Running in Kubernetes Cluster version v%v.%v (%v) - git (%v) commit %v - platform %v",
		v.Major, v.Minor, v.GitVersion, v.GitTreeState, v.GitCommit, v.Platform)

	return cfg, client, nil
}
