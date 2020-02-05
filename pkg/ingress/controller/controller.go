package controller

import "github.com/domgoer/manba-ingress/pkg/ingress/election"

type ManbaController struct {
	elector election.Elector

	stopCh chan struct{}
}
