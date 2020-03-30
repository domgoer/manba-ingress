package crud

import (
	"github.com/domgoer/manba-ingress/pkg/manba/state"
	manba "github.com/fagongzi/gateway/pkg/client"
	"github.com/pkg/errors"
)

// Registry can hold Kinds and their respective CRUD operations.
type Registry struct {
	types map[Kind]Actions
}

// NewPostProcess init mem db crud
func NewPostProcess(current *state.ManbaState) Registry {
	var r Registry
	r.MustRegister("cluster", &clusterPostAction{current})
	r.MustRegister("routing", &routingPostAction{current})
	r.MustRegister("server", &serverPostAction{current})
	r.MustRegister("api", &apiPostAction{current})
	r.MustRegister("bind", &bindPostAction{current})
	return r
}

// NewRawRegistry init manba crud
func NewRawRegistry(client manba.Client) Registry {
	var r Registry
	r.MustRegister("cluster", &clusterRawAction{client})
	r.MustRegister("routing", &routingRawAction{client})
	r.MustRegister("server", &serverRawAction{client})
	r.MustRegister("api", &apiRawAction{client})
	r.MustRegister("bind", &bindRawAction{client})
	return r
}

func (r *Registry) typesMap() map[Kind]Actions {
	if r.types == nil {
		r.types = make(map[Kind]Actions)
	}
	return r.types
}

// Register a kind with actions.
// An error will be returned if kind was previously registered.
func (r *Registry) Register(kind Kind, a Actions) error {
	if kind == "" {
		return errors.New("kind cannot be empty")
	}
	m := r.typesMap()
	if _, ok := m[kind]; ok {
		return errors.New("kind '" + string(kind) + "' already registered")
	}
	m[kind] = a
	return nil
}

// MustRegister is same as Register but panics on error.
func (r *Registry) MustRegister(kind Kind, a Actions) {
	err := r.Register(kind, a)
	if err != nil {
		panic(err)
	}
}

// Get returns actions associated with kind.
// An error will be returned if kind was never registered.
func (r *Registry) Get(kind Kind) (Actions, error) {
	if kind == "" {
		return nil, errors.New("kind cannot be empty")
	}
	m := r.typesMap()
	a, ok := m[kind]
	if !ok {
		return nil, errors.New("kind '" + string(kind) + "' is not registered")
	}
	return a, nil
}

// Create calls the registered create action of kind with arg
// and returns the result and error (if any).
func (r *Registry) Create(kind Kind, arg Arg) (Arg, error) {
	a, err := r.Get(kind)
	if err != nil {
		return nil, errors.Wrap(err, "create failed")
	}

	res, err := a.Create(arg)
	if err != nil {
		return nil, errors.Wrap(err, "create failed")
	}
	return res, nil
}

// Update calls the registered update action of kind with arg
// and returns the result and error (if any).
func (r *Registry) Update(kind Kind, arg Arg) (Arg, error) {
	a, err := r.Get(kind)
	if err != nil {
		return nil, errors.Wrap(err, "update failed")
	}

	res, err := a.Update(arg)
	if err != nil {
		return nil, errors.Wrap(err, "update failed")
	}
	return res, nil
}

// Delete calls the registered delete action of kind with arg
// and returns the result and error (if any).
func (r *Registry) Delete(kind Kind, arg ...Arg) (Arg, error) {
	a, err := r.Get(kind)
	if err != nil {
		return nil, errors.Wrap(err, "delete failed")
	}

	res, err := a.Delete(arg)
	if err != nil {
		return nil, errors.Wrap(err, "delete failed")
	}
	return res, nil
}

// Do calls an aciton based on op with arg and returns the result and error.
func (r *Registry) Do(kind Kind, op Op, arg Arg) (Arg, error) {
	a, err := r.Get(kind)
	if err != nil {
		return nil, errors.Wrapf(err, "%v failed", op)
	}

	var res Arg

	switch op.name {
	case Create.name:
		res, err = a.Create(arg)
	case Update.name:
		res, err = a.Update(arg)
	case Delete.name:
		res, err = a.Delete(arg)
	default:
		return nil, errors.New("unknown operation: " + op.name)
	}

	if err != nil {
		return nil, errors.Wrapf(err, "%v failed %v", op, arg)
	}
	return res, nil
}
