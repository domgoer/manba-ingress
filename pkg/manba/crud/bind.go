package crud

import (
	"strconv"

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
	return nil, c.currentState.Binds.Delete(strconv.Itoa(int(bindFromStruct(arg).ID)))
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
	cb := c.client.NewBindBuilder()
	id, err := cb.Use(bind).Commit()
	if err != nil {
		return nil, err
	}
	bind.ID = id
	return &state.Bind{Bind: bind.Bind}, nil

}

func (c *bindRawAction) Delete(arg Arg) (Arg, error) {
	event := eventFromArg(arg)
	bind := bindFromObj(event.Obj)
	err := c.client.RemoveBind(bind.ID)
	return bind, err
}

func (c *bindRawAction) Update(arg Arg) (Arg, error) {
	event := eventFromArg(arg)
	bind := bindFromObj(event.Obj)
	cb := c.client.NewBindBuilder()
	_, err := cb.Use(bind).Commit()
	return bind, err
}

func bindFromStruct(arg Event) *state.Bind {
	bind, ok := arg.Obj.(*state.Bind)
	if !ok {
		panic("unexpected type, expected *state.bind")
	}
	return bind
}
