package crud

import (
	"github.com/domgoer/manba-ingress/pkg/manba/state"
	manba "github.com/fagongzi/gateway/pkg/client"
)

// serverPostAction crud server in mem-db
type serverPostAction struct {
	currentState *state.ManbaState
}

func (c *serverPostAction) Create(arg Arg) (Arg, error) {
	return nil, c.currentState.Servers.Add(*arg.(*state.Server))
}

func (c *serverPostAction) Delete(arg Arg) (Arg, error) {
	return nil, c.currentState.Servers.Delete(arg.(*state.Server).Identifier())
}

func (c *serverPostAction) Update(arg Arg) (Arg, error) {
	return nil, c.currentState.Servers.Update(*arg.(*state.Server))
}

// serverRawAction crud server in manba
type serverRawAction struct {
	client manba.Client
}

func serverFromObj(obj interface{}) *state.Server {
	server, ok := obj.(*state.Server)
	if !ok {
		panic("unexpected type, expected *state.Server")
	}
	return server
}

func (c *serverRawAction) Create(arg Arg) (Arg, error) {
	event := eventFromArg(arg)
	server := serverFromObj(event.Obj)
	cb := c.client.NewServerBuilder()
	id, err := cb.Use(server.Server).Commit()
	if err != nil {
		return nil, err
	}
	server.ID = id
	return &state.Server{Server: server.Server}, nil

}

func (c *serverRawAction) Delete(arg Arg) (Arg, error) {
	event := eventFromArg(arg)
	server := serverFromObj(event.Obj)
	err := c.client.RemoveServer(server.ID)
	return server, err
}

func (c *serverRawAction) Update(arg Arg) (Arg, error) {
	event := eventFromArg(arg)
	server := serverFromObj(event.Obj)
	cb := c.client.NewServerBuilder()
	_, err := cb.Use(server.Server).Commit()
	return server, err
}
