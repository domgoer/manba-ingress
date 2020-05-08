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
func CreateInformers(k8sCli kubernetes.Interface, cfg *rest.Config, syncPeriod time.Duration, namespace, ingressClass string, updateChannel *channels.RingChannel) ([]cache.SharedIndexInformer, informers.SharedInformerFactory, configurationinformer.SharedInformerFactory) {
	reh := controller.ResourceEventHandler{
		UpdateCh:           updateChannel,
		IsValidIngresClass: annotations.IngressClassValidatorFunc(ingressClass),
	}

	factory := informers.NewSharedInformerFactoryWithOptions(
		k8sCli,
		syncPeriod,
		informers.WithNamespace(namespace),
	)

	confClient, _ := configurationclientv1.NewForConfig(cfg)
	manbaFactory := configurationinformer.NewSharedInformerFactoryWithOptions(confClient, syncPeriod, configurationinformer.WithNamespace(namespace))

	var informers []cache.SharedIndexInformer

	// create endpoint informer
	epInformer := factory.Core().V1().Endpoints().Informer()
	epInformer.AddEventHandler(controller.EndpointsEventHandler{
		UpdateCh: updateChannel,
	})
	informers = append(informers, epInformer)

	// create service informer
	svcInformer := factory.Core().V1().Services().Informer()
	svcInformer.AddEventHandler(reh)
	informers = append(informers, svcInformer)

	manbaIngInformer := manbaFactory.Configuration().V1beta1().ManbaIngresses().Informer()
	manbaIngInformer.AddEventHandler(reh)
	informers = append(informers, manbaIngInformer)

	manbaClusterInformer := manbaFactory.Configuration().V1beta1().ManbaClusters().Informer()
	manbaClusterInformer.AddEventHandler(reh)
	informers = append(informers, manbaClusterInformer)

	return informers, factory, manbaFactory
}
