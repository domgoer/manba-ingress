package controller

import (
	"log"

	manbaClient "github.com/fagongzi/gateway/pkg/client"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"os"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes/fake"

	"github.com/domgoer/manba-ingress/pkg/ingress/store"
	"github.com/eapache/channels"
)

var manbaController *ManbaController

func init() {
	os.Setenv("POD_NAME", "test")
	os.Setenv("POD_NAMESPACE", "test")

	mc, err := manbaClient.NewClient(time.Second, "127.0.0.1:2379")
	if err != nil {
		log.Fatal(err)
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
	}
	fakeClient := fake.NewSimpleClientset(pod)
	store2, _ := store.NewFakeStore([]runtime.Object{pod}, nil)
	manbaController, _ = NewManbaController(Config{
		Manba: Manba{
			Client: mc,
		},
		ElectionID:           "test",
		KubeClient:           fakeClient,
		IngressClass:         "test",
		ResyncPeriod:         time.Second * 10,
		SyncRateLimit:        0.3,
		UpdateStatus:         false,
		PublishService:       "",
		PublishStatusAddress: "",
		Concurrency:          10,
	}, channels.NewRingChannel(1024), store2)
}

func TestManbaController_Start(t *testing.T) {
	go manbaController.Start()
	time.Sleep(time.Second * 3)
	err := manbaController.Stop()
	assert.Nil(t, err)
}
