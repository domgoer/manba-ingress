package admission

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	configurationv1beta1 "github.com/domgoer/manba-ingress/pkg/apis/configuration/v1beta1"

	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	manbaIngressResource = metav1.GroupVersionResource{
		Group:    configurationv1beta1.SchemeGroupVersion.Group,
		Version:  configurationv1beta1.SchemeGroupVersion.Version,
		Resource: "manbaingresses",
	}

	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

// Server is an HTTP server that can validate ManbaIngress Controllers'
// Custom Resources using Kubernetes Admission Webhooks.
type Server struct {
	Validator ManbaValidator

	manager ctrl.Manager
}

// New returns a new server
func New(restConfig *rest.Config, listen, certDir string, validator ManbaValidator) (s *Server, err error) {
	s = new(Server)
	host, port, err := parseListen(listen)
	if err != nil {
		return nil, err
	}
	s.manager, err = ctrl.NewManager(restConfig, ctrl.Options{
		Port:    port,
		Host:    host,
		CertDir: certDir,
	})
	s.Validator = validator
	return s, err
}

// Handle parses AdmissionReview requests and responds back
// with the validation result of the entity.
func (s *Server) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Resource {
	case manbaIngressResource:
		ingress := new(configurationv1beta1.ManbaIngress)
		deserializer := codecs.UniversalDeserializer()
		_, _, err := deserializer.Decode(req.Object.Raw,
			nil, ingress)
		if err != nil {
			return webhook.Errored(http.StatusInternalServerError, err)
		}

		valid, msg, err := s.Validator.ValidateManbaIngress(ingress)
		if err != nil {
			return webhook.Errored(http.StatusInternalServerError, err)
		}
		if !valid {
			return webhook.Denied(msg)
		}
		return webhook.Allowed("The resource definition conforms to the specification")
	}
	return webhook.Allowed("unknown resource type")
}

// Start starts all registered Controllers and blocks until the Stop channel is closed.
// Returns an error if there is an error starting any controller.
func (s *Server) Start(stopCh <-chan struct{}) error {
	svr := s.manager.GetWebhookServer()
	svr.Register("/", &webhook.Admission{
		Handler: s,
	})
	return s.manager.Start(stopCh)
}

func parseListen(listen string) (host string, port int, err error) {
	strs := strings.Split(listen, ":")
	if len(strs) != 2 {
		err = fmt.Errorf("listen address must conform to <ip:port>")
		return
	}
	portStr := strs[1]
	port, atoiErr := strconv.Atoi(portStr)
	if atoiErr != nil {
		err = atoiErr
		return
	}
	return strs[0], port, nil
}
