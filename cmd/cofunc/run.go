package main

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	co "github.com/cofunclabs/cofunc"
	"github.com/cofunclabs/cofunc/config"
	"github.com/cofunclabs/cofunc/pkg/nameid"
	"github.com/cofunclabs/cofunc/service"
	"github.com/cofunclabs/cofunc/service/exported"
)

func runflowl(name string, toStdout bool, fullscreen bool) error {
	// If the argument 'name' not contains the suffix ".flowl", We will treat it as a flow name
	// So we will generate the full path of the flowl file based on the flow name.
	// if the arument 'name' contains the suffix ".flowl", we will treat it as a full path of the flowl file, so can open it directly.
	if !co.IsFlowl(name) {
		name = filepath.Join(config.FlowSourceDir(), name) + ".flowl"
	}
	f, err := os.Open(name)
	if err != nil {
		return err
	}

	fid := nameid.New(name)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := service.New()
	if err := svc.CreateFlow(ctx, fid, f); err != nil {
		return err
	}
	if _, err := svc.ReadyFlow(ctx, fid, toStdout); err != nil {
		return err
	}
	// toStdout will print the output of flow into stdout
	if toStdout {
		_, err := svc.StartFlow(ctx, fid)
		if err != nil {
			return err
		}
		return nil
	}

	var (
		lasterr error
		wg      sync.WaitGroup
	)
	wg.Add(2)
	// start the ui in a goroutine
	go func() {
		defer wg.Done()
		if err := startRunningView(fullscreen, func() (*exported.FlowRunningInsight, error) {
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
		fi, err := svc.StartFlow(ctx, fid)
		if err != nil {
			lasterr = err
		}
		_ = fi
		runCmdExited = true
	}()
	wg.Wait()

	if lasterr != nil {
		os.Exit(-1)
	}

	return nil
}
