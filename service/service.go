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

// Path2Name be used to convert the path of a flowl source file to the flow's name.
func Path2Name(path string, trimpath ...string) string {
	if len(trimpath) > 0 {
		path = strings.TrimPrefix(path, trimpath[0])
	} else {
		path = strings.TrimPrefix(path, config.FlowSourceDir())
	}
	return co.TruncFlowl(path)
}

// SVC is the service layer, it provides API to access and manage the flows
type SVC struct {
	rt *runtime.Runtime
	// availables store all available flows, the key is the string of flow's id.
	availables map[string]exported.FlowMetaInsight
}

// New create a service layer instance
func New() *SVC {
	if err := config.Init(); err != nil {
		panic(err)
	}
	all, err := restoreAvailables(config.FlowSourceDir())
	if err != nil {
		panic(err)
	}
	return &SVC{
		rt:         runtime.New(),
		availables: all,
	}
}

// LookupID be used to lookup 'nameid.ID' by the string of flow's id or name
func (s *SVC) LookupID(ctx context.Context, nameorid nameid.NameOrID) (nameid.ID, error) {
	return nameid.Guess(nameorid, func(id string) *nameid.NameID {
		if v, ok := s.availables[id]; ok {
			return nameid.Wrap(v.Name, v.ID)
		}
		return nil
	})
}

// ListAvailables returns the list of all available flows in the flow source directory that be defined by
// the environment variable 'CO_FLOW_SOURCE_DIR'.
func (s *SVC) ListAvailables(ctx context.Context) []exported.FlowMetaInsight {
	var availables []exported.FlowMetaInsight
	for _, f := range s.availables {
		availables = append(availables, f)
	}
	return availables
}

// GetAvailableMeta returns the meta of the flow with the flow id
func (s *SVC) GetAvailableMeta(ctx context.Context, id nameid.ID) (exported.FlowMetaInsight, error) {
	if v, ok := s.availables[id.ID()]; ok {
		return v, nil
	}
	return exported.FlowMetaInsight{}, fmt.Errorf("not found meta: flow '%s'", id)
}

// InsightFlow exports the statistics of the flow
func (s *SVC) InsightFlow(ctx context.Context, fid nameid.ID) (exported.FlowRunningInsight, error) {
	var fi exported.FlowRunningInsight
	export := func(body *runtime.FlowBody) error {
		fi = body.Export()
		return nil
	}
	err := s.rt.FetchFlow(ctx, fid, export)
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

// CreateFlow parse a flowl source file, then create a flow instance in runtime
func (s *SVC) CreateFlow(ctx context.Context, id nameid.ID, rd io.ReadCloser) error {
	if err := s.rt.ParseFlow(ctx, id, rd); err != nil {
		rd.Close()
		return err
	} else {
		rd.Close()
	}
	return nil
}

// ReadyFlow initialize the flow and make it ready to run
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

// StartFlow starts a flow without event triggers and returns the statistics of the flow
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

// StartOrWaitingEvent don't start the flows directly, it first start triggers of the flows,
// then execute the flows based on event from trigger. If not found any triggers of the flow, it will
// start the flow directly.
func (s *SVC) StartOrWaitingEvent(ctx context.Context, id nameid.ID) error {
	has, err := s.rt.HasTrigger(id)
	if err != nil {
		return err
	}
	if !has {
		_, err := s.StartFlow(ctx, id)
		return err
	}
	// NOTE: call StartEventTrigger will be blocked, that's goroutine will not finish until you cancel it
	if err := s.rt.StartEventTrigger(ctx, id); err != nil {
		// TODO: log
		_ = err
	}
	return nil
}

// ViewLog be used to view the log of a flow or a function, the argument 'id' is the flow's id, the 'seq'
// is the sequence of the function, the 'w' argument is the output destination of the log.
func (s *SVC) ViewLog(ctx context.Context, id nameid.ID, seq int, w io.Writer) error {
	dir, err := config.LogFunctionDir(id.ID(), seq)
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

// restoreAvailables returns all available flows in flow source directory, the method will
// parse the flowl source file and generate AST.
func restoreAvailables(dir string) (map[string]exported.FlowMetaInsight, error) {
	var (
		sources []string
		flows   map[string]exported.FlowMetaInsight = make(map[string]exported.FlowMetaInsight)
	)
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
		id := nameid.New(Path2Name(path, dir))
		meta := exported.FlowMetaInsight{
			Name: id.Name(),
			ID:   id.ID(),
		}
		if err := parseOneFlowl(path, &meta); err != nil {
			meta.Desc = fmt.Errorf("%w: parse '%s'", err, path).Error()
			// -1 means that have a error in the flowl source file
			meta.Total = -1
		}
		flows[meta.ID] = meta
	}
	return flows, nil
}

func parseOneFlowl(name string, meta *exported.FlowMetaInsight) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	q, ast, err := actuator.New(f)
	if err != nil {
		return err
	}
	var total int
	q.WalkNode(func(n actuator.Node) error {
		total++
		return nil
	})
	meta.Source = name
	meta.Total = total
	meta.Desc = ast.Desc()
	return nil
}
