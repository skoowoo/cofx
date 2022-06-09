//go:generate stringer -type=FlowStatus
package funcflow

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
	fn           *flowl.Function
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

func (f *Flow) SetWithLock(set func(*Flow)) {
	f.Lock()
	defer f.Unlock()
	set(f)
}

// Ready make the flow ready, will execute loader of the functions
func (f *Flow) Ready(ctx context.Context) error {
	f.Lock()
	defer f.Unlock()

	functions := f.runq.Functions
	f.fnTotal = len(functions)
	f.result = make(map[string]*FunctionResult)

	for _, v := range functions {
		if err := v.Loader.Load(); err != nil {
			return err
		}
		f.readyFnCount += 1

		f.result[v.Name] = &FunctionResult{
			fid:          f.ID,
			fn:           v,
			returnValues: make(map[string]string),
		}
	}
	f.status = FLOW_READY
	return nil
}

// ExecuteAndWaitFunc exec the flow, and will execute runner of the functions
func (f *Flow) ExecuteAndWaitFunc(ctx context.Context) error {
	// begin
	f.SetWithLock(func(s *Flow) {
		s.status = FLOW_RUNNING
		s.beginTime = time.Now()
	})

	// functions running
	f.runq.Step(func(first *flowl.Function) {
		batchFuncs := 0
		ch := make(chan *FunctionResult, 10)

		for p := first; p != nil; p = p.Parallel {
			batchFuncs += 1
			go func(fn *flowl.Function, r *FunctionResult) {
				r.err = fn.Runner.Run()
				select {
				case ch <- r:
				case <-ctx.Done():
				}
			}(p, f.result[p.Name])
		}
		// waiting
		for i := 0; i < batchFuncs; i++ {
			select {
			case r := <-ch:
				if r.err != nil {
					f.SetWithLock(func(s *Flow) {
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
	f.SetWithLock(func(s *Flow) {
		if s.status == FLOW_RUNNING {
			s.status = FLOW_STOPPED
		}
		f.endTime = time.Now()
	})
	return nil
}

// Cancel stop the flow, the running functions continue to run until ends
func (f *Flow) Cancel() {
	f.cancel()
}
