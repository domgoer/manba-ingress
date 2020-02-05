package main

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	apiv1 "k8s.io/api/core/v1"
)

type Config struct {
	// Admission controller server properties
	AdmissionWebhookListen   string
	AdmissionWebhookCertPath string
	AdmissionWebhookKeyPath  string

	// Manba connection details
	ManbaAPIServer        string
	ManbaAPIServerTimeout time.Duration
	ManbaWorkspace        string

	// Resource filtering
	WatchNamespace string
	IngressClass   string
	ElectionID     string

	// Rutnime behavior
	SyncPeriod    time.Duration
	SyncRateLimit float32

	// k8s connection details
	APIServerHost      string
	KubeConfigFilePath string
}

func flagSet() *pflag.FlagSet {
	flags := pflag.NewFlagSet("", pflag.ExitOnError)
	// Admission controller server properties
	flags.String("admission-webhook-listen", "off",
		`The address to start admission controller on (ip:port).
Setting it to 'off' disables the admission controller.`)
	flags.String("admission-webhook-cert-file", "/admission-webhook/tls.crt",
		`Path to the PEM-encoded certificate file for
TLS handshake`)
	flags.String("admission-webhook-key-file", "/admission-webhook/tls.key",
		`Path to the PEM-encoded private key file for
    TLS handshake`)

	flags.StringP("manba-api-server-addr", "s", "", "The address of the Manba API Server to connect to in the format of protocol://address:port, e.g. grpc://localhost:9092")
	flags.Duration("manba-api-server-timeout", time.Second*10, "The timeout of connection to Manba API Server")
	flags.String("manba-workspace", "",
		"Workspace in Kong Enterprise to be configured")

	// Resource filtering
	flags.String("watch-namespace", apiv1.NamespaceAll,
		`Namespace to watch for Ingress. Default is to watch all namespaces`)
	flags.String("ingress-class", "",
		`Name of the ingress class to route through this controller.`)
	flags.String("election-id", "ingress-controller-leader",
		`Election id to use for status update.`)
	// Rutnime behavior
	flags.Duration("sync-period", 600*time.Second,
		`Relist and confirm cloud resources this often.`)
	flags.Float32("sync-rate-limit", 0.3,
		`Define the sync frequency upper limit`)

	// k8s connection details
	flags.String("apiserver-host", "",
		`The address of the Kubernetes Apiserver to connect to in the format of 
protocol://address:port, e.g., "http://localhost:8080.
If not specified, the assumption is that the binary runs inside a 
Kubernetes cluster and local discovery is attempted.`)
	flags.String("kubeconfig", "", "Path to kubeconfig file with authorization and master location information.")
	return flags
}

func parseFlags() (cfg Config, err error) {
	flagSet := flagSet()

	// glog
	flag.Set("logtostderr", "true")

	flagSet.AddGoFlagSet(flag.CommandLine)
	flagSet.Parse(os.Args)

	// Workaround for this issue:
	// https://github.com/kubernetes/kubernetes/issues/17162
	flag.CommandLine.Parse([]string{})

	viper.SetEnvPrefix("CONTROLLER")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.BindPFlags(flagSet)

	for key, value := range viper.AllSettings() {
		glog.V(2).Infof("FLAG: --%s=%q", key, value)
	}

	// Admission controller server properties
	cfg.AdmissionWebhookListen = viper.GetString("admission-webhook-listen")
	cfg.AdmissionWebhookCertPath =
		viper.GetString("admission-webhook-cert-file")
	cfg.AdmissionWebhookKeyPath =
		viper.GetString("admission-webhook-key-file")

		// manba detail
	cfg.ManbaAPIServer = viper.GetString("manba-api-server-addr")
	cfg.ManbaWorkspace = viper.GetString("manba-workspace")
	cfg.ManbaAPIServerTimeout = viper.GetDuration("manba-api-server-timeout")

	// Resource filtering
	cfg.WatchNamespace = viper.GetString("watch-namespace")
	cfg.IngressClass = viper.GetString("ingress-class")
	cfg.ElectionID = viper.GetString("election-id")

	// Rutnime behavior
	cfg.SyncPeriod = viper.GetDuration("sync-period")
	cfg.SyncRateLimit = (float32)(viper.GetFloat64("sync-rate-limit"))

	// k8s connection details
	cfg.APIServerHost = viper.GetString("apiserver-host")
	cfg.KubeConfigFilePath = viper.GetString("kubeconfig")

	return
}
