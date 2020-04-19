package parser

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"

	configurationv1beta1 "github.com/domgoer/manba-ingress/pkg/apis/configuration/v1beta1"

	"github.com/domgoer/manba-ingress/pkg/ingress/annotations"
	"github.com/domgoer/manba-ingress/pkg/ingress/store"
	"github.com/domgoer/manba-ingress/pkg/utils"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	// ErrClusterSubSetNotFound by name
	ErrClusterSubSetNotFound = errors.New("cannot found cluster subset")
)

// Routing represents a Manba Routing and holds a reference to the Ingress
// rule.
type Routing struct {
	APIName     string
	ClusterName string
	metapb.Routing
}

// Service contains k8s service and manba cluster
type Service struct {
	Cluster *Cluster
	Servers []Server

	APIs []API

	Namespace string
	Backend   configurationv1beta1.ManbaCluster
}

// Cluster containers k8s service and manba cluster
type Cluster struct {
	metapb.Cluster

	Servers   []Server
	Namespace string
	K8SSbuSet configurationv1beta1.ManbaClusterSubSet
}

// Server contains k8s endpoint and manba server
type Server struct {
	metapb.Server
}

// API contains manba API
type API struct {
	metapb.API

	Namespace string
	HTTPRule  configurationv1beta1.ManbaHTTPRule
}

// Plugin implements manba Plugin
type Plugin struct {
	// metapb.Plugin
	Name string
}

// Proxy implements manba DispatchNode
type Proxy struct {
	ClusterName string
	metapb.DispatchNode
}

// Parser parses Kubernetes CRDs and Ingress rules and generates a
// Kong configuration.
type Parser struct {
	store store.Store
}

// ManbaState holds the configuration that should be applied to Manba.
type ManbaState struct {
	APIs     []API
	Servers  []Server
	Clusters []Cluster
	Routings []Routing
	Plugins  []Plugin
}

type parsedIngressRules struct {
	ServiceNameToServices map[string]*Service
}

// New returns a new parser backed with store.
func New(s store.Store) *Parser {
	return &Parser{store: s}
}

// Build creates a Manba configuration from Ingress and Custom resources
// defined in Kuberentes.
// It throws an error if there is an error returned from client-go.
func (p *Parser) Build() (*ManbaState, error) {
	var state ManbaState
	ings := p.store.ListManbaIngresses()
	// parse ingress rules
	parsedInfo, err := p.parseIngressRules(ings)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing ingress rules")
	}

	for _, service := range parsedInfo.ServiceNameToServices {

		err = p.fillOverride(service)
		if err != nil {
			return nil, errors.Wrap(err, "overriding ManbaIngress values")
		}
	}

	var keysMap = make(map[string]bool)
	// return true if everything is ok
	check := func(key string) bool {
		if _, ok := keysMap[key]; ok {
			return false
		}
		keysMap[key] = true
		return true
	}

	for _, service := range parsedInfo.ServiceNameToServices {
		for _, api := range service.APIs {
			if check(api.Name) {
				state.APIs = append(state.APIs, api)
			}
		}
		service.Cluster.Servers = service.Servers
		if check(service.Cluster.Name) {
			state.Clusters = append(state.Clusters, *service.Cluster)
		}

		for _, server := range service.Servers {
			if check(server.Addr) {
				state.Servers = append(state.Servers, service.Servers...)
			}
		}

		// TODO: add routing
	}

	return &state, nil
}

