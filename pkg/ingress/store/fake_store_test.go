package store

import (
	"testing"

	configurationv1beta1 "github.com/domgoer/manba-ingress/pkg/apis/configuration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/stretchr/testify/assert"
)

func TestNewFakeStore(t *testing.T) {
	var objs = []runtime.Object{
		&corev1.Service{
			Spec: corev1.ServiceSpec{
				Type: "",
			},
		},
	}
	var manbaObjs = []runtime.Object{
		&configurationv1beta1.ManbaIngress{
			Spec: configurationv1beta1.ManbaIngressSpec{},
		},
	}
	s, err := NewFakeStore(objs, manbaObjs)
	assert.Nil(t, err)
	assert.NotNil(t, s)
}

func TestFakeStoreEndpiont(t *testing.T) {

	endpoint := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
	}

	store, err := NewFakeStore([]runtime.Object{endpoint}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, store)
	c, err := store.GetEndpointsForService("default", "foo")
	assert.Nil(t, err)
	assert.NotNil(t, c)

	c, err = store.GetEndpointsForService("default", "does-not-exist")
	assert.NotNil(t, err)
	assert.Nil(t, c)
}

func TestFakeStoreService(t *testing.T) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
	}
	store, err := NewFakeStore([]runtime.Object{service}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, store)
	c, err := store.GetService("default", "foo")
	assert.Nil(t, err)
	assert.NotNil(t, c)

	c, err = store.GetService("default", "does-not-exist")
	assert.NotNil(t, err)
	assert.Nil(t, c)
}

func TestFakeStoreListServices(t *testing.T) {
	service1 := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo1",
			Namespace: "default",
			Labels: map[string]string{
				"app": "foo",
			},
		},
	}
	service2 := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo2",
			Namespace: "default",
			Labels: map[string]string{
				"app": "foo",
			},
		},
	}
	store, err := NewFakeStore([]runtime.Object{service1, service2}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, store)
	services, err := store.ListServices("default", map[string]string{
		"app": "foo",
	})
	assert.Nil(t, err)
	assert.Len(t, services, 2)
}

func TestFakeStoreListManbaIngresses(t *testing.T) {
	ingress1 := &configurationv1beta1.ManbaIngress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo1",
			Namespace: "default",
			Labels: map[string]string{
				"app": "foo",
			},
		},
	}
	ingress2 := &configurationv1beta1.ManbaIngress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo2",
			Namespace: "default",
			Labels: map[string]string{
				"app": "foo",
			},
		},
	}
	store, err := NewFakeStore(nil, []runtime.Object{ingress1, ingress2})
	assert.Nil(t, err)
	assert.NotNil(t, store)
	ingresses := store.ListManbaIngresses()
	assert.Len(t, ingresses, 2)
}

func TestFakeStoreGetManbaIngress(t *testing.T) {
	ingress := &configurationv1beta1.ManbaIngress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
	}

	store, err := NewFakeStore(nil, []runtime.Object{ingress})
	assert.Nil(t, err)
	assert.NotNil(t, store)

	i, err := store.GetManbaIngress("default", "foo")
	assert.Nil(t, err)
	assert.NotNil(t, i)
	i, err = store.GetManbaIngress("default", "do-not-exist")
	assert.NotNil(t, err)
	assert.Nil(t, i)
}

func TestFakeStoreGetManbaCluster(t *testing.T) {
	ingress := &configurationv1beta1.ManbaCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
	}

	store, err := NewFakeStore(nil, []runtime.Object{ingress})
	assert.Nil(t, err)
	assert.NotNil(t, store)

	i, err := store.GetManbaCluster("default", "foo")
	assert.Nil(t, err)
	assert.NotNil(t, i)
	i, err = store.GetManbaCluster("default", "do-not-exist")
	assert.NotNil(t, err)
	assert.Nil(t, i)
}
