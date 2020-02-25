package store

import (
	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Store is the interface that wraps the required methods to gather information
// about ingresses, services, secrets and ingress annotations.
type Store interface {
	GetEndpointsForService(namespace, name string) (*corev1.Endpoints, error)
	ListIngresses() []*networkingv1beta1.Ingress
	GetService(namespace, name string) (*corev1.Service, error)
	GetPodsForService(namespace, name string) ([]corev1.Pod, error)
}

type store struct {
	isValidIngresClass func(objectMeta *metav1.ObjectMeta) bool
}

func (s *store) GetEndpointsForService(namespace, name string) (*corev1.Endpoints, error) {
	panic("implement me")
}

func (s *store) ListIngresses() []*networkingv1beta1.Ingress {
	panic("implement me")
}

func (s *store) GetService(namespace, name string) (*corev1.Service, error) {
	panic("implement me")
}

func (s *store) GetPodsForService(namespace, name string) ([]corev1.Pod, error) {
	panic("implement me")
}

// New creates a new object store to be used in the ingress controller
func New(isValidIngresClassFunc func(objectMeta *metav1.ObjectMeta) bool) Store {
	return &store{
		isValidIngresClass: isValidIngresClassFunc,
	}
}
