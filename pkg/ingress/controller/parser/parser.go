package parser

import (
	"fmt"
	"sort"

	"github.com/domgoer/manba-ingress/pkg/ingress/store"
	"github.com/fagongzi/manba/pkg/pb/metapb"
	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
)

// Routing represents a Manba Routing and holds a reference to the Ingress
// rule.
type Routing struct {
	metapb.Routing
}

// Cluster contains k8s service and manba cluster
type Cluster struct {
	metapb.Cluster
	Servers []Server

	Namespace  string
	Backend    networkingv1beta1.IngressBackend
	K8SService corev1.Service
}

// Server contains k8s endpoint and manba server
type Server struct {
	metapb.Server

	K8SEndpoints corev1.Endpoints
}

// API contains manba API
type API struct {
	metapb.API

	Routings []Routing

	Ingress networkingv1beta1.Ingress
}

// Plugin implements manba Plugin
type Plugin struct {
	// metapb.Plugin
}

// Parser parses Kubernetes CRDs and Ingress rules and generates a
// Kong configuration.
type Parser struct {
	store store.Store
}

// ManbaSate holds the configuration that should be applied to Manba.
type ManbaSate struct {
	APIs     []API
	Clusters []Cluster
	Servers  []Server
	Routings []Routing
	Plugins  []Plugin
}

type parsedIngressRules struct {
}

// New returns a new parser backed with store.
func New(s store.Store) *Parser {
	return &Parser{store: s}
}

func (p *Parser) parseIngressRules(
	ingressList []*networkingv1beta1.Ingress) (*parsedIngressRules, error) {

	sort.SliceStable(ingressList, func(i, j int) bool {
		return ingressList[i].CreationTimestamp.Before(
			&ingressList[j].CreationTimestamp)
	})
	// generate the following:
	// Services and Routes
	var allDefaultBackends []networkingv1beta1.Ingress
	// secretNameToSNIs := make(map[string][]string)
	// serviceNameToServices := make(map[string]Server)

	for i := 0; i < len(ingressList); i++ {
		ingress := *ingressList[i]
		ingressSpec := ingress.Spec

		if ingressSpec.Backend != nil {
			allDefaultBackends = append(allDefaultBackends, ingress)

		}

		// processTLSSections(ingressSpec.TLS, ingress.Namespace, secretNameToSNIs)
		for i, rule := range ingressSpec.Rules {
			host := rule.Host
			if rule.HTTP == nil {
				continue
			}
			for j, rule := range rule.HTTP.Paths {
				path := rule.Path

				// isACMEChallenge := strings.HasPrefix(path, "/.well-known/acme-challenge/")

				if path == "" {
					path = "/"
				}

				_ = API{
					Ingress: ingress,
					API: metapb.API{
						Name:       fmt.Sprintf("%s.%s.%d%d", ingress.Namespace, ingress.Name, i, j),
						URLPattern: path,
						// no method fields in ingress rule
						Method: "*",
						Domain: host,
					},
				}
			}
		}

	}
	return nil, nil
}
