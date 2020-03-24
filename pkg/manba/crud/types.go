package crud

import "github.com/hbagdi/deck/crud"

type t string

// Kind represents Kind of an entity or object.
type Kind string

// Op represents
type Op struct {
	name string
}

func (op *Op) String() string {
	return op.name
}

var (
	// Create is a constant representing create operations.
	Create = Op{"Create"}
	// Update is a constant representing update operations.
	Update = Op{"Update"}
	// Delete is a constant representing delete operations.
	Delete = Op{"Delete"}
)

// Event represents an event to perform
// an imperative operation
// that gets Manba closer to the target state.
type Event struct {
	Op     Op
	Kind   Kind
	Obj    interface{}
	OldObj interface{}
}

func eventFromArg(arg crud.Arg) Event {
	event, ok := arg.(Event)
	if !ok {
		panic("unexpected type, expected diff.Event")
	}
	return event
}

// Arg is an argument to a callback function.
type Arg interface{}

// Actions is an interface for CRUD operations on any entity
type Actions interface {
	Create(Arg) (Arg, error)
	Delete(Arg) (Arg, error)
	Update(Arg) (Arg, error)
}