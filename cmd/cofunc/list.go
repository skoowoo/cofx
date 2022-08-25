package main

import (
	"context"

	"github.com/cofunclabs/cofunc/service"
)

func listFlows() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := service.New()
	availables, err := svc.ListAvailableFlows(ctx)
	if err != nil {
		return err
	}

	selected, err := startListingView(availables)
	if err != nil {
		return err
	}

	// to run the selected flow
	if selected.Source != "" {
		return runflowl(selected.Source)
	}
	return nil
}
