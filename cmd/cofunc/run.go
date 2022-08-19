package main

import (
	"context"
	"fmt"
	"os"

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

	return service.New().RunOneFlow(ctx, fid, f)
}
