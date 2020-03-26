package v1beta1

import (
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/golang/glog"
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

	API     *metapb.API     `json:"api,omitempty"`
	Routing *metapb.Routing `json:"routing,omitempty"`
	Proxy   *Proxy          `json:"proxy,omitempty"`
}

// Proxy for cluster
type Proxy struct {
	*metapb.DispatchNode
	Domain     string `json:"domain,omitempty"`
	Method     string `json:"method,omitempty"`
	URLPattern string `json:"urlPattern,omitempty"`
}

// ManbaIngressList is a top-level list type. The client methods for
// lists are automatically created.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ManbaIngressList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	// +optional
	Items []ManbaIngress `json:"items"`
}

// DeepCopyInto deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ManbaIngress) DeepCopyInto(out *ManbaIngress) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.API != nil {
		in, out := in.API, out.API
		b, err := in.Marshal()
		if err != nil {
			glog.Errorf("unexpected error copying configuration: %v", err)
		}
		err = out.Unmarshal(b)
		if err != nil {
			glog.Errorf("unexpected error copying configuration: %v", err)
		}
	}
	if in.Routing != nil {
		in, out := in.Routing, out.Routing
		b, err := in.Marshal()
		if err != nil {
			glog.Errorf("unexpected error copying configuration: %v", err)
		}
		err = out.Unmarshal(b)
		if err != nil {
			glog.Errorf("unexpected error copying configuration: %v", err)
		}
	}
	if in.Proxy != nil {
		in.Proxy.DeepCopyInto(out.Proxy)
	}
	return
}

// DeepCopyInto deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Proxy) DeepCopyInto(out *Proxy) {
	out.URLPattern = in.URLPattern
	out.Method = in.Method
	out.Domain = in.Domain
	inDN, outDN := in.DispatchNode, out.DispatchNode
	b, err := inDN.Marshal()
	if err != nil {
		glog.Errorf("unexpected error copying configuration: %v", err)
	}
	err = outDN.Unmarshal(b)
	if err != nil {
		glog.Errorf("unexpected error copying configuration: %v", err)
	}
}
