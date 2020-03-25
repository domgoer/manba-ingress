package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/domgoer/manba-ingress/pkg/utils"
	"github.com/eapache/channels"

	_ "github.com/domgoer/manba-ingress/pkg/manba/state"

	"github.com/domgoer/manba-ingress/pkg/ingress/annotations"
	"github.com/domgoer/manba-ingress/pkg/ingress/store"

	cache2 "github.com/domgoer/manba-ingress/pkg/cache"
	"github.com/domgoer/manba-ingress/pkg/ingress/controller"
	manbaClient "github.com/fagongzi/gateway/pkg/client"
	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sVersion "k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	_ "github.com/domgoer/manba-ingress/pkg/ingress/controller/parser"
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

	// validate flags
	if cfg.SyncPeriod.Seconds() < 10 {
		glog.Fatalf("resync period (%vs) is too low ", cfg.SyncPeriod.Seconds())
	}

	if cfg.ManbaConcurrency < 1 {
		glog.Fatalf("manba-admin-concurrency (%v) cannot be less than 1", cfg.ManbaConcurrency)
	}

	if cfg.PublishService == "" && cfg.PublishStatusAddress == "" {
		glog.Fatal("either --publish-service or --publish-status-address ",
			"must be specified")
	}

	// init k8s client
	restCfg, kubeClient, err := createApiserverClient(cfg.APIServerHost, cfg.KubeConfigFilePath)
	if err != nil {
		glog.Fatalf("create k8s client failed, err: %v", err)
	}

	if cfg.PublishService != "" {
		svc := cfg.PublishService
		ns, name, err := utils.ParseNameNS(svc)
		if err != nil {
			glog.Fatal(err)
		}
		_, err = kubeClient.CoreV1().Services(ns).Get(name, metav1.GetOptions{})
		if err != nil {
			glog.Fatalf("unexpected error getting information about service %v: %v", svc, err)
		}
	}

	// check namespace exists
	if cfg.WatchNamespace != "" {
		_, err = kubeClient.CoreV1().Namespaces().Get(cfg.WatchNamespace,
			metav1.GetOptions{})
		if err != nil {
			glog.Fatalf("no namespace with name %v found: %v",
				cfg.WatchNamespace, err)
		}
	}

	// init manba client
	manbaCli, err := manbaClient.NewClient(cfg.ManbaAPIServerTimeout, cfg.ManbaAPIServer)
	if err != nil {
		glog.Fatalf("create manba client failed, err: %v", err)
	}

	controllerConfig := controllerConfigFromCLIConfig(cfg)
	controllerConfig.KubeClient = kubeClient
	controllerConfig.Manba.Client = manbaCli

	updateChannel := channels.NewRingChannel(1024)

	var synced []cache.InformerSynced
	informers := cache2.CreateInformers(kubeClient, restCfg, cfg.SyncPeriod, cfg.WatchNamespace, cfg.IngressClass, updateChannel)

	stopCh := make(chan struct{})
	for _, informer := range informers {
		go informer.Run(stopCh)
		synced = append(synced, informer.HasSynced)
	}
	if !cache.WaitForCacheSync(stopCh, synced...) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
	}

	s := store.New(kubeClient, cache2.GetStore, annotations.IngressClassValidatorFuncFromObjectMeta(controllerConfig.IngressClass))
	manbaController, err := controller.NewManbaController(controllerConfig, updateChannel, s)
	if err != nil {
		glog.Fatalf("create manba controller failed, err: %v", err)
	}

	go handleSigterm(manbaController, stopCh, func(code int) {
		os.Exit(code)
	})

	manbaController.Start()
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

func controllerConfigFromCLIConfig(cfg Config) controller.Config {
	return controller.Config{
		ElectionID:    cfg.ElectionID,
		IngressClass:  cfg.IngressClass,
		ResyncPeriod:  cfg.SyncPeriod,
		SyncRateLimit: cfg.SyncRateLimit,
		Concurrency:   cfg.ManbaConcurrency,

		PublishService:       cfg.PublishService,
		PublishStatusAddress: cfg.PublishStatusAddress,
		UpdateStatus:         cfg.UpdateStatus,
	}
}

type exiter func(code int)

func handleSigterm(manbaC *controller.ManbaController, stopCh chan struct{},
	exit exiter) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	<-signalChan
	glog.Infof("Received SIGTERM, shutting down")

	exitCode := 0
	close(stopCh)
	if err := manbaC.Stop(); err != nil {
		glog.Infof("Error during shutdown %v", err)
		exitCode = 1
	}

	glog.Infof("Handled quit, awaiting pod deletion")
	time.Sleep(10 * time.Second)

	glog.Infof("Exiting with %v", exitCode)
	exit(exitCode)
}
