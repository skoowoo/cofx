package main

import (
	"context"
	"fmt"
	"os"

	co "github.com/cofunclabs/cofunc"
	"github.com/cofunclabs/cofunc/pkg/feedbackid"
)

func runFlowl(name string) error {
	if !co.IsFlowl(name) {
		return fmt.Errorf("not '.flowl': file '%s'", name)
	}
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
	}()

	sched := co.New()

	fid := feedbackid.NewDefaultID(name)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := sched.AddFlow(ctx, fid, f); err != nil {
		return err
	}
	if err := sched.ReadyFlow(ctx, fid); err != nil {
		return err
	}
	if err := sched.StartFlow(ctx, fid); err != nil {
		return err
	}
	return nil
}
