package crud

import (
	"testing"

	"github.com/domgoer/manba-ingress/pkg/manba/state"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/stretchr/testify/assert"
)

func Test_bindPostAction_Create(t *testing.T) {
	action := bindPostAction{
		currentState: manbaState,
	}

	_, err := action.Create(&state.Bind{
		Bind: metapb.Bind{
			ClusterID: 1,
			ServerID:  1,
		},
		Metadata: state.Metadata{},
	})
	assert.Nil(t, err)

	_, err = action.Create(&state.Bind{})
	assert.NotNil(t, err)
}

func Test_bindPostAction_Delete(t *testing.T) {
	action := bindPostAction{
		currentState: manbaState,
	}
	arg := &state.Bind{
		Bind: metapb.Bind{
			ClusterID: 2,
			ServerID:  2,
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