func (p *Parser) parseIngressRules(
	ingressList []*configurationv1beta1.ManbaIngress) (*parsedIngressRules, error) {

	sort.SliceStable(ingressList, func(i, j int) bool {
		return ingressList[i].CreationTimestamp.Before(
			&ingressList[j].CreationTimestamp)
	})
	// secretNameToSNIs := make(map[string][]string)
	serviceNameToServices := make(map[string]*Service)

	for i := 0; i < len(ingressList); i++ {
		ingress := *ingressList[i]
		ingressSpec := ingress.Spec

		var apis []API

		for j, rule := range ingressSpec.HTTP {
			base := API{
				API:       metapb.API{},
				Namespace: ingress.Namespace,
				HTTPRule:  rule,
			}

			base.fromManbaHTTPRule(&rule)

			for k, match := range rule.Match {

				api := base
				urlPattern := match.URI.Pattern

				if urlPattern == "" {
					urlPattern = "/"
				}

				api.Name = fmt.Sprintf("%s.%s.%d%d%d", ingress.Namespace, ingress.Name, i, j, k)
				api.URLPattern = urlPattern
				if match.Method != nil {
					api.Method = *match.Method
				} else {
					api.Method = "*"
				}

				// append to list
				apis = append(apis, api)
			}

			for _, backend := range rule.Route {
				serviceName := fmt.Sprintf("%s.%s.%s.%d.svc", ingress.Namespace, backend.Name, backend.Subset, backend.Port)

				service, ok := serviceNameToServices[serviceName]
				if !ok {
					cluster, err := p.store.GetManbaCluster(ingress.Namespace, backend.Name)
					if err != nil {
						glog.Errorf("getting manba cluster: %v", err)
						continue
					}
					subSet, err := p.getClusterSubset(cluster.Spec.Subsets, backend.Subset)
					if err != nil {
						glog.Errorf("getting manba subset: %v", err)
						continue
					}

					service = &Service{
						Cluster: &Cluster{
							Cluster: metapb.Cluster{
								Name: serviceName,
							},
							Namespace: ingress.Namespace,
							K8SSbuSet: subSet,
						},
						Namespace: ingress.Namespace,
						Backend:   *cluster,
					}
				}

				service.APIs = append(service.APIs, apis...)

				serviceNameToServices[serviceName] = service
			}

		}
	}

	return &parsedIngressRules{
		ServiceNameToServices: serviceNameToServices,
	}, nil
}

func (p *Parser) fillOverride(service *Service) error {
	// svcKey := service.Namespace + "/" + service.Backend.ServiceName
	//
	// var err error
	// service.Servers, err = p.getServiceEndpoints(service.Cluster.K8SService, service.Backend.ServicePort)
	// if err != nil {
	// 	glog.Errorf("error getting endpoints for '%v' service: %v",
	// 		svcKey, err)
	// }
	//
	// service.Servers, err = p.fillServersFromPods(service.Pods, service.Servers)
	// if err != nil {
	// 	glog.Errorf("error fill servers for '%v' service: %v",
	// 		svcKey, err)
	// }
	//
	// anns := service.Cluster.K8SService.Annotations
	// overrideCluster(service.Cluster, anns)

	return nil
}

func (p *Parser) getServiceEndpoints(svc corev1.Service,
	backendPort string) ([]Server, error) {
	var servers []Server
	var endpoints []utils.Endpoint
	var servicePort corev1.ServicePort
	svcKey := svc.Namespace + "/" + svc.Name

	for _, port := range svc.Spec.Ports {
		// targetPort could be a string, use the name or the port (int)
		if strconv.Itoa(int(port.Port)) == backendPort ||
			port.TargetPort.String() == backendPort ||
			port.Name == backendPort {
			servicePort = port
			break
		}
	}

	// Ingress with an ExternalName service and no port defined in the service.
	if len(svc.Spec.Ports) == 0 &&
		svc.Spec.Type == corev1.ServiceTypeExternalName {
		externalPort, err := strconv.Atoi(backendPort)
		if err != nil {
			glog.Warningf("only numeric ports are allowed in"+
				" ExternalName services: %v is not valid as a TCP/UDP port",
				backendPort)
			return servers, nil
		}

		servicePort = corev1.ServicePort{
			Protocol:   corev1.ProtocolTCP,
			Port:       int32(externalPort),
			TargetPort: intstr.FromString(backendPort),
		}
	}

	endpoints = getEndpoints(&svc, &servicePort,
		corev1.ProtocolTCP, p.store.GetEndpointsForService)
	if len(endpoints) == 0 {
		glog.Warningf("service %v does not have any active endpoints",
			svcKey)
	}
	for _, endpoint := range endpoints {
		if endpoint.Port != backendPort {
			continue
		}
		server := Server{
			Server: metapb.Server{
				Addr: endpoint.String(),
			},
		}
		servers = append(servers, server)
	}
	return servers, nil
}

