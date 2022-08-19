package service

import (
	"context"
	"io"

	"github.com/cofunclabs/cofunc/pkg/feedbackid"
	"github.com/cofunclabs/cofunc/runtime"
	"github.com/cofunclabs/cofunc/service/exported"
)

type SVC struct {
	rt *runtime.Runtime
}

func New() *SVC {
	return &SVC{
		rt: runtime.New(),
	}
}

func (s *SVC) GetFlowInsight(ctx context.Context, fid feedbackid.ID) (exported.FlowInsight, error) {
	var fi exported.FlowInsight
	read := func(body *runtime.FlowBody) error {
		fi = body.Export()
		return nil
	}
	err := s.rt.InspectFlow(ctx, fid, read)
	return fi, err
}

func (s *SVC) RunOneFlow(ctx context.Context, id feedbackid.ID, rd io.ReadCloser) error {
	if err := s.rt.AddFlow(ctx, id, rd); err != nil {
		rd.Close()
		return err
	} else {
		rd.Close()
	}
	if err := s.rt.ReadyFlow(ctx, id); err != nil {
		return err
	}
	if err := s.rt.StartFlow(ctx, id); err != nil {
		return err
	}
	return nil
}
