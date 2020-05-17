package k8s

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	k8sFake "k8s.io/client-go/kubernetes/fake"
)

func TestGetNodeIPOrName(t *testing.T) {
	node1 := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node-1",
		},
		Status: corev1.NodeStatus{
			Addresses: []corev1.NodeAddress{
				{
					Type:    corev1.NodeExternalIP,
					Address: "1.1.1.1",
				},
			},
		},
	}
	node2 := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node-2",
		},
		Status: corev1.NodeStatus{
			Addresses: []corev1.NodeAddress{
				{
					Type:    corev1.NodeInternalIP,
					Address: "1.1.1.2",
				},
			},
		},
	}
	fakeClient := k8sFake.NewSimpleClientset(node1, node2)
	address := GetNodeIPOrName(fakeClient, "test-node-1")
	assert.Equal(t, address, "1.1.1.1")
	address = GetNodeIPOrName(fakeClient, "test-node-2")
	assert.Equal(t, address, "1.1.1.2")
	address = GetNodeIPOrName(fakeClient, "do-not-exist")
	assert.Equal(t, address, "")
}

func TestGetPodDetails(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
			Labels: map[string]string{
				"app": "foo",
			},
		},
		Spec: corev1.PodSpec{
			NodeName: "test-node",
		},
	}
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
		Status: corev1.NodeStatus{
			Addresses: []corev1.NodeAddress{
				{
					Type:    corev1.NodeExternalIP,
					Address: "1.1.1.1",
				},
			},
		},
	}

	os.Setenv("POD_NAMESPACE", "default")
	os.Setenv("POD_NAME", "foo")

	fakeClient := k8sFake.NewSimpleClientset(pod, node)
	podInfo, err := GetPodDetails(fakeClient)
	assert.Nil(t, err)
	assert.Equal(t, podInfo, &PodInfo{
		Name:      "foo",
		Namespace: "default",
		NodeIP:    "1.1.1.1",
		Labels: map[string]string{
			"app": "foo",
		},
	})

}
