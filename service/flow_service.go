package service

import (
	"context"

	"github.com/cofunclabs/cofunc/pkg/feedbackid"
	"github.com/cofunclabs/cofunc/runtime"
	"github.com/cofunclabs/cofunc/service/exported"
)

func GetFlowInsight(ctx context.Context, rt *runtime.Runtime, fid feedbackid.ID) (exported.FlowInsight, error) {
	var fi exported.FlowInsight
	read := func(body *runtime.FlowBody) error {
		fi = body.Export()
		return nil
	}
	err := rt.InspectFlow(ctx, fid, read)
	return fi, err
}
