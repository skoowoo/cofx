package funcflow

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/autoflowlabs/funcflow/internal/flowl"
	"github.com/autoflowlabs/funcflow/pkg/feedbackid"
	"github.com/sirupsen/logrus"
)

type FlowStatus int

const (
	_FLOW_STOPPED FlowStatus = iota
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
	fnTotal      int32
	readyFnCount int32
	successCount int32
	result       map[string]*FunctionResult
}

func (f *Flow) SetStatus(set func(current FlowStatus) FlowStatus) {
	f.Lock()
	defer f.Unlock()
	f.status = set(f.status)
}
func (f *Flow) AddSuccessCount(n int32) {
	atomic.AddInt32(&f.successCount, n)
}

// Ready make the flow ready, will execute loader of the functions
func (f *Flow) Ready(ctx context.Context) error {
	functions := f.runq.Functions
	f.fnTotal = int32(len(functions))
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
	f.SetStatus(func(current FlowStatus) FlowStatus {
		return _FLOW_READY
	})
	return nil
}

// ExecuteAndWaiting exec the flow, and will execute runner of the functions
func (f *Flow) ExecuteAndWaiting(ctx context.Context) error {
	// begin
	f.SetStatus(func(current FlowStatus) FlowStatus {
		return _FLOW_RUNNING
	})
	f.beginTime = time.Now()

	// running
	queue := f.runq.Queue
	for e := queue.Front(); e != nil; e = e.Next() {
		fn := e.Value.(*flowl.Function)
		var wg sync.WaitGroup
		for p := fn; p != nil; p = p.Parallel {
			wg.Add(1)
			r := f.result[p.Name]

			go func(p *flowl.Function, r *FunctionResult) {
				defer func() {
					r.endTime = time.Now()
					wg.Done()
				}()
				r.beginTime = time.Now()
				if err := p.Runner.Run(); err != nil {
					f.SetStatus(func(current FlowStatus) FlowStatus {
						return _FLOW_ERROR
					})
					r.err = err
					logrus.Errorln(err)
					return
				}
				f.AddSuccessCount(1)

			}(p, r)
		}
		wg.Wait()
	}

	// end
	f.endTime = time.Now()
	f.SetStatus(func(current FlowStatus) FlowStatus {
		if current == _FLOW_RUNNING {
			return _FLOW_STOPPED
		}
		return current
	})
	return nil
}

// Cancel stop the flow, the running functions continue to run until ends
func (f *Flow) Cancel(ctx context.Context) {

}

// Kill force stop the flow, the running functions will be stopped immediately
func (f *Flow) Kill(ctx context.Context) {

}
