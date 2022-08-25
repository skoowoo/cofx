package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	co "github.com/cofunclabs/cofunc"
	"github.com/cofunclabs/cofunc/pkg/nameid"
	"github.com/cofunclabs/cofunc/service"
	"github.com/cofunclabs/cofunc/service/exported"
)

func runflowl(name string, fullscreen bool) error {
	if !co.IsFlowl(name) {
		return fmt.Errorf("not '.flowl': file '%s'", name)
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
	if _, err := svc.ReadyFlow(ctx, fid); err != nil {
		return err
	}

	var lasterr error
	var wg sync.WaitGroup
	wg.Add(2)
	// start the ui in a goroutine
	go func() {
		defer wg.Done()
		if err := startRunningView(fullscreen, func() (*exported.FlowInsight, error) {
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
