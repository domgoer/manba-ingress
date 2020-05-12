package k8s

import (
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PodInfo contains runtime information about the pod running the Ingres controller
type PodInfo struct {
	Name      string
	Namespace string
	NodeIP    string
	// Labels selectors of the running pod
	// This is used to search for other Ingress controller pods
	Labels map[string]string
}

// GetPodDetails returns runtime information about the pod:
// name, namespace and IP of the node where it is running
func GetPodDetails(kubeClient kubernetes.Interface) (*PodInfo, error) {
	podName := os.Getenv("POD_NAME")
	podNs := os.Getenv("POD_NAMESPACE")

	if podName == "" || podNs == "" {
		return nil, fmt.Errorf("unable to get POD information (missing POD_NAME or POD_NAMESPACE environment variable")
	}

	pod, _ := kubeClient.CoreV1().Pods(podNs).Get(podName, metav1.GetOptions{})
	if pod == nil {
		return nil, fmt.Errorf("unable to get POD information")
	}

	return &PodInfo{
		Name:      podName,
		Namespace: podNs,
		NodeIP:    GetNodeIPOrName(kubeClient, pod.Spec.NodeName),
		Labels:    pod.GetLabels(),
	}, nil
}

// GetNodeIPOrName returns the IP address or the name of a node in the cluster
func GetNodeIPOrName(kubeClient kubernetes.Interface, name string) string {
	node, err := kubeClient.CoreV1().Nodes().Get(name, metav1.GetOptions{})
	if err != nil {
		return ""
	}

	ip := ""

	for _, address := range node.Status.Addresses {
		if address.Type == corev1.NodeExternalIP {
			if address.Address != "" {
				ip = address.Address
				break
			}
		}
	}

	// Report the external IP address of the node
	if ip != "" {
		return ip
	}

	for _, address := range node.Status.Addresses {
		if address.Type == corev1.NodeInternalIP {
			if address.Address != "" {
				ip = address.Address
				break
			}
		}
	}

	return ip
}
