package utils

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ParseNameNS parses a string searching a namespace and name
func ParseNameNS(input string) (string, string, error) {
	nsName := strings.Split(input, "/")
	if len(nsName) != 2 {
		return "", "", fmt.Errorf("invalid format (namespace/name) found in '%v'", input)
	}

	return nsName[0], nsName[1], nil
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
