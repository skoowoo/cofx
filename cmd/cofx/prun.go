package main

import (
	"context"
	"os"
	"sync"
	"time"

	co "github.com/cofxlabs/cofx"
	"github.com/cofxlabs/cofx/pkg/nameid"
	"github.com/cofxlabs/cofx/service"
	"github.com/cofxlabs/cofx/service/exported"
)

func prunEntry(nameorid nameid.NameOrID, fullscreen bool) error {
	svc := service.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var fid nameid.ID

	// If the argument 'nameorid' not contains the suffix ".flowl", We will treat it as a flow name or id, so we will lookup the flowl source path through
	// name or id.
	// if the argument 'nameorid' contains the suffix ".flowl", we will treat it as a full path of the flowl file, so can open it directly.
	fp := nameorid.String()
	if !co.IsFlowl(fp) {
		id, err := svc.LookupID(ctx, nameorid)
		if err != nil {
			return err
		}
		meta, err := svc.GetAvailableMeta(ctx, id)
		if err != nil {
			return err
		}
		fp = meta.Source
		fid = id
	} else {
		fid = nameid.New(co.FlowlPath2Name(fp))
	}
	f, err := os.Open(fp)
	if err != nil {
		return err
	}

	if err := svc.AddFlow(ctx, fid, f); err != nil {
		return err
	}
	if _, err := svc.ReadyFlow(ctx, fid, nil); err != nil {
		return err
	}

	var (
		lasterr error
		wg      sync.WaitGroup
	)
	wg.Add(2)
	// start the ui in a goroutine
	go func() {
		defer func() {
			wg.Done()
			cancel()
			// cancel() be used to stop the event trigger goroutine
			svc.CancelRunningFlow(ctx, fid)
		}()

		if err := startPrunView(fullscreen, func() (*exported.FlowRunningInsight, error) {
			fi, err := svc.InsightFlow(ctx, fid)
			return &fi, err
		}); err != nil {
			lasterr = err
		}
	}()

	time.Sleep(time.Second)
	// start the flow in a goroutine
	go func() {
		defer wg.Done()
		err := svc.StartFlowOrEventFlow(ctx, fid)
		if err != nil {
			lasterr = err
		}
		runCmdExited = true
	}()
	wg.Wait()

	if lasterr != nil {
		os.Exit(-1)
	}

	return nil
}
