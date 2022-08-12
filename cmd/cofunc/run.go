package main

import (
	"context"
	"fmt"
	"os"

	co "github.com/cofunclabs/cofunc"
	"github.com/cofunclabs/cofunc/pkg/feedbackid"
	"github.com/cofunclabs/cofunc/runtime"
	"github.com/cofunclabs/cofunc/service"
)

func runflowl(name string) error {
	if !co.IsFlowl(name) {
		return fmt.Errorf("not '.flowl': file '%s'", name)
	}
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	rt := runtime.New()

	fid := feedbackid.NewDefaultID(name)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := rt.AddFlow(ctx, fid, f); err != nil {
		return err
	}
	if err := rt.ReadyFlow(ctx, fid); err != nil {
		return err
	}
	if err := rt.StartFlow(ctx, fid); err != nil {
		return err
	}
	fi, err := service.GetFlowInsight(ctx, rt, fid)
	if err != nil {
		return err
	}
	fi.JsonPrint(os.Stdout)
	return nil
}
