package v1beta1

import (
	"encoding/json"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/golang/glog"
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
}

// ManbaIngressList is a list of ManbaIngress
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ManbaIngressList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Items             []ManbaIngress `json:"items,omitempty"`
}

// ManbaIngressSpec api list
type ManbaIngressSpec struct {
	Rules []ManbaIngressRule             `json:"rules,omitempty"`
	TLS   []networkingv1beta1.IngressTLS `json:"tls,omitempty"`
}

// ManbaIngressRule implements manba api
type ManbaIngressRule struct {
	Host string `json:"host,omitempty"`
	// default value
	Method          *string                 `json:"method,omitempty"`
	Status          *string                 `json:"status,omitempty"`
	IPAccessControl *metapb.IPAccessControl `json:"ipAccessControl,omitempty"`
	DefaultValue    *metapb.HTTPResult      `json:"defaultValue,omitempty"`
	UseDefault      *bool                   `json:"useDefault,omitempty"`
	CircuitBreaker  *metapb.CircuitBreaker  `json:"circuitBreaker,omitempty"`
	AuthFilter      *string                 `json:"authFilter,omitempty"`
	MaxQPS          int64                   `json:"maxQPS,omitempty"`
	RateLimitOption *string                 `json:"rateLimitOption,omitempty"`
	Backend         *ManbaIngressBackend    `json:"backend,omitempty"`

	Paths []ManbaIngressPath `json:"paths,omitempty"`
}

// DeepCopyInto ...
func (in *ManbaIngressRule) DeepCopyInto(out *ManbaIngressRule) {
	deepcopy(in, out)
}

// ManbaIngressPath contains manba api and dispatchNodes and routing
type ManbaIngressPath struct {
	Method           string                   `json:"method,omitempty"`
	URLPattern       string                   `json:"urlPattern,omitempty"`
	Status           string                   `json:"status,omitempty"`
	DefaultValue     *metapb.HTTPResult       `json:"defaultValue,omitempty"`
	Perms            []string                 `json:"perms,omitempty"`
	AuthFilter       string                   `json:"authFilter,omitempty"`
	RenderTemplate   *metapb.RenderTemplate   `json:"renderTemplate,omitempty"`
	UseDefault       bool                     `json:"useDefault,omitempty"`
	MatchRule        string                   `json:"matchRule,omitempty"`
	Position         uint32                   `json:"position,omitempty"`
	Tags             []*metapb.PairValue      `json:"tags,omitempty"`
	WebSocketOptions *metapb.WebSocketOptions `json:"webSocketOptions,omitempty"`
	MaxQPS           int64                    `json:"maxQPS,omitempty"`
	CircuitBreaker   *metapb.CircuitBreaker   `json:"circuitBreaker,omitempty"`
	RateLimitOption  string                   `json:"rateLimitOption,omitempty"`
	Backends         []ManbaIngressBackend    `json:"backends,omitempty"`
	Route            *ManbaIngressRoute       `json:"route,omitempty"`
}

// DeepCopyInto ...
func (in *ManbaIngressPath) DeepCopyInto(out *ManbaIngressPath) {
	deepcopy(in, out)
}

// ManbaIngressBackend is dispatchNodes config in manba
type ManbaIngressBackend struct {
	ServiceName   string                `json:"serviceName,omitempty"`
	ServicePort   int32                 `json:"servicePort,omitempty"`
	URLRewrite    string                `json:"urlRewrite,omitempty"`
	AttrName      string                `json:"attrName,omitempty"`
	Validations   []*metapb.Validation  `json:"validations,omitempty"`
	Cache         *metapb.Cache         `json:"cache,omitempty"`
	DefaultValue  *metapb.HTTPResult    `json:"defaultValue,omitempty"`
	UseDefault    bool                  `json:"useDefault,omitempty"`
	BatchIndex    int32                 `json:"batchIndex,omitempty"`
	RetryStrategy *metapb.RetryStrategy `json:"retryStrategy,omitempty"`
	WriteTimeout  int64                 `json:"writeTimeout,omitempty"`
	ReadTimeout   int64                 `json:"readTimeout,omitempty"`
	HostType      string                `json:"hostType,omitempty"`
	CustomHost    string                `json:"customHost,omitempty"`
}

// DeepCopyInto ...
func (in *ManbaIngressBackend) DeepCopyInto(out *ManbaIngressBackend) {
	deepcopy(in, out)
}

// ManbaIngressRoute is manba routing
type ManbaIngressRoute struct {
	Status      string             `json:"status,omitempty"`
	Conditions  []metapb.Condition `json:"conditions,omitempty"`
	Strategy    string             `json:"strategy,omitempty"`
	TrafficRate int32              `json:"trafficRate,omitempty"`
}

// DeepCopyInto ...
func (in *ManbaIngressRoute) DeepCopyInto(out *ManbaIngressRoute) {
	deepcopy(in, out)
}

// DeepCopyInto deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ManbaIngress) DeepCopyInto(out *ManbaIngress) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Spec != nil {
		in, out := in.Spec, out.Spec
		deepcopy(in, out)
	}
}

func deepcopy(in, out interface{}) {
	b, err := json.Marshal(in)
	if err != nil {
		glog.Errorf("unexpected error copying configuration: %v", err)
	}
	err = json.Unmarshal(b, out)
	if err != nil {
		glog.Errorf("unexpected error copying configuration: %v", err)
	}
}
