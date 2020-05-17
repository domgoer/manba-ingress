package crud

import (
	"testing"

	"github.com/domgoer/manba-ingress/pkg/manba/state"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/stretchr/testify/assert"
)

func Test_routingPostAction_Create(t *testing.T) {
	action := routingPostAction{
		currentState: manbaState,
	}

	_, err := action.Create(&state.Routing{
		Routing: metapb.Routing{
			ID: 1,
		},
		Metadata: state.Metadata{},
	})
	assert.Nil(t, err)

	_, err = action.Create(&state.Routing{})
	assert.NotNil(t, err)
}

func Test_routingPostAction_Delete(t *testing.T) {
	action := routingPostAction{
		currentState: manbaState,
	}
	arg := &state.Routing{
		Routing: metapb.Routing{
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

func Test_routingPostAction_Update(t *testing.T) {
	action := routingPostAction{
		currentState: manbaState,
	}
	arg := &state.Routing{
		Routing: metapb.Routing{
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
