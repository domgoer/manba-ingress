package store

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	manbaFake "github.com/domgoer/manba-ingress/pkg/client/clientset/versioned/fake"

	configurationinformer "github.com/domgoer/manba-ingress/pkg/client/informers/externalversions"
	"github.com/domgoer/manba-ingress/pkg/ingress/annotations"
	k8sInformers "k8s.io/client-go/informers"
	k8sFake "k8s.io/client-go/kubernetes/fake"
)

// NewFakeStore creates a store backed by the objects passed in as arguments.
func NewFakeStore(
	k8sObjects []runtime.Object,manbaObjects []runtime.Object) (Store, error) {
	var s Store
	fakeClient := k8sFake.NewSimpleClientset(k8sObjects...)
	manbaFakeClient := manbaFake.NewSimpleClientset(manbaObjects...)

	s = &store{
		client:       fakeClient,
		factory:      k8sInformers.NewSharedInformerFactory(fakeClient, time.Hour),
		manbaFactory: configurationinformer.NewSharedInformerFactory(manbaFakeClient, time.Hour),

		isValidIngresClass: annotations.IngressClassValidatorFuncFromObjectMeta("kong"),
	}
	return s, nil
}
