package main

import (
	"context"
	"os"

	"github.com/cofunclabs/cofunc/pkg/feedbackid"
	"github.com/cofunclabs/cofunc/service"
)

func viewLog(id string, seq int) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := service.New()
	if err := svc.ViewLog(ctx, feedbackid.WrapID(id), seq, os.Stdout); err != nil {
		return err
	}

	return nil
}
