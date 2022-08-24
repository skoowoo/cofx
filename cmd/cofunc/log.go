package main

import (
	"context"
	"os"

	"github.com/cofunclabs/cofunc/pkg/nameid"
	"github.com/cofunclabs/cofunc/service"
)

func viewLog(id string, seq int) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := service.New()
	if err := svc.ViewLog(ctx, nameid.WrapID(id), seq, os.Stdout); err != nil {
		return err
	}

	return nil
}
