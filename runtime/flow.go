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
	FlowUnknown FlowStatus = iota
	FlowStopped
	FlowRunning
	FlowReady
	FlowError
	FlowAdded
	FlowUpdated
)

var statusTable = map[FlowStatus]string{
	FlowUnknown: "UNKNOWN",
	FlowAdded:   "ADDED",
	FlowError:   "ERROR",
	FlowReady:   "READY",
	FlowRunning: "RUNNING",
	FlowStopped: "STOPPED",
	FlowUpdated: "UPDATED",
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
		Name:    "",
		ID:      b.id.Value(),
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
			insight.Nodes = append(insight.Nodes, exported.NodeInsight{
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
		status FlowStatus = FlowReady
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
			if r.status == FlowStopped {
				status = FlowStopped
				f.progress.PutDone(seq)
			}
			if r.status == FlowRunning {
				status = FlowRunning
				f.progress.PutRunning(seq)
			}
			if r.status == FlowError {
				status = FlowError
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
