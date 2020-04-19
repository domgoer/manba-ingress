package v1beta1

import (
	"encoding/json"
	"github.com/golang/glog"
)

// DeepCopyInto ...
func (in *ManbaHTTPRule) DeepCopyInto(out *ManbaHTTPRule) {
	deepcopy(in, out)
}

// DeepCopyInto ...
func (in *TrafficPolicy) DeepCopyInto(out *TrafficPolicy) {
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
