package service

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	co "github.com/cofunclabs/cofunc"
	"github.com/cofunclabs/cofunc/config"
	"github.com/cofunclabs/cofunc/pkg/logfile"
	"github.com/cofunclabs/cofunc/pkg/nameid"
	"github.com/cofunclabs/cofunc/runtime"
	"github.com/cofunclabs/cofunc/runtime/actuator"
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

func (s *SVC) InsightFlow(ctx context.Context, fid nameid.ID) (exported.FlowRunningInsight, error) {
	var fi exported.FlowRunningInsight
	read := func(body *runtime.FlowBody) error {
		fi = body.Export()
		return nil
	}
	err := s.rt.FetchFlow(ctx, fid, read)
	return fi, err
}

func (s *SVC) RunFlow(ctx context.Context, id nameid.ID, rd io.ReadCloser) error {
	if err := s.rt.ParseFlow(ctx, id, rd); err != nil {
		rd.Close()
		return err
	} else {
		rd.Close()
	}
	if err := s.rt.InitFlow(ctx, id, runtime.GetStdoutLogger); err != nil {
		return err
	}
	if err := s.rt.ExecFlow(ctx, id); err != nil {
		return err
	}
	return nil
}

func (s *SVC) CreateFlow(ctx context.Context, id nameid.ID, rd io.ReadCloser) error {
	if err := s.rt.ParseFlow(ctx, id, rd); err != nil {
		rd.Close()
		return err
	} else {
		rd.Close()
	}
	return nil
}

func (s *SVC) ReadyFlow(ctx context.Context, id nameid.ID, toStdout bool) (exported.FlowRunningInsight, error) {
	var get runtime.GetLogger
	if toStdout {
		get = runtime.GetStdoutLogger
	} else {
		get = runtime.GetDefaultLogger(id)
	}
	if err := s.rt.InitFlow(ctx, id, get); err != nil {
		return exported.FlowRunningInsight{}, err
	}
	fi, err := s.InsightFlow(ctx, id)
	if err != nil {
		return exported.FlowRunningInsight{}, err
	}
	return fi, nil
}

func (s *SVC) StartFlow(ctx context.Context, id nameid.ID) (exported.FlowRunningInsight, error) {
	if err := s.rt.ExecFlow(ctx, id); err != nil {
		return exported.FlowRunningInsight{}, err
	}
	fi, err := s.InsightFlow(ctx, id)
	if err != nil {
		return exported.FlowRunningInsight{}, err
	}
	return fi, nil
}

// ViewLog be used to view the log of a flow or a function, the argument 'id' is the flow's id, the 'seq'
// is the sequence of the function, the 'w' argument is the output destination of the log.
func (s *SVC) ViewLog(ctx context.Context, id nameid.ID, seq int, w io.Writer) error {
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

// ListAvailableFlows returns the list of all available flows in flow source directory, the method will
// parse the flowl source file and generate AST.
func (s *SVC) ListAvailableFlows(ctx context.Context) ([]exported.FlowMetaInsight, error) {
	var (
		sources []string
		flows   []exported.FlowMetaInsight
	)
	dir := config.FlowSourceDir()
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("%w: access path '%s'", err, path)
		}
		if info.IsDir() {
			return nil
		}
		sources = append(sources, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: walk dir '%s' to list flows", err, dir)
	}
	for _, path := range sources {
		id := nameid.New(co.TruncFlowl(strings.TrimPrefix(path, dir)))
		meta := exported.FlowMetaInsight{
			Name: id.Name(),
			ID:   id.Value(),
		}
		if err := parseOneFlowl(path, &meta); err != nil {
			return nil, fmt.Errorf("%w: parse '%s'", err, path)
		}
		flows = append(flows, meta)
	}
	return flows, nil
}

func parseOneFlowl(name string, meta *exported.FlowMetaInsight) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	q, _, err := actuator.New(f)
	if err != nil {
		return err
	}
	var total int
	q.ForfuncNode(func(n actuator.Node) error {
		total++
		return nil
	})
	meta.Source = name
	meta.Total = total
	return nil
}
