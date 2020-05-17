package parser

import (
	"testing"

	"github.com/fagongzi/gateway/pkg/pb/metapb"

	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/api/networking/v1beta1"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	configurationv1beta1 "github.com/domgoer/manba-ingress/pkg/apis/configuration/v1beta1"
	"github.com/domgoer/manba-ingress/pkg/ingress/store"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestParser_Build(t *testing.T) {
	var method = "POST"
	var rate int32 = 20
	ingress := &configurationv1beta1.ManbaIngress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ing",
			Namespace: "default",
		},
		Spec: configurationv1beta1.ManbaIngressSpec{
			HTTP: []configurationv1beta1.ManbaHTTPRule{
				{
					Match: []configurationv1beta1.ManbaHTTPMatch{
						{
							Host: "test",
							Rules: []configurationv1beta1.ManbaHTTPMatchRule{
								{
									URI:       configurationv1beta1.ManbaHTTPURIMatch{Pattern: "/"},
									Method:    &method,
									MatchType: "all",
								},
							},
						},
					},
					Route: []configurationv1beta1.ManbaHTTPRoute{
						{
							Cluster: configurationv1beta1.ManbaHTTPRouteCluster{
								Name:   "test-cls",
								Subset: "v1",
								Port: intstr.IntOrString{
									IntVal: 8080,
								},
							},
						},
					},
					Mirror: []configurationv1beta1.ManbaHTTPRouting{
						{
							Cluster: configurationv1beta1.ManbaHTTPRouteCluster{
								Name:   "test-cls",
								Subset: "v1",
								Port: intstr.IntOrString{
									IntVal: 8080,
								},
							},
							Rate: &rate,
						},
					},
				},
			},
			TLS: v1beta1.IngressTLS{
				Hosts:      []string{"test"},
				SecretName: "test-secret",
			},
		},
	}
	cluster := &configurationv1beta1.ManbaCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cls",
			Namespace: "default",
		},
		Spec: configurationv1beta1.ManbaClusterSpec{
			Subsets: []configurationv1beta1.ManbaClusterSubSet{
				{
					Name: "v1",
					Labels: map[string]string{
						"app": "test",
					},
					TrafficPolicy: &configurationv1beta1.TrafficPolicy{
						MaxQPS: 500,
					},
				},
			},
		},
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-svc",
			Namespace: "default",
			Labels: map[string]string{
				"app": "test",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "test",
			},
			Ports: []corev1.ServicePort{
				{
					Name: "test-port",
					Port: 8080,
				},
			},
		},
	}
	endpoint := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-svc",
			Namespace: "default",
			Labels: map[string]string{
				"app": "test",
			},
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: []corev1.EndpointAddress{
					{
						IP: "1.1.1.1",
					},
					{
						IP: "1.1.1.2",
					},
				},
				Ports: []corev1.EndpointPort{
					{
						Name:     "test-port",
						Port:     8080,
						Protocol: corev1.ProtocolTCP,
					},
				},
			},
		},
	}

	want := &ManbaState{
		APIs: []API{
			{
				API: metapb.API{
					Name:       "default.test-ing.0000",
					URLPattern: "/",
					Method:     method,
					Domain:     "test",
					Status:     metapb.Up,
					Position:   1,
				},
				Namespace: "default",
				Proxies: map[string]Proxy{
					"default.test-cls.v1.8080.svc": {
						ClusterName:  "default.test-cls.v1.8080.svc",
						DispatchNode: metapb.DispatchNode{},
					},
				},
				HTTPRule: configurationv1beta1.ManbaHTTPRule{
					Match: []configurationv1beta1.ManbaHTTPMatch{
						{
							Host: "test",
							Rules: []configurationv1beta1.ManbaHTTPMatchRule{
								{
									URI: configurationv1beta1.ManbaHTTPURIMatch{
										Pattern: "/",
									},
									Method:    &method,
									MatchType: "all",
								},
							},
						},
					},
					Route: []configurationv1beta1.ManbaHTTPRoute{
						{
							Cluster: configurationv1beta1.ManbaHTTPRouteCluster{
								Name:   "test-cls",
								Subset: "v1",
								Port: intstr.IntOrString{
									IntVal: 8080,
								},
							},
						},
					},
					Mirror: []configurationv1beta1.ManbaHTTPRouting{
						{
							Cluster: configurationv1beta1.ManbaHTTPRouteCluster{
								Name:   "test-cls",
								Subset: "v1",
								Port: intstr.IntOrString{
									IntVal: 8080,
								},
							},
							Rate: &rate,
						},
					},
				},
				Routings: []Routing{
					{
						APIName:     "default.test-ing.0000",
						ClusterName: "default.test-cls.v1.8080.svc",
						Routing: metapb.Routing{
							Conditions:  nil,
							Strategy:    metapb.Copy,
							TrafficRate: rate,
							Status:      metapb.Up,
							Name:        "default.test-ing.0000.mirror.0",
						},
					},
				},
			},
		},
		Servers: []Server{
			{
				Server: metapb.Server{
					Addr:   "1.1.1.1:8080",
					MaxQPS: 250,
				},
			},
			{
				Server: metapb.Server{
					Addr:   "1.1.1.2:8080",
					MaxQPS: 250,
				},
			},
		},
		Clusters: []Cluster{
			{
				Cluster: metapb.Cluster{
					Name:        "default.test-cls.v1.8080.svc",
					LoadBalance: metapb.RoundRobin,
				},
				Port:      "8080",
				Namespace: "default",
				K8SSbuSet: configurationv1beta1.ManbaClusterSubSet{
					Name: "v1",
					Labels: map[string]string{
						"app": "test",
					},
					TrafficPolicy: &configurationv1beta1.TrafficPolicy{
						MaxQPS: 500,
					},
				},
				Servers: []*Server{
					{
						Server: metapb.Server{
							Addr:   "1.1.1.1:8080",
							MaxQPS: 250,
						},
					},
					{
						Server: metapb.Server{
							Addr:   "1.1.1.2:8080",
							MaxQPS: 250,
						},
					},
				},
			},
		},
		Routings: []Routing{
			{
				APIName:     "default.test-ing.0000",
				ClusterName: "default.test-cls.v1.8080.svc",
				Routing: metapb.Routing{
					Strategy:    metapb.Copy,
					TrafficRate: rate,
					Status:      metapb.Up,
					Name:        "default.test-ing.0000.mirror.0",
				},
			},
		},
		Plugins: nil,
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
	}
	fakeStore, err := store.NewFakeStore([]runtime.Object{service, endpoint, secret}, []runtime.Object{ingress, cluster})
	assert.Nil(t, err)

	parser := New(fakeStore)

	ms, err := parser.Build()
	assert.Nil(t, err)
	assert.Equal(t, ms, want)
}
