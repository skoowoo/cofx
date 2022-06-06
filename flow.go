package funcflow

import (
	"context"
	"sync"
	"time"

	"github.com/autoflowlabs/funcflow/internal/flowl"
	"github.com/autoflowlabs/funcflow/pkg/feedbackid"
)

type FlowStatus int

const (
	_FLOW_UNKNOWN FlowStatus = iota
	_FLOW_STOPPED
	_FLOW_RUNNING
	_FLOW_READY
	_FLOW_ERROR
	_FLOW_ADDED
	_FLOW_UPDATED
	_FLOW_RUNNING_AND_UPDATED
	_FLOW_DELETED
	_FLOW_RUNNING_AND_DELETED
)

type FunctionResult struct {
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
			fn:           v,
			returnValues: make(map[string]string),
		}
	}
	f.status = _FLOW_READY
	return nil
}

// ExecuteAndWaitFunc exec the flow, and will execute runner of the functions
func (f *Flow) ExecuteAndWaitFunc(ctx context.Context) error {
	// begin
	f.SetWithLock(func(s *Flow) {
		s.status = _FLOW_RUNNING
		s.beginTime = time.Now()
	})

	// functions running
	ch := make(chan *FunctionResult, 10)
	batchFuncs := 0
	queue := f.runq.Queue
	for e := queue.Front(); e != nil; e = e.Next() {
		fn := e.Value.(*flowl.Function)
		for p := fn; p != nil; p = p.Parallel {
			batchFuncs += 1
			go func(ctx context.Context, p *flowl.Function, r *FunctionResult) {
				r.err = p.Runner.Run()
				ch <- r
			}(ctx, p, f.result[p.Name])
		}
		// waiting
		for batchFuncs > 0 {
			select {
			case r := <-ch:
				if r.err != nil {
					f.SetWithLock(func(s *Flow) {
						s.status = _FLOW_ERROR
					})
				}
				batchFuncs -= 1
			case <-ctx.Done():
				// canced
			}
		}
	}

	// end
	f.SetWithLock(func(s *Flow) {
		if s.status == _FLOW_RUNNING {
			s.status = _FLOW_STOPPED
		}
		f.endTime = time.Now()
	})
	return nil
}

// Cancel stop the flow, the running functions continue to run until ends
func (f *Flow) Cancel() {
	f.cancel()
}

// Kill force stop the flow, the running functions will be stopped immediately
func (f *Flow) Kill() {
	f.Cancel()
}
