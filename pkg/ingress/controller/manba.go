/*
Copyright 2015 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"crypto/sha256"
	"encoding/json"
	"reflect"

	"github.com/domgoer/manba-ingress/pkg/ingress/controller/parser"
	"github.com/domgoer/manba-ingress/pkg/manba/diff"
	"github.com/domgoer/manba-ingress/pkg/manba/dump"
	"github.com/domgoer/manba-ingress/pkg/manba/solver"
	"github.com/domgoer/manba-ingress/pkg/manba/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// OnUpdate is called periodically by syncQueue to keep the configuration in sync.
// returning nil implies the synchronization finished correctly.
// Returning an error means requeue the update.
func (m *ManbaController) OnUpdate(state *parser.ManbaState) error {
	target := m.toStable(state)

	jsonConfig, err := json.Marshal(target)
	if err != nil {
		return errors.Wrap(err, "marshaling Man declarative configuration to JSON")
	}
	shaSum := sha256.Sum256(jsonConfig)
	if reflect.DeepEqual(m.runningConfigHash, shaSum) {
		glog.Info("no configuration change, skipping sync to Manba")
		return nil
	}
	m.onUpdate(target)
	return nil
}

func (m *ManbaController) onUpdate(targetRaw *dump.ManbaRawState) error {
	client := m.cfg.Client

	targetState, err := state.Get(targetRaw)
	if err != nil {
		return errors.Wrap(err, "get target state")
	}

	raw, err := dump.Get(client)
	if err != nil {
		return errors.Wrap(err, "loading configuration from manba")
	}

	currentState, err := state.Get(raw)
	if err != nil {
		return errors.Wrap(err, "get current state")
	}

	syncer, err := diff.NewSyncer(currentState, targetState)
	if err != nil {
		return errors.Wrap(err, "new syncer")
	}
	syncer.SilenceWarnings = true
	_, err = solver.Solve(nil, syncer, client, m.cfg.Concurrency)
	return err
}

func (m *ManbaController) toStable(state *parser.ManbaState) *dump.ManbaRawState {
	var s dump.ManbaRawState
	// for _, api := range state.APIs {
	// 	s.APIs = append(s.APIs, api.API)
	// }
	// sort.SliceStable(s.APIs, func(i, j int) bool {
	// 	return s.APIs[i].Name < s.APIs[j].Name
	// })
	return &s
}
