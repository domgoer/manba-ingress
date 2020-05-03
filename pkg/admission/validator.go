package admission

import (
	"fmt"
	"regexp"

	configurationv1beta1 "github.com/domgoer/manba-ingress/pkg/apis/configuration/v1beta1"
	configurationinformer "github.com/domgoer/manba-ingress/pkg/client/informers/externalversions"
	"k8s.io/apimachinery/pkg/api/errors"
)

// ManbaValidator validates Manba entities.
type ManbaValidator interface {
	ValidateManbaIngress(*configurationv1beta1.ManbaIngress) (bool, string, error)
}

// validator implements ManbaValidator
type validator struct {
	manbaInformer configurationinformer.SharedInformerFactory
}

var _ ManbaValidator = &validator{}

// NewValidator returns new validator with manba factory
func NewValidator(informer configurationinformer.SharedInformerFactory) *validator {
	return &validator{manbaInformer: informer}
}

// ValidatePlugin checks if manba ingress is valid. It does so by
// checking whether cluster is existed, parameter values and others.
// If an error occurs during validation, it is returned as the last argument.
// The first boolean communicates if manba ingress is valid or not and string
// holds a message if the entity is not valid
func (v *validator) ValidateManbaIngress(ingress *configurationv1beta1.ManbaIngress) (bool, string, error) {
	var clusters []configurationv1beta1.ManbaHTTPRouteCluster
	for _, rule := range ingress.Spec.HTTP {
		for _, route := range rule.Route {
			clusters = append(clusters, route.Cluster)

			// check parameter value
			if !v.isManbaHTTPRouteMatchValid(route.Match) {
				return false, "manba http route match value must conform to the regular expression rule", nil
			}
		}

		for _, mirror := range rule.Mirror {
			clusters = append(clusters, mirror.Cluster)
		}

		for _, split := range rule.Split {
			clusters = append(clusters, split.Cluster)
		}

	}

	// check cluster exists
	for _, cluster := range clusters {
		exist, err := v.isClusterExist(ingress.GetNamespace(), cluster)
		if err != nil {
			return false, "", err
		}
		if !exist {
			return false, fmt.Sprintf("manba cluster %s/%s not found", ingress.GetNamespace(), cluster.Name), nil
		}
	}
	return true, "", nil
}

func (v *validator) isClusterExist(namespace string, cluster configurationv1beta1.ManbaHTTPRouteCluster) (bool, error) {
	cls, err := v.manbaInformer.Configuration().V1beta1().ManbaClusters().Lister().ManbaClusters(namespace).Get(cluster.Name)
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	exist := false
	for _, subSet := range cls.Spec.Subsets {
		if cluster.Subset == subSet.Name {
			exist = true
		}
	}
	return exist, nil
}

func (v *validator) isManbaHTTPRouteMatchValid(match *configurationv1beta1.ManbaHTTPRouteMatch) bool {
	if match == nil {
		return true
	}

	// return true if v matches regexp rule
	isValid := func(data map[string]string) bool {
		for _, v := range data {
			_, err := regexp.Compile(v)
			if err != nil {
				return false
			}
		}
		return true
	}

	return isValid(match.PathValue) && isValid(match.Query) && isValid(match.JSONBody) && isValid(match.Header) && isValid(match.FormData) && isValid(match.Cookie)
}
