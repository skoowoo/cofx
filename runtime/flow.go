package runtime

import (
	"errors"
	"sync"
	"time"

	"github.com/cofunclabs/cofunc/parser"
	"github.com/cofunclabs/cofunc/pkg/feedbackid"
	"github.com/cofunclabs/cofunc/pkg/logout"
	"github.com/cofunclabs/cofunc/runtime/actuator"
	"github.com/cofunclabs/cofunc/service/exported"
)

type StatusType string

const (
	StatusAdded   = StatusType("ADDED")
	StatusReady   = StatusType("READY")
	StatusRunning = StatusType("RUNNING")
	StatusStopped = StatusType("STOPPED")
	StatusUpdated = StatusType("UPDATED")
)

type functionMetricsBody struct {
	fid feedbackid.ID
	// Last start time
	begin time.Time
	// Last end time
	end time.Time
	// Number of runs
	runs int
	// Whether the last time was executed
	executed bool
	// Whether there is an error in the function execution
	err    error
	status StatusType

	node actuator.Node
}

type functionMetrics struct {
	sync.Mutex
	functionMetricsBody
}

func (fm *functionMetrics) WithLock(exec func(body *functionMetricsBody)) {
	fm.Lock()
	defer fm.Unlock()
	exec(&fm.functionMetricsBody)
}

func (fm *functionMetrics) IsStatus(status StatusType) bool {
	fm.Lock()
	defer fm.Unlock()
	return fm.status == status
}

type progress struct {
	nodes   []int
	done    []int
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
	id     feedbackid.ID
	status StatusType
	begin  time.Time
	end    time.Time
	// Save the result metrics of function execution
	// the map is seq->functionMetrics
	metrics  map[int]*functionMetrics
	progress progress

	logger *logout.Output

	runq *actuator.RunQueue
	ast  *parser.AST
}

func (b *FlowBody) Export() exported.FlowInsight {
	insight := exported.FlowInsight{
		Name:    "",
		ID:      b.id.Value(),
		Status:  string(b.status),
		Begin:   b.begin,
		End:     b.end,
		Total:   len(b.progress.nodes),
		Running: len(b.progress.running),
		Done:    len(b.progress.done),
	}
	for _, seq := range b.progress.nodes {
		fr := b.metrics[seq]
		fr.WithLock(func(rb *functionMetricsBody) {
			insight.Nodes = append(insight.Nodes, exported.NodeInsight{
				Seq:       seq,
				Step:      rb.node.(actuator.Task).Step(),
				Name:      rb.node.Name(),
				Status:    string(rb.status),
				LastError: rb.err,
				Runs:      rb.runs,
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

func newflow(id feedbackid.ID, runq *actuator.RunQueue, ast *parser.AST) *Flow {
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

	f.progress.Reset()

	var (
		status StatusType = StatusReady
		begin  time.Time  = time.Now()
		end    time.Time
	)
	for seq, m := range f.metrics {
		m.WithLock(func(body *functionMetricsBody) {
			if m.begin.Unix() < begin.Unix() {
				begin = m.begin
			}
			if m.end.Unix() > end.Unix() {
				end = m.end
			}
			if m.status == StatusStopped {
				status = StatusStopped
				f.progress.PutDone(seq)
			}
			if m.status == StatusRunning {
				status = StatusRunning
				f.progress.PutRunning(seq)
			}
		})
	}

	f.status = status
	f.begin = begin
	f.end = end
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
	// The purpose of using a function to execute the code block is to avoid the deadlock,
	// because the 'f.Refresh()' method will also lock the 'f'
	err := func() error {
		f.Lock()
		defer f.Unlock()

		for _, m := range f.metrics {
			if !m.IsStatus(StatusStopped) {
				return errors.New("not stopped")
			}
		}
		for _, m := range f.metrics {
			m.WithLock(func(body *functionMetricsBody) {
				body.status = StatusReady
				body.begin = time.Time{}
				body.end = time.Time{}
				body.executed = false
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

func (f *Flow) GetMetrics(seq int) *functionMetrics {
	f.Lock()
	defer f.Unlock()
	return f.metrics[seq]
}
