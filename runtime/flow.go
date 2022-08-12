//go:generate stringer -type=FlowStatus
package runtime

import (
	"sync"
	"time"

	"github.com/cofunclabs/cofunc/generator"
	"github.com/cofunclabs/cofunc/parser"
	"github.com/cofunclabs/cofunc/pkg/feedbackid"
	"github.com/cofunclabs/cofunc/service/exported"
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

var statusTable = map[FlowStatus]string{
	_flow_unknown: "UNKNOWN",
	_flow_added:   "ADDED",
	_flow_error:   "ERROR",
	_flow_ready:   "READY",
	_flow_running: "RUNNING",
	_flow_stopped: "STOPPED",
	_flow_updated: "UPDATED",
}

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

type functionResult struct {
	sync.Mutex
	functionResultBody
}

func (fr *functionResult) WithLock(exec func(body *functionResultBody)) {
	fr.Lock()
	defer fr.Unlock()
	exec(&fr.functionResultBody)
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
	status FlowStatus
	begin  time.Time
	end    time.Time
	// Save the result of function execution, the map key is node's seq
	results  map[int]*functionResult
	progress progress

	runq *generator.RunQueue
	ast  *parser.AST
}

func (b *FlowBody) Export() exported.FlowInsight {
	insight := exported.FlowInsight{
		Status:  statusTable[b.status],
		Begin:   b.begin,
		End:     b.end,
		Total:   len(b.progress.nodes),
		Running: len(b.progress.running),
		Done:    len(b.progress.done),
	}
	for _, seq := range b.progress.nodes {
		fr := b.results[seq]
		fr.WithLock(func(rb *functionResultBody) {
			insight.Nodes = append(insight.Nodes, struct {
				Seq       int    "json:\"seq\""
				Step      int    "json:\"step\""
				Name      string "json:\"name\""
				LastError error  "json:\"last_error\""
				Status    string "json:\"status\""
			}{
				Seq:       seq,
				Step:      rb.node.(generator.NodeExtend).Step(),
				Name:      rb.node.Name(),
				Status:    statusTable[rb.status],
				LastError: rb.err,
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

func newflow(id feedbackid.ID, runq *generator.RunQueue, ast *parser.AST) *Flow {
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

func (f *Flow) GetResult(seq int) *functionResult {
	f.Lock()
	defer f.Unlock()
	return f.results[seq]
}
