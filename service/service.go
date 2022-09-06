package service

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	co "github.com/cofunclabs/cofunc"
	"github.com/cofunclabs/cofunc/config"
	"github.com/cofunclabs/cofunc/pkg/nameid"
	"github.com/cofunclabs/cofunc/runtime"
	"github.com/cofunclabs/cofunc/runtime/actuator"
	"github.com/cofunclabs/cofunc/service/crontrigger"
	"github.com/cofunclabs/cofunc/service/exported"
	"github.com/cofunclabs/cofunc/service/logset"
	"github.com/cofunclabs/cofunc/service/resource"
	"github.com/cofunclabs/cofunc/std"
)

// SVC is the service layer, it provides API to access and manage the flows
type SVC struct {
	rt *runtime.Runtime
	// availables store all available flows, the key is the string of flow's id.
	availables map[string]exported.FlowMetaInsight
	// log service for flow and function
	log *logset.Logset
	// cron service for flow and function
	cron *crontrigger.CronTrigger
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
	// Create log service
	log := logset.New(logset.WithAddr(config.LogDir()))
	if err := log.Restore(); err != nil {
		panic(err)
	}
	// Create cron trigger service
	cron := crontrigger.New()
	cron.Start()
	// Create http trigger service

	return &SVC{
		rt:         runtime.New(),
		availables: all,
		log:        log,
		cron:       cron,
	}
}

// ListStdFunctions returns the list of the manifests of all standard functions.
func (s *SVC) ListStdFunctions(ctx context.Context) []exported.ListStdFunctions {
	var list []exported.ListStdFunctions
	all := std.ListAll()
	for _, m := range all {
		list = append(list, exported.ListStdFunctions{
			Name: m.Name,
			Desc: m.Description,
		})
	}
	return list
}

// InspectStdFunction returns the manifest of the standard function
func (s *SVC) InspectStdFunction(ctx context.Context, name string) exported.InspectStdFunction {
	m, _, _ := std.Lookup(name)
	if m == nil {
		return exported.InspectStdFunction{}
	}
	return exported.InspectStdFunction(*m)
}

// LookupID be used to lookup 'nameid.ID' by the string of flow's id or name.
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
	if err := s.rt.InitFlow(ctx, id); err != nil {
		return err
	}
	if err := s.rt.ExecFlow(ctx, id); err != nil {
		return err
	}
	return nil
}

// CancelRunningFlow cancels a running flow, the canceled flow not be started again automatically.
func (s *SVC) CancelRunningFlow(ctx context.Context, id nameid.ID) error {
	return s.rt.CancelFlow(ctx, id)
}

// AddFlow parse a flowl source file and add a flow instance into runtime
func (s *SVC) AddFlow(ctx context.Context, id nameid.ID, rd io.ReadCloser) error {
	defer rd.Close()
	if err := s.rt.ParseFlow(ctx, id, rd); err != nil {
		return err
	}
	return nil
}

// ReadyFlow initialize the flow and make it ready to run
func (s *SVC) ReadyFlow(ctx context.Context, id nameid.ID, toStdout bool) (exported.FlowRunningInsight, error) {
	createLogWriter := func(writerid string) (io.Writer, error) {
		return s.log.CreateBucket(id.ID()).CreateWriter(writerid)
	}
	beforeExec := func(id nameid.ID) error {
		// TODO:
		return nil
	}
	afterExec := func(id nameid.ID) error {
		// TODO:
		return nil
	}
	copy := func() resource.Resources {
		return resource.Resources{
			CronTrigger: s.cron,
		}
	}
	var opts = []runtime.FlowOption{runtime.WithBeforeFunc(beforeExec), runtime.WithAfterFunc(afterExec), runtime.WithCopyResources(copy)}
	if !toStdout {
		opts = append(opts, runtime.WithCreateLogwriter(createLogWriter))
	}
	if err := s.rt.InitFlow(ctx, id, opts...); err != nil {
		return exported.FlowRunningInsight{}, err
	}
	fi, err := s.InsightFlow(ctx, id)
	if err != nil {
		return exported.FlowRunningInsight{}, err
	}
	return fi, nil
}

// StartFlow starts a flow without event triggers, it will return a channel that can be used to
// wait for the flow to be finished.
func (s *SVC) StartFlow(ctx context.Context, id nameid.ID) chan error {
	wait := make(chan error, 1)
	go func() {
		defer close(wait)
		// NOTE: here used a new context to avoid the context be canceled by others
		ctx, cancel := context.WithCancel(context.Background())
		s.rt.FetchFlow(ctx, id, func(fb *runtime.FlowBody) error {
			fb.SetCancel(cancel)
			return nil
		})
		wait <- s.rt.ExecFlow(ctx, id)
	}()
	return wait
}

// StartFlowAndWait starts a flow without event triggers and wait for it to be finished.
func (s *SVC) StartFlowAndWait(ctx context.Context, id nameid.ID) error {
	return <-s.StartFlow(ctx, id)
}

// StartEventFlow starts a flow with event triggers, it will run in a goroutine,
// so the invoking will return immediately.
func (s *SVC) StartEventFlow(ctx context.Context, id nameid.ID) chan error {
	wait := make(chan error, 1)
	go func() {
		// NOTE: here used a new context to avoid the context be canceled by others
		ctx, cancel := context.WithCancel(context.Background())
		s.rt.FetchFlow(ctx, id, func(fb *runtime.FlowBody) error {
			fb.SetCancel(cancel)
			return nil
		})
		wait <- s.rt.StartEventTrigger(ctx, id)
	}()
	return wait
}

// StartEventFlowAndWait starts a flow with event triggers and wait for it to be finished.
func (s *SVC) StartEventFlowAndWait(ctx context.Context, id nameid.ID) error {
	return <-s.StartEventFlow(ctx, id)
}

// StartFlowOrEventFlow don't start the flows directly, it first start triggers of the flows,
// then execute the flows based on event from trigger. If not found any triggers of the flow, it will
// start the flow directly.
func (s *SVC) StartFlowOrEventFlow(ctx context.Context, id nameid.ID) error {
	has, err := s.rt.HasTrigger(id)
	if err != nil {
		return err
	}
	if !has {
		return s.StartFlowAndWait(ctx, id)
	}
	// NOTE: call StartEventTrigger will be blocked, that's goroutine will not finish until you cancel it
	return s.StartEventFlowAndWait(ctx, id)
}

// ViewLog be used to view the log of a flow or a function, the argument 'id' is the flow's id, the 'seq'
// is the sequence of the function, the 'w' argument is the output destination of the log.
func (s *SVC) ViewLog(ctx context.Context, id nameid.ID, seq int, w io.Writer) error {
	bucket, err := s.log.GetBucket(id.ID())
	if err != nil {
		return err
	}
	rd, err := bucket.CreateReader(strconv.Itoa(seq))
	if err != nil {
		return err
	}
	defer rd.Close()

	if _, err := io.Copy(w, rd); err != nil {
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
		id := nameid.New(co.FlowlPath2Name(path, dir))
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
