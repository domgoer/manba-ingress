package store

import (
	"testing"

	configurationv1beta1 "github.com/domgoer/manba-ingress/pkg/apis/configuration/v1beta1"
	corev1 "k8s.io/api/core/v1"
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
