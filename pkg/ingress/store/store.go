package store

import (
	"fmt"
	"k8s.io/client-go/tools/cache"

	"github.com/domgoer/manba-ingress/pkg/client/informers/externalversions"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/runtime"

	configurationv1beta1 "github.com/domgoer/manba-ingress/pkg/apis/configuration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	manbaIngress = iota
	manbaCluster
	service
	endpoint
)

// Store is the interface that wraps the required methods to gather information
// about ingresses, services, secrets and ingress annotations.
type Store interface {
	// GetEndpointsForService list all endpoints by service selector
	GetEndpointsForService(namespace, name string) (*corev1.Endpoints, error)
	GetService(namespace, name string) (*corev1.Service, error)
	ListServices(namespace string, label map[string]string) ([]*corev1.Service, error)
	GetManbaIngress(namespace, name string) (*configurationv1beta1.ManbaIngress, error)
	GetManbaCluster(namespace, name string) (*configurationv1beta1.ManbaCluster, error)
	ListManbaIngresses() []*configurationv1beta1.ManbaIngress
	GetSecret(namespace, name string) (*corev1.Secret, error)
}

type store struct {
	client       kubernetes.Interface
	factory      informers.SharedInformerFactory
	manbaFactory externalversions.SharedInformerFactory

	isValidIngresClass func(objectMeta *metav1.ObjectMeta) bool
}

func (s *store) GetSecret(namespace, name string) (*corev1.Secret, error) {
	return s.factory.Core().V1().Secrets().Lister().Secrets(namespace).Get(name)
}

func (s *store) ListServices(namespace string, label map[string]string) ([]*corev1.Service, error) {
	return s.factory.Core().V1().Services().Lister().Services(namespace).List(labels.SelectorFromSet(label))
}

var ingressConversionScheme *runtime.Scheme

func init() {
	ingressConversionScheme = runtime.NewScheme()
	extensionsv1beta1.AddToScheme(ingressConversionScheme)
	networkingv1beta1.AddToScheme(ingressConversionScheme)
}

func (s *store) GetEndpointsForService(namespace, name string) (*corev1.Endpoints, error) {
	key := fmt.Sprintf("%v/%v", namespace, name)
	eps, exists, err := s.getStore(endpoint).GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("could not find endpoints for service %v", key)
	}
	return eps.(*corev1.Endpoints), nil
}

// ListManbaIngresses returns the list of Manba Ingresses
func (s *store) ListManbaIngresses() []*configurationv1beta1.ManbaIngress {
	// filter ingress rules
	var ingresses []*configurationv1beta1.ManbaIngress
	for _, item := range s.getStore(manbaIngress).List() {
		ing, ok := item.(*configurationv1beta1.ManbaIngress)
		if !ok {
			glog.Warningf("invalid type for ingress, %v", item)
			continue
		}
		if !s.isValidIngresClass(&ing.ObjectMeta) {
			continue
		}
		ingresses = append(ingresses, ing)
	}

	return ingresses
}

func (s *store) GetService(namespace, name string) (*corev1.Service, error) {
	key := fmt.Sprintf("%v/%v", namespace, name)
	service, exists, err := s.getStore(service).GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("service %v was not found", key)
	}
	return service.(*corev1.Service), nil
}

func (s *store) GetManbaIngress(namespace, name string) (*configurationv1beta1.ManbaIngress, error) {
	key := fmt.Sprintf("%v/%v", namespace, name)
	p, exist, err := s.getStore(manbaIngress).GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("ManbaIngress %v was not found", key)
	}
	return p.(*configurationv1beta1.ManbaIngress), nil
}

func (s *store) GetManbaCluster(namespace, name string) (*configurationv1beta1.ManbaCluster, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	p, exist, err := s.getStore(manbaCluster).GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("ManbaCluster %v was not found", key)
	}
	return p.(*configurationv1beta1.ManbaCluster), nil
}

func (s *store) getStore(t int) cache.Store {
	switch t {
	case manbaCluster:
		return s.manbaFactory.Configuration().V1beta1().ManbaClusters().Informer().GetStore()
	case manbaIngress:
		return s.manbaFactory.Configuration().V1beta1().ManbaIngresses().Informer().GetStore()
	case service:
		return s.factory.Core().V1().Services().Informer().GetStore()
	case endpoint:
		return s.factory.Core().V1().Endpoints().Informer().GetStore()
	}
	return nil
}

// New creates a new object store to be used in the ingress controller
func New(kc kubernetes.Interface, factory informers.SharedInformerFactory, manbaFactory externalversions.SharedInformerFactory, isValidIngresClassFunc func(objectMeta *metav1.ObjectMeta) bool) Store {
	return &store{
		client:             kc,
		factory:            factory,
		manbaFactory:       manbaFactory,
		isValidIngresClass: isValidIngresClassFunc,
	}
}
