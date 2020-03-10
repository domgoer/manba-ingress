package crud

import (
	"github.com/domgoer/manba-ingress/pkg/manba/state"
	manba "github.com/fagongzi/gateway/pkg/client"
)

// routingPostAction crud routing in mem-db
type routingPostAction struct {
	currentState *state.ManbaState
}

func (c *routingPostAction) Create(arg Arg) (Arg, error) {
	return nil, c.currentState.Routings.Add(*arg.(*state.Routing))
}

func (c *routingPostAction) Delete(arg Arg) (Arg, error) {
	return nil, c.currentState.Routings.Delete(arg.(*state.Routing).Identifier())
}

func (c *routingPostAction) Update(arg Arg) (Arg, error) {
	return nil, c.currentState.Routings.Update(*arg.(*state.Routing))
}

// routingRawAction crud routing in manba
type routingRawAction struct {
	client manba.Client
}

func routingFromObj(obj interface{}) *state.Routing {
	routing, ok := obj.(*state.Routing)
	if !ok {
		panic("unexpected type, expected *state.Routing")
	}
	return routing
}

func (c *routingRawAction) Create(arg Arg) (Arg, error) {
	event := eventFromArg(arg)
	routing := routingFromObj(event.Obj)
	cb := c.client.NewRoutingBuilder()
	id, err := cb.Use(routing.Routing).Commit()
	if err != nil {
		return nil, err
	}
	routing.ID = id
	return &state.Routing{Routing: routing.Routing}, nil

}

func (c *routingRawAction) Delete(arg Arg) (Arg, error) {
	event := eventFromArg(arg)
	routing := routingFromObj(event.Obj)
	err := c.client.RemoveRouting(routing.ID)
	return routing, err
}

func (c *routingRawAction) Update(arg Arg) (Arg, error) {
	event := eventFromArg(arg)
	routing := routingFromObj(event.Obj)
	cb := c.client.NewRoutingBuilder()
	_, err := cb.Use(routing.Routing).Commit()
	return routing, err
}

func routingFromStruct(arg Event) *state.Routing {
	routing, ok := arg.Obj.(*state.Routing)
	if !ok {
		panic("unexpected type, expected *state.routing")
	}
	return routing
}
