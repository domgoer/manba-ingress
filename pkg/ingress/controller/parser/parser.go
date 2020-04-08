package parser

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

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

// Routing represents a Manba Routing and holds a reference to the Ingress
// rule.
type Routing struct {
	APIName     string
	ClusterName string
	metapb.Routing
}

// Service contains k8s service and manba cluster
type Service struct {
	metapb.Cluster
	Servers []Server

	APIs []API

	Namespace  string
	K8SService corev1.Service
	Backend    networkingv1beta1.IngressBackend
}

// Server contains k8s endpoint and manba server
type Server struct {
	metapb.Server

	K8SPod corev1.Pod
}

// API contains manba API
type API struct {
	metapb.API

	Routings []Routing
	Proxies  []Proxy
	Ingress  networkingv1beta1.Ingress
}

// Plugin implements manba Plugin
type Plugin struct {
	// metapb.Plugin
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
	Services []Service
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
func (p *Parser) Build() (*ManbaState, error) {
	var state ManbaState
	ings := p.store.ListIngresses()
	// parse ingress rules
	parsedInfo, err := p.parseIngressRules(ings)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing ingress rules")
	}

	// populate Kubernetes Service
	for key, service := range parsedInfo.ServiceNameToServices {
		k8sSvc, err := p.store.GetService(service.Namespace, service.Backend.ServiceName)
		if err != nil {
			glog.Errorf("getting service: %v", err)
		}
		if k8sSvc != nil {
			service.K8SService = *k8sSvc
		}

		parsedInfo.ServiceNameToServices[key] = service
	}

	for _, service := range parsedInfo.ServiceNameToServices {
		svcKey := service.Namespace + "/" + service.Backend.ServiceName

		service.Servers, err = p.getServiceEndpoints(service.K8SService, service.Backend.ServicePort.String())
		if err != nil {
			glog.Errorf("error getting endpoints for '%v' service: %v",
				svcKey, err)
		}

		service.Servers, err = p.fillServers(service)
		if err != nil {
			glog.Errorf("error fill servers for '%v' service: %v",
				svcKey, err)
		}

		state.Services = append(state.Services, service)
	}

	for _, svc := range state.Services {
		state.Servers = append(state.Servers, svc.Servers...)
	}

	if err != nil {
		return nil, errors.Wrap(err, "building servers")
	}

	// merge ManbaIngress with APIS and Routings
	err = p.fillOverrides(state)
	if err != nil {
		return nil, errors.Wrap(err, "overriding ManbaIngress values")
	}

	// fill api before routing
	state.APIs = p.fillAPIs(state)

	state.Routings = p.fillRoutings(state)
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
						Cluster: metapb.Cluster{
							Name: fmt.Sprintf("%s.%s.%s.svc", rule.Backend.ServiceName, ingress.Namespace, rule.Backend.ServicePort.String()),
						},
						Namespace: ingress.Namespace,
						Backend:   rule.Backend,
					}
				}

				proxy := Proxy{
					ClusterName: service.GetName(),
				}
				i.Proxies = append(i.Proxies, proxy)

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
				Cluster: metapb.Cluster{
					Name: serviceName,
				},
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

		proxy := Proxy{
			ClusterName: service.GetName(),
		}
		i.Proxies = append(i.Proxies, proxy)

		service.APIs = append(service.APIs, i)
		serviceNameToServices[serviceName] = service
	}

	return &parsedIngressRules{
		ServiceNameToServices: serviceNameToServices,
	}, nil
}

func (p *Parser) fillServersFromPods(pods []corev1.Pod, svrs []Server) ([]Server, error) {
	// key: pod ip
	var podMap = make(map[string]corev1.Pod, len(pods))
	for _, pod := range pods {
		podMap[pod.Status.PodIP] = pod
	}
	var servers []Server
	for _, svr := range svrs {
		eip := strings.Split(svr.GetAddr(), ":")[0]
		if p, ok := podMap[eip]; ok {
			anns := p.GetAnnotations()
			svr.MaxQPS = annotations.ExtractMaxQPS(anns)
			svr.CircuitBreaker = annotations.ExtractCircuitBreaker(anns)
		}
		servers = append(servers, svr)
	}
	return servers, nil
}

func (p *Parser) fillOverrides(state ManbaState) error {
	for i := 0; i < len(state.Services); i++ {
		// Services
		anns := state.Services[i].K8SService.Annotations
		manbaIngress, err := p.getManbaIngressForService(state.Services[i].K8SService)
		if err != nil {
			glog.Errorf("error getting manbaIngress %v", err)
		}

		overrideService(&state.Services[i], manbaIngress, anns)

		// Routes
		for j := 0; j < len(state.Services[i].APIs); j++ {
			manbaIngress, err := p.getManbaIngressFromIngress(&state.Services[i].APIs[j].Ingress)
			if err != nil {
				glog.Errorf("error getting manbaIngress %v", err)
			}
			overrideAPI(&state.Services[i].APIs[j], manbaIngress)
		}
	}
	return nil
}

