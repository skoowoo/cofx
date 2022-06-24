package main

import (
	"context"
	"errors"
	"os"

	co "github.com/cofunclabs/cofunc"
	"github.com/cofunclabs/cofunc/pkg/feedbackid"
)

func runFlowl(name string) error {
	if !co.IsFlowl(name) {
		return errors.New("file is not a flowl: " + name)
	}
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
	}()

	sd := co.NewScheduler()

	fid := feedbackid.NewDefaultID(name)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := sd.AddFlow(ctx, fid, f); err != nil {
		return err
	}
	if err := sd.ReadyFlow(ctx, fid); err != nil {
		return err
	}
	if err := sd.StartFlow(ctx, fid); err != nil {
		return err
	}
	return nil
}
