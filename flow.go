//go:generate stringer -type=FlowStatus
package cofunc

import (
	"context"
	"sync"
	"time"

	"github.com/cofunclabs/cofunc/internal/flowl"
	"github.com/cofunclabs/cofunc/pkg/feedbackid"
)

type FlowStatus int

const (
	FLOW_UNKNOWN FlowStatus = iota
	FLOW_STOPPED
	FLOW_RUNNING
	FLOW_READY
	FLOW_ERROR
	FLOW_ADDED
	FLOW_UPDATED
	FLOW_RUNNING_AND_UPDATED
	FLOW_DELETED
	FLOW_RUNNING_AND_DELETED
)

type FunctionResult struct {
	fid          feedbackid.ID
	fnode        *flowl.FunctionNode
	returnValues map[string]string
	beginTime    time.Time
	endTime      time.Time
	err          error
}

// Flow
//
type Flow struct {
	sync.RWMutex
	ID           feedbackid.ID
	runq         *flowl.RunQueue
	blockstore   *flowl.BlockStore
	status       FlowStatus
	beginTime    time.Time
	endTime      time.Time
	fnTotal      int
	readyFnCount int
	successCount int
	result       map[string]*FunctionResult
	cancel       context.CancelFunc
}

func (fw *Flow) SetWithLock(set func(*Flow)) {
	fw.Lock()
	defer fw.Unlock()
	set(fw)
}

// Ready make the flow ready, will execute loader of the functions
func (fw *Flow) Ready(ctx context.Context) error {
	fw.Lock()
	defer fw.Unlock()

	nodes := fw.runq.FNodes
	fw.fnTotal = len(nodes)
	fw.result = make(map[string]*FunctionResult)

	for _, n := range nodes {
		if err := n.Driver.Load(ctx, n.Args()); err != nil {
			return err
		}
		fw.readyFnCount += 1

		fw.result[n.Name] = &FunctionResult{
			fid:          fw.ID,
			fnode:        n,
			returnValues: make(map[string]string),
		}
	}
	fw.status = FLOW_READY
	return nil
}

// ExecuteAndWaitFunc exec the flow, and will execute runner of the functions
func (fw *Flow) ExecuteAndWaitFunc(ctx context.Context) error {
	// begin
	fw.SetWithLock(func(s *Flow) {
		s.status = FLOW_RUNNING
		s.beginTime = time.Now()
	})

	// functions running
	fw.runq.Step(func(first *flowl.FunctionNode) {
		batchFuncs := 0
		ch := make(chan *FunctionResult, 10)

		for p := first; p != nil; p = p.Parallel {
			batchFuncs += 1
			go func(fn *flowl.FunctionNode, r *FunctionResult) {
				r.returnValues, r.err = fn.Driver.Run(ctx)
				select {
				case ch <- r:
				case <-ctx.Done():
				}
			}(p, fw.result[p.Name])
		}
		// waiting
		for i := 0; i < batchFuncs; i++ {
			select {
			case r := <-ch:
				if r.err != nil {
					fw.SetWithLock(func(s *Flow) {
						s.status = FLOW_ERROR
					})
				}
			case <-ctx.Done():
				// canced
				close(ch)
			}
		}
	})

	// end
	fw.SetWithLock(func(s *Flow) {
		if s.status == FLOW_RUNNING {
			s.status = FLOW_STOPPED
		}
		fw.endTime = time.Now()
	})
	return nil
}

// Cancel stop the flow, the running functions continue to run until ends
func (fw *Flow) Cancel() {
	fw.cancel()
}
