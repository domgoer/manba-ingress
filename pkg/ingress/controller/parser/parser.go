package parser

import (
	"fmt"
	"sort"

	"github.com/domgoer/manba-ingress/pkg/ingress/annotations"
	"github.com/domgoer/manba-ingress/pkg/ingress/store"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
)

// Routing represents a Manba Routing and holds a reference to the Ingress
// rule.
type Routing struct {
	metapb.Routing
}

// Service represents a container of apis with the same host
type Service struct {
	APIs []API

	Namespace string
	Backend   networkingv1beta1.IngressBackend
}

// Cluster contains k8s service and manba cluster
type Cluster struct {
	metapb.Cluster
	Servers []Server

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
	Nodes    []Cluster
	Ingress  networkingv1beta1.Ingress
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
	ServiceNameToServices map[string]Service
}

// New returns a new parser backed with store.
func New(s store.Store) *Parser {
	return &Parser{store: s}
}

// Build creates a Manba configuration from Ingress and Custom resources
// defined in Kuberentes.
// It throws an error if there is an error returned from client-go.
func (p *Parser) Build() (*ManbaSate, error) {
	var state ManbaSate
	ings := p.store.ListIngresses()
	// parse ingress rules
	parsedInfo, err := p.parseIngressRules(ings)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing ingress rules")
	}

	// add the apis to the state
	for _, service := range parsedInfo.ServiceNameToServices {
		state.APIs = append(state.APIs, service.APIs...)
	}

	state.Clusters, err = p.getClusters(parsedInfo.ServiceNameToServices)
	if err != nil {
		return nil, errors.Wrap(err, "building clusters")
	}

	// merge ManbaIngress with APIS and Routings
	err = p.fillOverrides(state)
	if err != nil {
		return nil, errors.Wrap(err, "overriding ManbaIngress values")
	}

	err = p.fillServers(state)
	if err != nil {
		return nil, errors.Wrap(err, "building servers")
	}

	return &state, nil
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
	serviceNameToServices := make(map[string]Service)

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

				i := API{
					API: metapb.API{
						Name:       fmt.Sprintf("%s.%s.%d%d", ingress.Namespace, ingress.Name, i, j),
						URLPattern: path,
						// no method fields in ingress rule
						Method: "*",
					},
					Ingress: ingress,
				}

				if host != "" {
					i.API.Domain = host
				}

				serviceName := ingress.Namespace + "." +
					rule.Backend.ServiceName + "." +
					rule.Backend.ServicePort.String()

				service, ok := serviceNameToServices[serviceName]
				if !ok {
					service = Service{
						Namespace: ingress.Namespace,
						Backend:   rule.Backend,
					}
				}

				service.APIs = append(service.APIs, i)
				serviceNameToServices[serviceName] = service
			}
		}

	}

	sort.SliceStable(allDefaultBackends, func(i, j int) bool {
		return allDefaultBackends[i].CreationTimestamp.Before(&allDefaultBackends[j].CreationTimestamp)
	})
	// Process the default backend
	if len(allDefaultBackends) > 0 {
		ingress := allDefaultBackends[0]
		defaultBackend := allDefaultBackends[0].Spec.Backend
		serviceName := allDefaultBackends[0].Namespace + "." +
			defaultBackend.ServiceName + "." +
			defaultBackend.ServicePort.String()
		service, ok := serviceNameToServices[serviceName]
		if !ok {
			service = Service{
				Namespace: ingress.Namespace,
				Backend:   *defaultBackend,
			}
		}

		i := API{
			API: metapb.API{
				Name:       serviceName,
				URLPattern: "/",
				// no method fields in ingress rule
				Method: "*",
			},
			Ingress: ingress,
		}

		service.APIs = append(service.APIs, i)
		serviceNameToServices[serviceName] = service
	}

	return &parsedIngressRules{
		ServiceNameToServices: serviceNameToServices,
	}, nil
}

func (p *Parser) getClusters(serviceMap map[string]Service) ([]Cluster, error) {
	var clusters []Cluster
	for _, service := range serviceMap {
		clusterName := fmt.Sprintf("%s.%s.%s.svc", service.Backend.ServiceName, service.Namespace, service.Backend.ServicePort.String())
		cluster := Cluster{
			Cluster: metapb.Cluster{
				Name: clusterName,
			},
		}

		k8sSvc, err := p.store.GetService(service.Namespace, service.Backend.ServiceName)
		if err != nil {
			return nil, errors.Wrap(err, "get k8s service")
		}

		cluster.Cluster.LoadBalance = annotations.ExtractLoadBalancer(k8sSvc.GetAnnotations())
	}
	return clusters, nil
}

func (p *Parser) getServers(pods []corev1.Pod) ([]Server, error) {
	var servers []Server
	for _, pod := range pods {
		server := Server{
			Server: metapb.Server{
				Addr:           "",
				MaxQPS:         0,
				CircuitBreaker: &metapb.CircuitBreaker{},
				HeathCheck:     &metapb.HeathCheck{},
			},
		}
	}
	return nil, nil
}

func (p *Parser) fillOverrides(state ManbaSate) error {
	return nil
}

func (p *Parser) fillServers(state ManbaSate) error {
	for _, cls := range state.Clusters {
		state.Servers = append(state.Servers, cls.Servers...)
	}
	return nil
}
