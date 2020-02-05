package cache

import (
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func CreateInformers(k8sCli kubernetes.Interface, syncPeriod time.Duration, namespace string) []cache.SharedIndexInformer {

	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		k8sCli,
		syncPeriod,
		informers.WithNamespace(namespace),
	)

	var informers []cache.SharedIndexInformer

	// create ingress informer
	ingInformer := informerFactory.Networking().V1beta1().Ingresses().Informer()
	informers = append(informers, ingInformer)

	// create endpoint informer
	epInformer := informerFactory.Core().V1().Endpoints().Informer()
	informers = append(informers, epInformer)

	// create service informer
	svcInformer := informerFactory.Core().V1().Services().Informer()
	informers = append(informers, svcInformer)
	return informers
}
