package cache

import (
	"time"

	configurationclientv1 "github.com/domgoer/manba-ingress/pkg/client/clientset/versioned"
	configurationinformer "github.com/domgoer/manba-ingress/pkg/client/informers/externalversions"
	"github.com/domgoer/manba-ingress/pkg/ingress/annotations"
	"github.com/domgoer/manba-ingress/pkg/ingress/controller"
	"github.com/eapache/channels"
	"k8s.io/client-go/rest"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// CreateInformers creates ingress, ep, svc, manbaIngress and pods' informers
func CreateInformers(k8sCli kubernetes.Interface, cfg *rest.Config, syncPeriod time.Duration, namespace, ingressClass string, updateChannel *channels.RingChannel) []cache.SharedIndexInformer {
	reh := controller.ResourceEventHandler{
		UpdateCh:           updateChannel,
		IsValidIngresClass: annotations.IngressClassValidatorFunc(ingressClass),
	}

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
	ingInformer.AddEventHandler(reh)
	informers = append(informers, ingInformer)

	// create endpoint informer
	epInformer := informerFactory.Core().V1().Endpoints().Informer()
	storesMap["endpoint"] = epInformer.GetStore()
	epInformer.AddEventHandler(controller.EndpointsEventHandler{
		UpdateCh: updateChannel,
	})
	informers = append(informers, epInformer)

	// create service informer
	svcInformer := informerFactory.Core().V1().Services().Informer()
	storesMap["service"] = svcInformer.GetStore()
	svcInformer.AddEventHandler(reh)
	informers = append(informers, svcInformer)

	manbaIngInformer := manbaInformerFactory.Configuration().V1beta1().ManbaIngresses().Informer()
	storesMap["manbaIng"] = manbaIngInformer.GetStore()
	manbaIngInformer.AddEventHandler(reh)
	informers = append(informers, manbaIngInformer)

	return informers
}
