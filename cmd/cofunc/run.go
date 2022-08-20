package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	co "github.com/cofunclabs/cofunc"
	"github.com/cofunclabs/cofunc/pkg/feedbackid"
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

	fid := feedbackid.NewDefaultID(name)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := service.New()
	if err := svc.CreateFlow(ctx, fid, f); err != nil {
		return err
	}
	fi, err := svc.ReadyFlow(ctx, fid)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)
	// start the ui in a goroutine
	go func() {
		defer wg.Done()
		if err := startRunningUI(svc, &fi); err != nil {
			log.Fatalln(err)
			os.Exit(-1)
		}
	}()

	time.Sleep(time.Second)
	// start the flow in a goroutine
	go func() {
		defer wg.Done()
		fi, err := svc.StartFlow(ctx, fid)
		if err != nil {
			log.Fatalln(err)
			os.Exit(-1)
		}
		_ = fi
	}()
	wg.Wait()

	return nil
}
