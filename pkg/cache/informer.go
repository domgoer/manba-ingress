package cache

import (
	"time"

	configurationclientv1 "github.com/domgoer/manba-ingress/pkg/client/clientset/versioned"
	configurationinformer "github.com/domgoer/manba-ingress/pkg/client/informers/externalversions"
	"k8s.io/client-go/rest"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func CreateInformers(k8sCli kubernetes.Interface, cfg *rest.Config, syncPeriod time.Duration, namespace string) []cache.SharedIndexInformer {

	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		k8sCli,
		syncPeriod,
		informers.WithNamespace(namespace),
	)

	confClient, _ := configurationclientv1.NewForConfig(cfg)
	manbaInformerFactory := configurationinformer.NewSharedInformerFactoryWithOptions(confClient, syncPeriod, configurationinformer.WithNamespace(namespace))

	var informers []cache.SharedIndexInformer

	// create ingress informer
	ingInformer := informerFactory.Networking().V1beta1().Ingresses().Informer()
	storesMap["ingress"] = ingInformer.GetStore()
	informers = append(informers, ingInformer)

	// create endpoint informer
	epInformer := informerFactory.Core().V1().Endpoints().Informer()
	storesMap["endpoint"] = ingInformer.GetStore()
	informers = append(informers, epInformer)

	// create service informer
	svcInformer := informerFactory.Core().V1().Services().Informer()
	storesMap["service"] = ingInformer.GetStore()
	informers = append(informers, svcInformer)

	podInformer := informerFactory.Core().V1().Pods().Informer()
	storesMap["pod"] = podInformer.GetStore()
	informers = append(informers, podInformer)

	manbaIngInformer := manbaInformerFactory.Configuration().V1beta1().ManbaIngresses().Informer()
	storesMap["manbaIng"] = manbaIngInformer.GetStore()
	informers = append(informers, manbaIngInformer)

	return informers
}
