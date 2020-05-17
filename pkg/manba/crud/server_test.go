package crud

import (
	"testing"

	"github.com/domgoer/manba-ingress/pkg/manba/state"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/stretchr/testify/assert"
)

func Test_serverPostAction_Create(t *testing.T) {
	action := serverPostAction{
		currentState: manbaState,
	}

	_, err := action.Create(&state.Server{
		Server: metapb.Server{
			ID: 1,
		},
		Metadata: state.Metadata{},
	})
	assert.Nil(t, err)

	_, err = action.Create(&state.Server{})
	assert.NotNil(t, err)
}

func Test_serverPostAction_Delete(t *testing.T) {
	action := serverPostAction{
		currentState: manbaState,
	}
	arg := &state.Server{
		Server: metapb.Server{
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

func Test_serverPostAction_Update(t *testing.T) {
	action := serverPostAction{
		currentState: manbaState,
	}
	arg := &state.Server{
		Server: metapb.Server{
			ID: 3,
		},
		Metadata: state.Metadata{},
	}
	_, err := action.Create(arg)
	assert.Nil(t, err)
	arg.Addr = "updated"
	_, err = action.Update(arg)
	assert.Nil(t, err)
}
