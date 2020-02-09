package controller

import (
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
