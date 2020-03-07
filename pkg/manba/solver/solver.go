package solver

import (
	manba "github.com/fagongzi/gateway/pkg/client"
)

// Stats holds the stats related to a Solve.
type Stats struct {
	CreateOps int
	UpdateOps int
	DeleteOps int
}

// Solve generates a diff and walks the graph.
func Solve(doneCh chan struct{}, syncer interface{},
	client *manba.Client, parallelism int, dry bool) (Stats, []error) {

}
