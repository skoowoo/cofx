//go:generate stringer -type=FlowStatus
package runtime

import (
	"sync"
	"time"

	"github.com/cofunclabs/cofunc/generator"
	"github.com/cofunclabs/cofunc/parser"
	"github.com/cofunclabs/cofunc/pkg/feedbackid"
)

type FlowStatus int

const (
	_flow_unknown FlowStatus = iota
	_flow_stopped
	_flow_running
	_flow_ready
	_flow_error
	_flow_added
	_flow_updated
)

type functionResultBody struct {
	fid  feedbackid.ID
	node generator.Node
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
	status FlowStatus
}

type FunctionResult struct {
	sync.Mutex
	functionResultBody
}

func (fr *FunctionResult) WithLock(exec func(body *functionResultBody)) {
	fr.Lock()
	defer fr.Unlock()
	exec(&fr.functionResultBody)
}

type progress struct {
	total   int
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

type flowBody struct {
	id     feedbackid.ID
	status FlowStatus
	begin  time.Time
	end    time.Time
	// Save the result of function execution, the map key is node's seq
	results  map[int]*FunctionResult
	progress progress

	runq *generator.RunQueue
	ast  *parser.AST
}

// Flow
//
type Flow struct {
	sync.RWMutex
	flowBody
}

func newflow(id feedbackid.ID, runq *generator.RunQueue, ast *parser.AST) *Flow {
	return &Flow{
		flowBody: flowBody{
			id:   id,
			runq: runq,
			ast:  ast,
		},
	}
}

func (f *Flow) WithLock(exec func(body *flowBody) error) error {
	f.Lock()
	defer f.Unlock()
	return exec(&f.flowBody)
}

func (f *Flow) Refresh() error {
	f.Lock()
	defer f.Unlock()

	var (
		status FlowStatus = _flow_ready
		begin  time.Time  = time.Now()
		end    time.Time
	)
	for seq, r := range f.results {
		r.WithLock(func(body *functionResultBody) {
			if r.begin.Unix() < begin.Unix() {
				begin = r.begin
			}
			if r.end.Unix() > end.Unix() {
				end = r.end
			}
			if r.status == _flow_stopped {
				status = _flow_stopped
				f.progress.PutDone(seq)
			}
			if r.status == _flow_running {
				status = _flow_running
				f.progress.PutRunning(seq)
			}
			if r.status == _flow_error {
				status = _flow_error
				f.progress.PutDone(seq)
			}
		})
	}

	f.status = status
	f.begin = begin
	f.end = end
	return nil
}

func (f *Flow) GetRunQ() *generator.RunQueue {
	f.Lock()
	defer f.Unlock()
	return f.runq
}

func (f *Flow) GetAST() *parser.AST {
	f.Lock()
	defer f.Unlock()
	return f.ast
}

func (f *Flow) GetResult(seq int) *FunctionResult {
	f.Lock()
	defer f.Unlock()
	return f.results[seq]
}

func (f *Flow) readField(read ...func(flowBody) error) error {
	f.RLock()
	defer f.RUnlock()
	for _, rd := range read {
		if err := rd(f.flowBody); err != nil {
			return err
		}
	}
	return nil
}
