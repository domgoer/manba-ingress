package crud

import (
	"testing"

	"github.com/domgoer/manba-ingress/pkg/manba/state"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/stretchr/testify/assert"
)

var manbaState *state.ManbaState

func init() {
	manbaState, _ = state.NewManbaState()
}

func Test_apiPostAction_Create(t *testing.T) {
	action := apiPostAction{
		currentState: manbaState,
	}

	_, err := action.Create(&state.API{
		API: metapb.API{
			ID: 1,
		},
		Metadata: state.Metadata{},
	})
	assert.Nil(t, err)

	_, err = action.Create(&state.API{})
	assert.NotNil(t, err)
}

func Test_apiPostAction_Delete(t *testing.T) {
	action := apiPostAction{
		currentState: manbaState,
	}
	arg := &state.API{
		API: metapb.API{
			ID: 2,
		},
		Metadata: state.Metadata{},
	}
	_, err := action.Create(arg)
	assert.Nil(t, err)
	_, err = action.Delete(arg)
	assert.Nil(t, err)
	_, err = action.Delete(arg)
	assert.NotNil(t, err)
}

func Test_apiPostAction_Update(t *testing.T) {
	action := apiPostAction{
		currentState: manbaState,
	}
	arg := &state.API{
		API: metapb.API{
			ID: 3,
		},
		Metadata: state.Metadata{},
	}
	_, err := action.Create(arg)
	assert.Nil(t, err)
	arg.Name = "updated"
	_, err = action.Update(arg)
	assert.Nil(t, err)
}