// getEndpoints returns a list of <endpoint ip>:<port> for a given service/target port combination.
func getEndpoints(
	s *corev1.Service,
	port *corev1.ServicePort,
	proto corev1.Protocol,
	getEndpoints func(string, string) (*corev1.Endpoints, error),
) []utils.Endpoint {

	upsServers := []utils.Endpoint{}

	if s == nil || port == nil {
		return upsServers
	}

	// avoid duplicated upstream servers when the service
	// contains multiple port definitions sharing the same
	// targetport.
	adus := make(map[string]bool)

	// ExternalName services
	if s.Spec.Type == corev1.ServiceTypeExternalName {
		glog.V(3).Infof("Ingress using a service %v of type=ExternalName", s.Name)

		targetPort := port.TargetPort.IntValue()
		// check for invalid port value
		if targetPort <= 0 {
			glog.Errorf("ExternalName service with an invalid port: %v", targetPort)
			return upsServers
		}

		return append(upsServers, utils.Endpoint{
			Address: s.Spec.ExternalName,
			Port:    fmt.Sprintf("%v", targetPort),
		})
	}

	glog.V(3).Infof("getting endpoints for service %v/%v and port %v", s.Namespace, s.Name, port.String())
	ep, err := getEndpoints(s.Namespace, s.Name)
	if err != nil {
		glog.Warningf("unexpected error obtaining service endpoints: %v", err)
		return upsServers
	}

	for _, ss := range ep.Subsets {
		for _, epPort := range ss.Ports {

			if !reflect.DeepEqual(epPort.Protocol, proto) {
				continue
			}

			var targetPort int32

			if port.Name == "" {
				// port.Name is optional if there is only one port
				targetPort = epPort.Port
			} else if port.Name == epPort.Name {
				targetPort = epPort.Port
			}

			// check for invalid port value
			if targetPort <= 0 {
				continue
			}

			for _, epAddress := range ss.Addresses {
				ep := fmt.Sprintf("%v:%v", epAddress.IP, targetPort)
				if _, exists := adus[ep]; exists {
					continue
				}
				ups := utils.Endpoint{
					Address: epAddress.IP,
					Port:    fmt.Sprintf("%v", targetPort),
				}
				upsServers = append(upsServers, ups)
				adus[ep] = true
			}
		}
	}

	glog.V(3).Infof("endpoints found: %v", upsServers)
	return upsServers
}

func (p *Parser) getManbaIngressForService(service corev1.Service) (*configurationv1beta1.ManbaIngress, error) {
	configName := annotations.ExtractConfigurationName(service.GetAnnotations())
	if configName == "" {
		return nil, nil
	}
	return p.store.GetManbaIngress(service.Namespace, configName)
}

func overrideCluster(cluster *Cluster, annos map[string]string) {
	if cluster == nil {
		return
	}

	cluster.LoadBalance = annotations.ExtractLoadBalancer(annos)
}

func (p *Parser) getManbaIngressFromIngress(ingress *networkingv1beta1.Ingress) (*configurationv1beta1.ManbaIngress, error) {
	configName := annotations.ExtractConfigurationName(ingress.GetAnnotations())
	if configName != "" {
		mi, err := p.store.GetManbaIngress(ingress.Namespace, configName)
		if err == nil {
			return mi, nil
		}
	}

	return p.store.GetManbaIngress(ingress.Namespace, ingress.Name)
}

func (p *Parser) getClusterSubset(list []configurationv1beta1.ManbaClusterSubSet, subset string) (res configurationv1beta1.ManbaClusterSubSet, err error) {
	for _, data := range list {
		if data.Name == subset {
			res = data
			return
		}
	}
	err = ErrClusterSubSetNotFound
	return
}

func (a *API) fromManbaHTTPRule(rule *configurationv1beta1.ManbaHTTPRule) {
	if a == nil {
		a = &API{}
	}

	meta := a.API
	meta.DefaultValue = rule.DefaultValue
	meta.RenderTemplate = rule.RenderTemplate
	if rule.AuthFilter != nil {
		meta.AuthFilter = *rule.AuthFilter
	}
	// meta.MaxQPS = path.MaxQPS
	// meta.CircuitBreaker = path.CircuitBreaker
	// meta.RateLimitOption = path.RateLimitOption

	a.API = meta

}
