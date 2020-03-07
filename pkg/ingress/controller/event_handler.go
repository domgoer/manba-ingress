package controller

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"

	"github.com/eapache/channels"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceEventHandler is "ingress.class" aware resource
// handler.
type ResourceEventHandler struct {
	IsValidIngresClass func(object metav1.Object) bool
	UpdateCh           *channels.RingChannel
}

// EventType type of event associated with an informer
type EventType string

const (
	// CreateEvent event associated with new objects in an informer
	CreateEvent EventType = "CREATE"
	// UpdateEvent event associated with an object update in an informer
	UpdateEvent EventType = "UPDATE"
	// DeleteEvent event associated when an object is removed from an informer
	DeleteEvent EventType = "DELETE"
	// ConfigurationEvent event associated when a controller configuration object is created or updated
	ConfigurationEvent EventType = "CONFIGURATION"
)

// Event holds the context of an event.
type Event struct {
	Type EventType
	Obj  interface{}
	Old  interface{}
}

// OnAdd is invoked whenever a resource is added.
func (r ResourceEventHandler) OnAdd(obj interface{}) {
	object, err := meta.Accessor(obj)
	if err != nil {
		return
	}
	if !r.IsValidIngresClass(object) {
		return
	}
	r.UpdateCh.In() <- Event{
		Type: CreateEvent,
		Obj:  obj,
	}
}

// OnUpdate is invoked whenever a resource is updated.
func (r ResourceEventHandler) OnUpdate(old, obj interface{}) {
	oldObj, err := meta.Accessor(old)
	if err != nil {
		return
	}
	curObj, err := meta.Accessor(obj)
	if err != nil {
		return
	}
	validOld := r.IsValidIngresClass(oldObj)
	validCur := r.IsValidIngresClass(curObj)

	if !validCur && !validOld {
		return
	}

	r.UpdateCh.In() <- Event{
		Type: UpdateEvent,
		Obj:  obj,
		Old:  old,
	}
}

// OnDelete is invoked whenever a resource is deleted
func (r ResourceEventHandler) OnDelete(obj interface{}) {
	object, err := meta.Accessor(obj)
	if err != nil {
		return
	}
	if !r.IsValidIngresClass(object) {
		return
	}

	r.UpdateCh.In() <- Event{
		Type: DeleteEvent,
		Obj:  obj,
	}
}

// EndpointsEventHandler handles create, update and delete events for
// endpoint resources in k8s.
// It is not ingress.class aware and the OnUpdate method filters out
// events with same set of endpoints.
type EndpointsEventHandler struct {
	UpdateCh *channels.RingChannel
}

// OnAdd is invoked whenever a resource is added.
func (reh EndpointsEventHandler) OnAdd(obj interface{}) {
	reh.UpdateCh.In() <- Event{
		Type: CreateEvent,
		Obj:  obj,
	}
}

// OnDelete is invoked whenever a resource is deleted.
func (reh EndpointsEventHandler) OnDelete(obj interface{}) {
	reh.UpdateCh.In() <- Event{
		Type: DeleteEvent,
		Obj:  obj,
	}
}

// OnUpdate is invoked whenever an Endpoint is changed.
// If the endpoints are same as before, an update is not sent on
// the UpdateCh.
func (reh EndpointsEventHandler) OnUpdate(old, cur interface{}) {
	oep := old.(*corev1.Endpoints)
	ocur := cur.(*corev1.Endpoints)
	if !reflect.DeepEqual(ocur.Subsets, oep.Subsets) {
		reh.UpdateCh.In() <- Event{
			Type: UpdateEvent,
			Obj:  cur,
		}
	}
}
