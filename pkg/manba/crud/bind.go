package crud

import (
	"github.com/domgoer/manba-ingress/pkg/manba/state"
	manba "github.com/fagongzi/gateway/pkg/client"
)

// bindPostAction crud bind in mem-db
type bindPostAction struct {
	currentState *state.ManbaState
}

func (c *bindPostAction) Create(arg Arg) (Arg, error) {
	return nil, c.currentState.Binds.Add(*arg.(*state.Bind))
}

func (c *bindPostAction) Delete(arg Arg) (Arg, error) {
	return nil, c.currentState.Binds.Delete(arg.(*state.Bind).Identifier())
}

func (c *bindPostAction) Update(arg Arg) (Arg, error) {
	return nil, c.currentState.Binds.Update(*arg.(*state.Bind))
}

// bindRawAction crud bind in manba
type bindRawAction struct {
	client manba.Client
}

func bindFromObj(obj interface{}) *state.Bind {
	bind, ok := obj.(*state.Bind)
	if !ok {
		panic("unexpected type, expected *state.Bind")
	}
	return bind
}

func (c *bindRawAction) Create(arg Arg) (Arg, error) {
	event := eventFromArg(arg)
	bind := bindFromObj(event.Obj)
	err := c.client.AddBind(bind.GetClusterID(), bind.GetServerID())
	if err != nil {
		return nil, err
	}
	return &state.Bind{Bind: bind.Bind}, nil

}

func (c *bindRawAction) Delete(arg Arg) (Arg, error) {
	event := eventFromArg(arg)
	bind := bindFromObj(event.Obj)
	err := c.client.RemoveBind(bind.GetClusterID(), bind.GetServerID())
	return bind, err
}

// Update bind wont be updated
func (c *bindRawAction) Update(arg Arg) (Arg, error) {
	return nil, nil
}
