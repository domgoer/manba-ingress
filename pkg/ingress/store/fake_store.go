package store

import (
	"fmt"
	"strings"

	configurationv1beta1 "github.com/domgoer/manba-ingress/pkg/apis/configuration/v1beta1"
	"github.com/domgoer/manba-ingress/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s.io/apimachinery/pkg/runtime"

	manbaFake "github.com/domgoer/manba-ingress/pkg/client/clientset/versioned/fake"

	k8sFake "k8s.io/client-go/kubernetes/fake"
)

type fakeStore struct {
	client      kubernetes.Interface
	manbaClient versioned.Interface
}

// NewFakeStore creates a store backed by the objects passed in as arguments.
func NewFakeStore(
	k8sObjects []runtime.Object, manbaObjects []runtime.Object) (Store, error) {
	var s Store
	fakeClient := k8sFake.NewSimpleClientset(k8sObjects...)
	manbaFakeClient := manbaFake.NewSimpleClientset(manbaObjects...)

	s = &fakeStore{
		client:      fakeClient,
		manbaClient: manbaFakeClient,
	}
	return s, nil
}

func (f *fakeStore) GetEndpointsForService(namespace, name string) (*corev1.Endpoints, error) {
	return f.client.CoreV1().Endpoints(namespace).Get(name, metav1.GetOptions{})
}

func (f *fakeStore) GetService(namespace, name string) (*corev1.Service, error) {
	return f.client.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
}

func (f *fakeStore) ListServices(namespace string, label map[string]string) ([]*corev1.Service, error) {
	var labelList []string
	for k, v := range label {
		labelList = append(labelList, fmt.Sprintf("%s=%s", k, v))
	}
	svc, err := f.client.CoreV1().Services(namespace).List(metav1.ListOptions{
		LabelSelector: strings.Join(labelList, ","),
	})
	if err != nil {
		return nil, err
	}
	var res []*corev1.Service
	for _, item := range svc.Items {
		res = append(res, &item)
	}
	return res, nil
}

func (f *fakeStore) GetManbaIngress(namespace, name string) (*configurationv1beta1.ManbaIngress, error) {
	return f.manbaClient.ConfigurationV1beta1().ManbaIngresses(namespace).Get(name, metav1.GetOptions{})
}

func (f *fakeStore) GetManbaCluster(namespace, name string) (*configurationv1beta1.ManbaCluster, error) {
	return f.manbaClient.ConfigurationV1beta1().ManbaClusters(namespace).Get(name, metav1.GetOptions{})
}

func (f *fakeStore) ListManbaIngresses() []*configurationv1beta1.ManbaIngress {
	ing, err := f.manbaClient.ConfigurationV1beta1().ManbaIngresses(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	var res []*configurationv1beta1.ManbaIngress
	for _, item := range ing.Items {
		res = append(res, &item)
	}
	return res
}

func (f *fakeStore) GetSecret(namespace, name string) (*corev1.Secret, error) {
	return f.client.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
}
