package runtime

import (
	"errors"
	"sync"
	"time"

	"github.com/cofunclabs/cofunc/parser"
	"github.com/cofunclabs/cofunc/pkg/logfile"
	"github.com/cofunclabs/cofunc/pkg/nameid"
	"github.com/cofunclabs/cofunc/runtime/actuator"
	"github.com/cofunclabs/cofunc/service/exported"
)

type StatusType string

const (
	StatusAdded    = StatusType("ADDED")
	StatusReady    = StatusType("READY")
	StatusRunning  = StatusType("RUNNING")
	StatusStopped  = StatusType("STOPPED")
	StatusKilled   = StatusType("KILLED")
	StatusCanceled = StatusType("CANCELED")
)

type functionStatisticsBody struct {
	// Flow id
	fid nameid.ID
	// The start time of the last running
	begin time.Time
	// The end time of the last running
	end time.Time
	// The duration of the last running
	duration int64
	// Number of runs
	runs int
	// Whether there is an error in the function execution
	err error

	status StatusType
	node   actuator.Node
}

type functionStatistics struct {
	sync.Mutex
	functionStatisticsBody
}

func (fs *functionStatistics) WithLock(exec func(body *functionStatisticsBody)) {
	fs.Lock()
	defer fs.Unlock()
	exec(&fs.functionStatisticsBody)
}

func (fs *functionStatistics) IsStatus(status StatusType) bool {
	fs.Lock()
	defer fs.Unlock()
	return fs.status == status
}

type progress struct {
	// Stored seq number of all nodes in a flow
	nodes []int
	// Stored seq number of all fnished nodes in a flow
	done []int
	// Stored seq number of all running nodes in a flow; The key is the seq number, which is easy to find.
	running map[int]struct{}
}

func (p *progress) PutRunning(seq int) {
	if p.running == nil {
		p.running = make(map[int]struct{})
	}
	p.running[seq] = struct{}{}
}

func (p *progress) ResetRunning() {
	p.running = make(map[int]struct{})
}

func (p *progress) PutDone(seq int) {
	p.done = append(p.done, seq)
	delete(p.running, seq)
}

func (p *progress) Reset() {
	p.done = p.done[0:0]
	p.running = make(map[int]struct{})
}

type FlowBody struct {
	// Flow id
	id nameid.ID
	// The start time of the last running
	begin time.Time
	// The duration of the last running
	duration int64
	// Save the result statistics of function execution
	// the map is seq->functionStatistics
	statistics map[int]*functionStatistics
	// Saved the execution progress of all nodes
	progress progress

	status StatusType
	logger *logfile.Logfile

	runq *actuator.RunQueue
	ast  *parser.AST
}

func (b *FlowBody) Logger() *logfile.Logfile {
	return b.logger
}

func (b *FlowBody) Export() exported.FlowRunningInsight {
	insight := exported.FlowRunningInsight{
		Name:     b.id.Name(),
		ID:       b.id.ID(),
		Status:   string(b.status),
		Begin:    b.begin,
		Duration: b.duration,
		Total:    len(b.progress.nodes),
		Running:  len(b.progress.running),
		Done:     len(b.progress.done),
	}
	for _, seq := range b.progress.nodes {
		fm := b.statistics[seq]
		fm.WithLock(func(mb *functionStatisticsBody) {
			insight.Nodes = append(insight.Nodes, exported.NodeRunningInsight{
				Seq:       seq,
				Step:      mb.node.(actuator.Task).Step(),
				Function:  mb.node.(actuator.Task).Driver().FunctionName(),
				Driver:    mb.node.(actuator.Task).Driver().Name(),
				Name:      mb.node.Name(),
				Status:    string(mb.status),
				LastError: mb.err,
				Runs:      mb.runs,
				Duration:  mb.duration,
			})
		})
	}
	return insight
}

// Flow
//
type Flow struct {
	sync.RWMutex
	FlowBody
}

func newflow(id nameid.ID, runq *actuator.RunQueue, ast *parser.AST) *Flow {
	return &Flow{
		FlowBody: FlowBody{
			id:   id,
			runq: runq,
			ast:  ast,
		},
	}
}

func (f *Flow) WithLock(exec func(body *FlowBody) error) error {
	f.Lock()
	defer f.Unlock()
	return exec(&f.FlowBody)
}

func (f *Flow) Refresh() error {
	f.Lock()
	defer f.Unlock()

	if f.status == StatusRunning {
		f.duration = time.Since(f.begin).Milliseconds()
	}
	var isready bool = true
	f.progress.Reset()
	for seq, s := range f.statistics {
		s.WithLock(func(body *functionStatisticsBody) {
			if s.status != StatusReady {
				isready = false
			}
			switch s.status {
			case StatusStopped:
				f.progress.PutDone(seq)
			case StatusRunning:
				f.progress.PutRunning(seq)
			}
		})
	}
	if isready {
		f.status = StatusReady
	}
	return nil
}

func (f *Flow) IsReady() bool {
	f.Lock()
	defer f.Unlock()
	return f.status == StatusReady
}

func (f *Flow) IsStopped() bool {
	f.Lock()
	defer f.Unlock()
	return f.status == StatusStopped
}

func (f *Flow) IsRunning() bool {
	f.Lock()
	defer f.Unlock()
	return f.status == StatusRunning
}

func (f *Flow) IsAdded() bool {
	f.Lock()
	defer f.Unlock()
	return f.status == StatusAdded
}

func (f *Flow) ToReady() error {
	if f.IsReady() {
		return nil
	}
	// The purpose of using a function to execute the code block is to avoid the deadlock,
	// because the 'f.Refresh()' method will also lock the 'f'
	err := func() error {
		f.Lock()
		defer f.Unlock()

		for _, s := range f.statistics {
			if !s.IsStatus(StatusStopped) {
				return errors.New("not stopped")
			}
		}
		for _, s := range f.statistics {
			s.WithLock(func(body *functionStatisticsBody) {
				body.status = StatusReady
				body.begin = time.Time{}
				body.end = time.Time{}
				body.err = nil
				body.runs = 0
			})
		}

		return f.logger.Reset()
	}() // To avoid deadlock
	if err != nil {
		return err
	}

	return f.Refresh()
}

func (f *Flow) GetRunQ() *actuator.RunQueue {
	f.Lock()
	defer f.Unlock()
	return f.runq
}

func (f *Flow) GetAST() *parser.AST {
	f.Lock()
	defer f.Unlock()
	return f.ast
}

func (f *Flow) GetStatistics(seq int) *functionStatistics {
	f.Lock()
	defer f.Unlock()
	return f.statistics[seq]
}
