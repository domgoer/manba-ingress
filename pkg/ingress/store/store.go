package store

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/runtime"

	configurationv1beta1 "github.com/domgoer/manba-ingress/pkg/apis/configuration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	ep  = "endpoint"
	ing = "ingress"
	svc = "service"
	mi  = "manbaIng"
)

// Store is the interface that wraps the required methods to gather information
// about ingresses, services, secrets and ingress annotations.
type Store interface {
	GetEndpointsForService(namespace, name string) (*corev1.Endpoints, error)
	ListIngresses() []*networkingv1beta1.Ingress
	GetService(namespace, name string) (*corev1.Service, error)
	GetPodsForService(namespace, name string) ([]corev1.Pod, error)
	GetManbaIngress(namespace, name string) (*configurationv1beta1.ManbaIngress, error)
}

type store struct {
	client             kubernetes.Interface
	getStore           func(string) cache.Store
	isValidIngresClass func(objectMeta *metav1.ObjectMeta) bool
}

var ingressConversionScheme *runtime.Scheme

func init() {
	ingressConversionScheme = runtime.NewScheme()
	extensionsv1beta1.AddToScheme(ingressConversionScheme)
	networkingv1beta1.AddToScheme(ingressConversionScheme)
}

func (s *store) GetEndpointsForService(namespace, name string) (*corev1.Endpoints, error) {
	key := fmt.Sprintf("%v/%v", namespace, name)
	eps, exists, err := s.getStore(ep).GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("could not find endpoints for service %v", key)
	}
	return eps.(*corev1.Endpoints), nil
}

// ListIngresses returns the list of Ingresses
func (s *store) ListIngresses() []*networkingv1beta1.Ingress {
	// filter ingress rules
	var ingresses []*networkingv1beta1.Ingress
	for _, item := range s.getStore(ing).List() {
		ing := networkingIngressV1Beta1(item)
		if !s.isValidIngresClass(&ing.ObjectMeta) {
			continue
		}
		ingresses = append(ingresses, ing)
	}

	return ingresses
}

func (s *store) GetService(namespace, name string) (*corev1.Service, error) {
	key := fmt.Sprintf("%v/%v", namespace, name)
	service, exists, err := s.getStore(svc).GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("service %v was not found", key)
	}
	return service.(*corev1.Service), nil
}

func (s *store) GetPodsForService(namespace, name string) ([]corev1.Pod, error) {
	svc, err := s.GetService(namespace, name)
	if err != nil {
		return nil, err
	}
	var selector []string
	for k, v := range svc.Spec.Selector {
		selector = append(selector, k+"="+v)
	}

	pods, err := s.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: strings.Join(selector, ","),
	})
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

func (s *store) GetManbaIngress(namespace, name string) (*configurationv1beta1.ManbaIngress, error) {
	key := fmt.Sprintf("%v/%v", namespace, name)
	p, exist, err := s.getStore(mi).GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("ManbaIngress %v was not found", key)
	}
	return p.(*configurationv1beta1.ManbaIngress), nil
}

// New creates a new object store to be used in the ingress controller
func New(kc kubernetes.Interface, getStore func(string) cache.Store, isValidIngresClassFunc func(objectMeta *metav1.ObjectMeta) bool) Store {
	return &store{
		client:             kc,
		getStore:           getStore,
		isValidIngresClass: isValidIngresClassFunc,
	}
}

func networkingIngressV1Beta1(obj interface{}) *networkingv1beta1.Ingress {
	networkingIngress, okNetworking := obj.(*networkingv1beta1.Ingress)
	if okNetworking {
		return networkingIngress
	}
	extensionsIngress, okExtension := obj.(*extensionsv1beta1.Ingress)
	if !okExtension {
		glog.Errorf("ingress resource can not be casted to extensions.Ingress" +
			" or networking.Ingress")
		return nil
	}
	networkingIngress = &networkingv1beta1.Ingress{}
	err := ingressConversionScheme.Convert(extensionsIngress, networkingIngress, nil)
	if err != nil {
		glog.Error("failed to convert extensions.Ingress "+
			"to networking.Ingress", err)
		return nil
	}
	return networkingIngress
}
