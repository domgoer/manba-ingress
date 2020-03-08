module github.com/domgoer/manba-ingress

go 1.13

replace k8s.io/client-go v11.0.0+incompatible => k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90

require (
	github.com/eapache/channels v1.1.0
	github.com/eapache/queue v1.1.0 // indirect
	github.com/fagongzi/gateway v2.5.1+incompatible
	github.com/fagongzi/grpcx v1.1.0 // indirect
	github.com/fagongzi/log v0.0.0-20191122063922-293b75312445 // indirect
	github.com/fagongzi/manba v2.5.1+incompatible
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/hashicorp/go-memdb v1.1.0
	github.com/hbagdi/deck v1.0.2
	github.com/hbagdi/go-kong v0.11.0
	github.com/kong/kubernetes-ingress-controller v0.0.5 // indirect
	github.com/labstack/echo v3.3.10+incompatible // indirect
	github.com/labstack/gommon v0.3.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.2
	github.com/valyala/fasttemplate v1.1.0 // indirect
	github.com/yudai/gojsondiff v1.0.0
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/pool.v3 v3.1.1
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20200124190032-861946025e34 // indirect
	sigs.k8s.io/controller-runtime v0.5.0
)