func (p *Parser) fillClusterByManbaIngress(service *Service, manbaIngress *configurationv1beta1.ManbaIngress) {
	if manbaIngress == nil || manbaIngress.Proxy == nil {
		return
	}
}

func (p *Parser) fillServers(services ...Service) ([]Server, error) {
	var servers []Server
	for _, svc := range services {
		pods, err := p.store.GetPodsForService(svc.K8SService.Namespace, svc.K8SService.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "get pods for service failed, namespace: <%s>, name: <%s>", svc.K8SService.Namespace, svc.K8SService.Name)
		}
		svrs, err := p.fillServersFromPods(pods, svc.Servers)
		if err != nil {
			return nil, errors.Wrapf(err, "fill pods from service failed, namespace: <%s>, name: <%s>", svc.K8SService.Namespace, svc.K8SService.Name)
		}
		servers = append(servers, svrs...)
	}
	return servers, nil
}

func (p *Parser) fillAPIs(state ManbaState) []API {
	var res []API
	for _, cls := range state.Services {
		res = append(res, cls.APIs...)
	}
	return res
}

func (p *Parser) fillRoutings(state ManbaState) []Routing {
	var res []Routing
	for _, api := range state.APIs {
		res = append(res, api.Routings...)
	}
	return res
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

func overrideService(service *Service, manbaIngress *configurationv1beta1.ManbaIngress, anns map[string]string) {
	if service == nil {
		return
	}

	overrideServiceByManbaIngress(service, manbaIngress)
	overrideServiceByAnnotation(service, anns)
}

func overrideServer(server *Server, pods []corev1.Pod) {
	for _, pod := range pods {
		if strings.HasPrefix(server.Addr, pod.Status.PodIP+":") {
			anns := pod.GetAnnotations()
			server.MaxQPS = annotations.ExtractMaxQPS(anns)
			server.CircuitBreaker = annotations.ExtractCircuitBreaker(anns)
			return
		}
	}

	glog.Warningf("failed find pod by server addr '%s'", server.Addr)
}

// overrideServiceByAnnotation sets the Service protocol via annotation
func overrideServiceByAnnotation(service *Service, anns map[string]string) {
	service.Cluster.LoadBalance = annotations.ExtractLoadBalancer(anns)
}

// overrideServiceByManbaIngress sets Service fields by ManbaIngress
func overrideServiceByManbaIngress(service *Service, manbaIngress *configurationv1beta1.ManbaIngress) {
	if manbaIngress == nil || manbaIngress.Proxy == nil {
		return
	}

	// s := manbaIngress.Proxy
	p := manbaIngress.Proxy

	for idx, api := range service.APIs {
		// match
		if api.URLPattern == p.URLPattern && api.Domain == p.Domain && api.Method == p.Method {
			api.API.Nodes = append(api.API.Nodes, p.DispatchNode)
			service.APIs[idx] = api
		}
	}
}

// overrideAPI  sets Route fields by ManbaIngress first, then by annotation
func overrideAPI(api *API, manbaIngress *configurationv1beta1.ManbaIngress) {
	if api == nil {
		return
	}

	overrideAPIByManbaIngress(api, manbaIngress)
	overrideAPIByAnnotation(api, api.Ingress.GetAnnotations())

}

// overrideAPIByManbaIngress sets API fields by KongIngress
func overrideAPIByManbaIngress(api *API, manbaIngress *configurationv1beta1.ManbaIngress) {
	if manbaIngress == nil || manbaIngress.API == nil {
		return
	}

	a := manbaIngress.API

	api.Status = a.Status

	// TODO: copy
	api.IPAccessControl = a.IPAccessControl
	api.DefaultValue = a.DefaultValue
	api.Perms = a.Perms
	api.AuthFilter = a.AuthFilter
	api.RenderTemplate = a.RenderTemplate
	api.UseDefault = a.UseDefault
	api.MatchRule = a.MatchRule
	api.Position = a.Position
	api.Tags = a.Tags
	api.WebSocketOptions = a.WebSocketOptions

}

// overrideAPIByAnnotation sets API protocols via annotation
func overrideAPIByAnnotation(api *API, anns map[string]string) {

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

func (p *Parser) getPodsFromService(service Service) ([]corev1.Pod, error) {
	return p.store.GetPodsForService(service.Namespace, service.K8SService.Name)
}
