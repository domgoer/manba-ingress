package annotations

import (
	"encoding/json"
	"strconv"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ingressClassKey = "kubernetes.io/ingress.class"

	loadBalanceAnnotationKey = "configuration.manba.io/loadBalancer"
	maxQPSAnnotationKey      = "configuration.manba.io/maxQPS"

	circuitBreakerAnnotationKey = "configuration.manba.io/circuitBreaker"

	// DefaultIngressClass defines the default class used
	// by Kong's ingress controller.
	DefaultIngressClass = "manba"
)

// IngressClassValidatorFuncFromObjectMeta returns a function which
// can validate if an ObjectMeta belongs to an the ingressClass or not.
func IngressClassValidatorFuncFromObjectMeta(
	ingressClass string) func(obj *metav1.ObjectMeta) bool {

	return func(obj *metav1.ObjectMeta) bool {
		ingress := obj.GetAnnotations()[ingressClassKey]
		return validIngress(ingress, ingressClass)
	}
}

func validIngress(ingressAnnotationValue, ingressClass string) bool {
	// we have 2 valid combinations
	// 1 - ingress with default class | blank annotation on ingress
	// 2 - ingress with specific class | same annotation on ingress
	//
	// and 2 invalid combinations
	// 3 - ingress with default class | fixed annotation on ingress
	// 4 - ingress with specific class | different annotation on ingress
	if ingressAnnotationValue == "" && ingressClass == DefaultIngressClass {
		return true
	}
	return ingressAnnotationValue == ingressClass
}

// ExtractLoadBalancer extracts the lb supplied in the annotation
func ExtractLoadBalancer(anns map[string]string) metapb.LoadBalance {
	return metapb.LoadBalance(metapb.LoadBalance_value[anns[loadBalanceAnnotationKey]])
}

// ExtractMaxQPS extracts the max qps of server
func ExtractMaxQPS(anns map[string]string) int64 {
	i, _ := strconv.Atoi(anns[maxQPSAnnotationKey])
	return int64(i)
}

// ExtrackCircuitBreaker extracts the circuitBreaker of server
func ExtrackCircuitBreaker(anns map[string]string) *metapb.CircuitBreaker {
	data := anns[circuitBreakerAnnotationKey]
	if data == "" {
		return nil
	}

	res := new(metapb.CircuitBreaker)
	err := json.Unmarshal([]byte(data), res)
	if err != nil {
		return nil
	}
	return res
}

func HasManbaServiceAnnotation(anns map[string]string) bool {
	return true
}
