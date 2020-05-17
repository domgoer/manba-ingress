package crud

import (
	"testing"

	"github.com/domgoer/manba-ingress/pkg/manba/state"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/stretchr/testify/assert"
)

func Test_clusterPostAction_Create(t *testing.T) {
	action := clusterPostAction{
		currentState: manbaState,
	}

	_, err := action.Create(&state.Cluster{
		Cluster: metapb.Cluster{
			ID: 1,
		},
		Metadata: state.Metadata{},
	})
	assert.Nil(t, err)

	_, err = action.Create(&state.Cluster{})
	assert.NotNil(t, err)
}

func Test_clusterPostAction_Delete(t *testing.T) {
	action := clusterPostAction{
		currentState: manbaState,
	}
	arg := &state.Cluster{
		Cluster: metapb.Cluster{
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

func Test_clusterPostAction_Update(t *testing.T) {
	action := clusterPostAction{
		currentState: manbaState,
	}
	arg := &state.Cluster{
		Cluster: metapb.Cluster{
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
