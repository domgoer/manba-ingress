package solver

import (
	"errors"
	"strings"

	"github.com/domgoer/manba-ingress/pkg/manba/crud"
	"github.com/domgoer/manba-ingress/pkg/manba/diff"
	manba "github.com/fagongzi/gateway/pkg/client"
	"github.com/golang/glog"
)

// Stats holds the stats related to a Solve.
type Stats struct {
	CreateOps int
	UpdateOps int
	DeleteOps int
}

// Solve generates a diff and walks the graph.
func Solve(doneCh chan struct{}, syncer *diff.Syncer,
	client manba.Client, parallelism int) (Stats, error) {
	r := crud.NewRawRegistry(client)

	var stats Stats
	recordOp := func(op crud.Op) {
		switch op {
		case crud.Create:
			stats.CreateOps = stats.CreateOps + 1
		case crud.Update:
			stats.UpdateOps = stats.UpdateOps + 1
		case crud.Delete:
			stats.DeleteOps = stats.DeleteOps + 1
		}
	}

	errs := syncer.Run(doneCh, parallelism, func(a crud.Arg) (crud.Arg, error) {
		var err error
		var result crud.Arg
		e, ok := a.(crud.Event)
		if !ok {
			return nil, errors.New("unknown operation")
		}
		switch e.Op {
		case crud.Create:
			glog.Infof("creating <%s>, data: %+v", e.Kind, e.Obj)
		case crud.Update:
			diffString, err := Diff(e.OldObj, e.Obj)
			if err != nil {
				return nil, err
			}
			glog.Infof("old: %+v, new: %+v",e.OldObj, e.Obj)
			glog.Infof("updating <%s>, diff: <%s>, data: %+v", e.Kind, diffString, e.Obj)
		case crud.Delete:
			glog.Infof("deleting <%s>, data: %+v", e.Kind, e.Obj)
		default:
			panic("unknown operation " + e.Op.String())
		}

		// sync mode
		// fire the request to Manba
		result, err = r.Do(e.Kind, e.Op, e)
		if err != nil {
			return nil, err
		}
		// record operation in both: diff and sync commands
		recordOp(e.Op)

		return result, nil
	})
	var list []string
	for _, e := range errs {
		list = append(list, e.Error())
	}
	return stats, errors.New(strings.Join(list, "\n"))

}
