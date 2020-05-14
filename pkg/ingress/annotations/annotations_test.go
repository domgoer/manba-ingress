package annotations

import (
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIngressClassValidatorFuncFromObjectMeta(t *testing.T) {
	validFunc := IngressClassValidatorFuncFromObjectMeta("work")

	shouldTrue := validFunc(&metav1.ObjectMeta{
		Annotations:map[string]string{
			"kubernetes.io/ingress.class": "work",
		},
	})

	shouldFalse := validFunc(&metav1.ObjectMeta{
		Annotations:map[string]string{
			"kubernetes.io/ingress.class": "not-workd",
		},
	})

	assert.True(t,shouldTrue)
	assert.False(t,shouldFalse)
}

func TestIngressClassValidatorFunc(t *testing.T) {
	validFunc := IngressClassValidatorFunc("work")
	shouldTrue := validFunc(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"kubernetes.io/ingress.class": "work",
			},
		},
	})

	shouldFalse := validFunc(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"kubernetes.io/ingress.class": "not-work",
			},
		},
	})
	assert.True(t,shouldTrue)
	assert.False(t,shouldFalse)
}
