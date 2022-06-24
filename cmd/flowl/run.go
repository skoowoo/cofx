package main

import (
	"context"
	"os"

	"github.com/cofunclabs/cofunc"
	"github.com/cofunclabs/cofunc/internal/flowl"
	. "github.com/cofunclabs/cofunc/pkg/assertutils"
	"github.com/cofunclabs/cofunc/pkg/feedbackid"
)

func runFlowl(name string) error {
	if err := flowl.IsFlowl(name); err != nil {
		return err
	}
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
	}()

	ctrl := cofunc.NewController()
	PanicIfNil(ctrl)

	fid := feedbackid.NewDefaultID(name)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ctrl.AddFlow(ctx, fid, f); err != nil {
		return err
	}
	if err := ctrl.ReadyFlow(ctx, fid); err != nil {
		return err
	}
	if err := ctrl.StartFlow(ctx, fid); err != nil {
		return err
	}
	return nil
}
