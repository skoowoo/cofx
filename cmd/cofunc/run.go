package main

import (
	"context"
	"os"

	co "github.com/cofunclabs/cofunc"
	"github.com/cofunclabs/cofunc/pkg/nameid"
	"github.com/cofunclabs/cofunc/service"
)

func runflowl(nameorid nameid.NameOrID) error {
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
		fid = nameid.New(service.Path2Name(fp))
	}
	f, err := os.Open(fp)
	if err != nil {
		return err
	}

	if err := svc.CreateFlow(ctx, fid, f); err != nil {
		return err
	}
	if _, err := svc.ReadyFlow(ctx, fid, true); err != nil {
		return err
	}
	if err := svc.StartOrWaitingEvent(ctx, fid); err != nil {
		return err
	}
	return nil
}
