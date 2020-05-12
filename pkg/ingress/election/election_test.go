package election

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/leaderelection"
)

func TestNewElection(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	_, err := NewElection(Config{
		ElectionID:        "test",
		ResourceName:      "configmap",
		ResourceNamespace: "default",
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(context context.Context) {},
			OnStoppedLeading: func() {},
			OnNewLeader:      func(identity string) {},
		},
	}, fakeClient)

	assert.Nil(t, err)
}
