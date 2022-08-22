package service

import (
	"context"
	"io"

	"github.com/cofunclabs/cofunc/config"
	"github.com/cofunclabs/cofunc/pkg/feedbackid"
	"github.com/cofunclabs/cofunc/pkg/logfile"
	"github.com/cofunclabs/cofunc/runtime"
	"github.com/cofunclabs/cofunc/service/exported"
)

type SVC struct {
	rt *runtime.Runtime
}

func New() *SVC {
	if err := config.Init(); err != nil {
		panic(err)
	}
	return &SVC{
		rt: runtime.New(),
	}
}

func (s *SVC) InsightFlow(ctx context.Context, fid feedbackid.ID) (exported.FlowInsight, error) {
	var fi exported.FlowInsight
	read := func(body *runtime.FlowBody) error {
		fi = body.Export()
		return nil
	}
	err := s.rt.FetchFlow(ctx, fid, read)
	return fi, err
}

func (s *SVC) RunFlow(ctx context.Context, id feedbackid.ID, rd io.ReadCloser) error {
	if err := s.rt.ParseFlow(ctx, id, rd); err != nil {
		rd.Close()
		return err
	} else {
		rd.Close()
	}
	if err := s.rt.InitFlow(ctx, id); err != nil {
		return err
	}
	if err := s.rt.ExecFlow(ctx, id); err != nil {
		return err
	}
	return nil
}

func (s *SVC) CreateFlow(ctx context.Context, id feedbackid.ID, rd io.ReadCloser) error {
	if err := s.rt.ParseFlow(ctx, id, rd); err != nil {
		rd.Close()
		return err
	} else {
		rd.Close()
	}
	return nil
}

func (s *SVC) ReadyFlow(ctx context.Context, id feedbackid.ID) (exported.FlowInsight, error) {
	if err := s.rt.InitFlow(ctx, id); err != nil {
		return exported.FlowInsight{}, err
	}
	fi, err := s.InsightFlow(ctx, id)
	if err != nil {
		return exported.FlowInsight{}, err
	}
	return fi, nil
}

func (s *SVC) StartFlow(ctx context.Context, id feedbackid.ID) (exported.FlowInsight, error) {
	if err := s.rt.ExecFlow(ctx, id); err != nil {
		return exported.FlowInsight{}, err
	}
	fi, err := s.InsightFlow(ctx, id)
	if err != nil {
		return exported.FlowInsight{}, err
	}
	return fi, nil
}

// ViewLog be used to view the log of a flow or a function, the argument 'id' is the flow's id, the 'seq'
// is the sequence of the function, the 'w' argument is the output destination of the log.
func (s *SVC) ViewLog(ctx context.Context, id feedbackid.ID, seq int, w io.Writer) error {
	dir, err := config.LogFunctionDir(id.Value(), seq)
	if err != nil {
		return err
	}
	logger, err := logfile.File(config.LogFunctionFile(dir))
	if err != nil {
		return err
	}
	defer logger.Close()
	if _, err := io.Copy(w, logger); err != nil {
		return err
	}
	return nil
}
