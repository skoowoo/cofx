package main

import (
	"context"
	"os"

	"github.com/cofxlabs/cofx/pkg/nameid"
	"github.com/cofxlabs/cofx/service"
)

func viewLog(nameorid nameid.NameOrID, seq int) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := service.New()
	id, err := svc.LookupID(ctx, nameorid)
	if err != nil {
		return err
	}
	if err := svc.ViewLog(ctx, id, seq, os.Stdout); err != nil {
		return err
	}

	return nil
}
