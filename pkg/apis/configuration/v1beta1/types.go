package v1beta1

import (
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ManbaIngress is a top-level type. A client is created for it.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ManbaIngress struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec *ManbaIngressSpec `json:"spec,omitempty"`

	// Status *ManbaIngressStatus `json:"status,omitempty"`
	Status networkingv1beta1.IngressStatus `json:"status,omitempty"`
}

// ManbaIngressList is a list of ManbaIngress
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ManbaIngressList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManbaIngress `json:"items,omitempty"`
}

// ManbaIngressSpec api list
type ManbaIngressSpec struct {
	Hosts []string        `json:"hosts,omitempty"`
	HTTP  []ManbaHTTPRule `json:"https,omitempty"`
	// TLS   []networkingv1beta1.IngressTLS `json:"tls,omitempty"`
}

// ManbaHTTPRule implements manba api
type ManbaHTTPRule struct {
	Match           []ManbaHTTPMatch        `json:"match,omitempty"`
	Rewrite         *ManbaHTTPURIRewrite    `json:"rewrite,omtiempty"`
	IPAccessControl *metapb.IPAccessControl `json:"ipAccessControl,omitempty"`
	Retry           *metapb.RetryStrategy   `json:"retries,omitempty"`
	DefaultValue    *metapb.HTTPResult      `json:"defaultValue,omitempty"`
	RenderTemplate  *metapb.RenderTemplate  `json:"renderTemplate,omitempty"`
	AuthFilter      *string                 `json:"authFilter,omitempty"`
	TrafficPolicy   *TrafficPolicy          `json:"trafficPolicy,omitempty"`
	Route           []ManbaHTTPRoute        `json:"paths,omitempty"`
}

type ManbaHTTPMatch struct {
	URI    ManbaHTTPURIMatch `json:"uri,omitempty"`
	Method *string           `json:"method,omitempty"`
}

type ManbaHTTPURIMatch struct {
	Pattern string `json:"pattern,omitempty"`
}

type ManbaHTTPURIRewrite struct {
	URI string `json:"uri,omitempty"`
}

func (m *ManbaHTTPURIRewrite) GetURI() string {
	if m == nil {
		return ""
	}
	return m.URI
}

type ManbaHTTPRoute struct {
	Name   string `json:"name"`
	Subset string `json:"subset"`
	Port   uint32 `json:"port"`
}

// ManbaCluster is top level of manba cluster
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ManbaCluster struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ManbaClusterSpec `json:"spec,omitempty"`
}

// ManbaClusterList is a list of ManbaCluster
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ManbaClusterList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManbaCluster `json:"items,omitempty"`
}

// ManbaClusterSpec details of ManbaCluster
type ManbaClusterSpec struct {
	TrafficPolicy *TrafficPolicy       `json:"trafficPolicy,omitempty"`
	Subsets       []ManbaClusterSubSet `json:"subsets"`
}

// ManbaClusterSubSet represents service in k8s
type ManbaClusterSubSet struct {
	// Name of subset, like v1
	Name string `json:"name"`
	// Labels used to list service by labels
	Labels map[string]string `json:"labels,omitempty"`
	// TrafficPolicy for cluster, if cluster has 5 servers,
	// single server's maxQPS is trafficPolicy.MaxQPS/5
	TrafficPolicy *TrafficPolicy `json:"trafficPolicy,omitempty"`
}

type TrafficPolicy struct {
	LoadBalancer    *string                `json:"loadBalancer,omitempty"`
	MaxQPS          uint64                 `json:"maxQPS"`
	CircuitBreaker  *metapb.CircuitBreaker `json:"circuitBreaker,omitempty"`
	RateLimitOption *string                `json:"rateLimitOption,omitempty"`
}
