package crud

import (
	"strconv"

	"github.com/domgoer/manba-ingress/pkg/manba/state"
	manba "github.com/fagongzi/gateway/pkg/client"
)

// apiPostAction crud api in mem-db
type apiPostAction struct {
	currentState *state.ManbaState
}

func (c *apiPostAction) Create(arg Arg) (Arg, error) {
	return nil, c.currentState.APIs.Add(*arg.(*state.API))
}

func (c *apiPostAction) Delete(arg Arg) (Arg, error) {
	return nil, c.currentState.APIs.Delete(strconv.Itoa(int(apiFromStruct(arg).ID)))
}

func (c *apiPostAction) Update(arg Arg) (Arg, error) {
	return nil, c.currentState.APIs.Update(*arg.(*state.API))
}

// apiRawAction crud api in manba
type apiRawAction struct {
	client manba.Client
}

func apiFromObj(obj interface{}) *state.API {
	api, ok := obj.(*state.API)
	if !ok {
		panic("unexpected type, expected *state.API")
	}
	return api
}

func (c *apiRawAction) Create(arg Arg) (Arg, error) {
	event := eventFromArg(arg)
	api := apiFromObj(event.Obj)
	cb := c.client.NewAPIBuilder()
	id, err := cb.Use(api).Commit()
	if err != nil {
		return nil, err
	}
	api.ID = id
	return &state.API{API: api.API}, nil

}

func (c *apiRawAction) Delete(arg Arg) (Arg, error) {
	event := eventFromArg(arg)
	api := apiFromObj(event.Obj)
	err := c.client.RemoveAPI(api.ID)
	return api, err
}

func (c *apiRawAction) Update(arg Arg) (Arg, error) {
	event := eventFromArg(arg)
	api := apiFromObj(event.Obj)
	cb := c.client.NewAPIBuilder()
	_, err := cb.Use(api).Commit()
	return api, err
}

func apiFromStruct(arg Event) *state.API {
	api, ok := arg.Obj.(*state.API)
	if !ok {
		panic("unexpected type, expected *state.api")
	}
	return api
}
