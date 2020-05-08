package parser

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"

	configurationv1beta1 "github.com/domgoer/manba-ingress/pkg/apis/configuration/v1beta1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"

	"github.com/domgoer/manba-ingress/pkg/ingress/store"
	"github.com/domgoer/manba-ingress/pkg/utils"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
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
	Servers []*Server

	APIs []*API

	Namespace string
	Backend   configurationv1beta1.ManbaCluster
}

// Cluster containers k8s service and manba cluster
type Cluster struct {
	metapb.Cluster

	Servers   []*Server
	Port      string
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
	// Proxies key: clusterName, value: Proxy
	Proxies  map[string]Proxy
	HTTPRule configurationv1beta1.ManbaHTTPRule
	Routings []Routing
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
// Manba configuration.
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
				state.APIs = append(state.APIs, *api)
			}

			for _, routing := range api.Routings {
				if check(routing.Name) {
					state.Routings = append(state.Routings, routing)
				}
			}
		}
		service.Cluster.Servers = service.Servers
		if check(service.Cluster.Name) {
			state.Clusters = append(state.Clusters, *service.Cluster)
		}

		for _, server := range service.Servers {
			if check(server.Addr) {
				for _, svr := range service.Servers {
					state.Servers = append(state.Servers, *svr)
				}
			}
		}
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

		var apis []*API

		for j, rule := range ingressSpec.HTTP {
			base := API{
				API:       metapb.API{},
				Namespace: ingress.Namespace,
				HTTPRule:  rule,
			}

			base.fromManbaHTTPRule(&rule)

			for k, match := range rule.Match {

				for g, rule := range match.Rules {

					api := base
					api.Name = fmt.Sprintf("%s.%s.%d%d%d%d", ingress.Namespace, ingress.Name, i, j, k, g)
					api.Domain = match.Host
					api.MatchRule = metapb.MatchRule(metapb.MatchRule_value[rule.MatchType])
					api.Position = uint32(g + 1)
					api.Status = metapb.Up

					urlPattern := rule.URI.Pattern
					if urlPattern == "" {
						urlPattern = "/"
					}
					api.URLPattern = urlPattern

					if rule.Method != nil {
						api.Method = *rule.Method
					} else {
						api.Method = "*"
					}
					_, _, err := p.getTLS(api.Domain, ingress.Namespace, ingressSpec.TLS)
					if err != nil {
						glog.Errorf("getting secret failed, err: %v", err)
					} else {
						// TODO: set tls
					}

					// append to list
					apis = append(apis, &api)

				}

			}

			for _, route := range rule.Route {

				cls := route.Cluster

				serviceName := fmt.Sprintf("%s.%s.%s.%s.svc", ingress.Namespace, cls.Name, cls.Subset, cls.Port.String())

				service, ok := serviceNameToServices[serviceName]
				if !ok {
					cluster, err := p.store.GetManbaCluster(ingress.Namespace, cls.Name)
					if err != nil {
						glog.Errorf("getting manba cluster: %v", err)
						continue
					}
					subSet, err := p.getClusterSubset(cluster.Spec.Subsets, cls.Subset)
					if err != nil {
						glog.Errorf("getting manba subset: %v", err)
						continue
					}

					if subSet.TrafficPolicy == nil {
						subSet.TrafficPolicy = cluster.Spec.TrafficPolicy
					}

					service = &Service{
						Cluster: &Cluster{
							Cluster: metapb.Cluster{
								Name: serviceName,
							},
							Port:      cls.Port.String(),
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

func (p *Parser) getTLS(host, namespace string, tls networkingv1beta1.IngressTLS) (certData []byte, keyData []byte, err error) {
	for _, h := range tls.Hosts {
		if host == h {
			var secret *corev1.Secret
			secret, err = p.store.GetSecret(namespace, tls.SecretName)
			if err != nil {
				return
			}

			certData = secret.Data["tls.crt"]
			keyData = secret.Data["tls.key"]
			return
		}
	}
	return
}

func (p *Parser) fillOverride(service *Service) error {
	cls := service.Cluster
	namespace := cls.Namespace

	// fill servers
	servers, err := p.getServiceEndpoints(cls.K8SSbuSet, namespace, cls.Port)
	if err != nil {
		return err
	}

	qps := math.MaxInt32
	traffic := cls.K8SSbuSet.TrafficPolicy
	if len(servers) != 0 && traffic != nil {
		qps = int(traffic.MaxQPS) / len(servers)

		for _, svr := range servers {
			svr.CircuitBreaker = traffic.CircuitBreaker
			svr.MaxQPS = int64(qps)
		}

		// fill cluster
		if traffic.LoadBalancer != nil {
			service.Cluster.LoadBalance = metapb.LoadBalance(metapb.LoadBalance_value[*traffic.LoadBalancer])
		}

	}

	service.Servers = servers

	// fill api nodes
	for _, api := range service.APIs {
		rule := api.HTTPRule
		for _, r := range api.HTTPRule.Route {
			proxy := Proxy{
				ClusterName: service.Cluster.Name,
			}

			proxy.fromManbaHTTPRule(&rule)
			proxy.fromManbaHTTPRoute(&r)

			// ini map
			if api.Proxies == nil {
				api.Proxies = make(map[string]Proxy)
			}
			if _, ok := api.Proxies[proxy.ClusterName]; !ok {
				api.Proxies[proxy.ClusterName] = proxy
			}
		}

		parseRouting := func(m configurationv1beta1.ManbaHTTPRouting, override func(Routing) Routing) Routing {
			var rate int32 = 100
			if m.Rate != nil {
				rate = *m.Rate
			}
			return override(Routing{
				APIName:     api.Name,
				ClusterName: fmt.Sprintf("%s.%s.%s.%s.svc", api.Namespace, m.Cluster.Name, m.Cluster.Subset, m.Cluster.Port.String()),
				Routing: metapb.Routing{
					TrafficRate: rate,
					Status:      metapb.Up,
					Conditions:  m.Conditions,
				},
			})
		}

		for i, mirror := range api.HTTPRule.Mirror {
			api.Routings = append(api.Routings, parseRouting(mirror, func(r Routing) Routing {
				r.Name = fmt.Sprintf("%s.mirror.%d", api.Name, i)
				r.Strategy = metapb.Copy
				return r
			}))
		}
		for i, split := range api.HTTPRule.Split {
			api.Routings = append(api.Routings, parseRouting(split, func(r Routing) Routing {
				r.Name = fmt.Sprintf("%s.split.%d", api.Name, i)
				r.Strategy = metapb.Split
				return r
			}))
		}
	}
	return nil
}

func (p *Parser) getServiceEndpoints(subset configurationv1beta1.ManbaClusterSubSet, namespace string,
	backendPort string) ([]*Server, error) {
	var servers []*Server
	var endpoints []utils.Endpoint
	svcs, err := p.store.ListServices(namespace, subset.Labels)
	if err != nil {
		return nil, err
	}

	for _, svc := range svcs {
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

		endpoints = getEndpoints(svc, &servicePort,
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
			servers = append(servers, &server)
		}

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
	if meta.DefaultValue != nil {
		meta.UseDefault = true
	}
	meta.IPAccessControl = rule.IPAccessControl
	meta.RenderTemplate = rule.RenderTemplate
	if rule.AuthFilter != nil {
		meta.AuthFilter = *rule.AuthFilter
	}
	// meta.MaxQPS = path.MaxQPS
	// meta.CircuitBreaker = path.CircuitBreaker
	// meta.RateLimitOption = path.RateLimitOption

	a.API = meta

}

func (p *Proxy) fromManbaHTTPRule(rule *configurationv1beta1.ManbaHTTPRule) {
	node := p.DispatchNode
	if rule.Rewrite != nil {
		node.URLRewrite = rule.Rewrite.URI
	}
	node.RetryStrategy = rule.Retry
	p.DispatchNode = node
}

func (p *Proxy) fromManbaHTTPRoute(route *configurationv1beta1.ManbaHTTPRoute) {
	node := p.DispatchNode
	node.WriteTimeout = route.WriteTimeout
	node.ReadTimeout = route.ReadTimeout
	node.DefaultValue = route.DefaultValue
	if node.DefaultValue != nil {
		node.UseDefault = true
	}
	node.Cache = route.Cache
	node.AttrName = route.AttrName
	node.Validations = route.Match.ToManbaValidations()
	node.BatchIndex = route.BatchIndex

	if route.Rewrite != nil && route.Rewrite.URI != "" {
		node.URLRewrite = route.Rewrite.URI
	}

	p.DispatchNode = node
}
