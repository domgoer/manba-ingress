package state

import (
	"reflect"
	"testing"

	memdb "github.com/hashicorp/go-memdb"
)

func TestClusterCollection_GetAll(t *testing.T) {
	type fields struct {
		db *memdb.MemDB
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*Cluster
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ClusterCollection{
				db: tt.fields.db,
			}
			got, err := c.GetAll()
			if (err != nil) != tt.wantErr {
				t.Errorf("ClusterCollection.GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClusterCollection.GetAll() = %v, want %v", got, tt.want)
			}
		})
	}
}
