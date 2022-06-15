package cofunc

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/cofunclabs/cofunc/internal/flowl"
	"github.com/cofunclabs/cofunc/pkg/feedbackid"
)

// FlowController
//
type FlowController struct {
	flowstore *FlowStore
}

func NewFlowController() *FlowController {
	fc := &FlowController{}
	fc.flowstore = &FlowStore{
		entity: make(map[string]*Flow),
	}
	return fc
}

func (fc *FlowController) AddFlow(ctx context.Context, fid feedbackid.ID, rd io.Reader) error {
	rq, bs, err := flowl.Parse(rd)
	if err != nil {
		return err
	}
	flow := &Flow{
		ID:         fid,
		runq:       rq,
		blockstore: bs,
		status:     FLOW_ADDED,
	}
	if err := fc.flowstore.Store(flow.ID.Value(), flow); err != nil {
		return err
	}
	return nil
}

func (fc *FlowController) ReadyFlow(ctx context.Context, fid feedbackid.ID) error {
	flow, err := fc.flowstore.Get(fid.Value())
	if err != nil {
		return err
	}

	ready := func(ctx context.Context, f *Flow) error {
		if f.status == FLOW_READY || f.status == FLOW_RUNNING {
			return nil
		}
		nodes := f.runq.FNodes
		f.fnTotal = len(nodes)
		f.result = make(map[string]*FunctionResult)

		for _, n := range nodes {
			if err := n.Driver.Load(ctx, n.Args()); err != nil {
				return err
			}
			f.readyFnCount += 1

			f.result[n.Name] = &FunctionResult{
				fid:          f.ID,
				fnode:        n,
				returnValues: make(map[string]string),
			}
		}
		f.status = FLOW_READY
		return nil
	}

	return fc.flowstore.ModifyEntity(ctx, flow, ready)
}

func (fc *FlowController) StartFlow(ctx context.Context, fid feedbackid.ID) error {
	flow, err := fc.flowstore.Get(fid.Value())
	if err != nil {
		return err
	}

	if err := fc.flowstore.ModifyEntity(ctx, flow, markFlowAsRunning); err != nil {
		return err
	}

	flow.runq.Step(func(batch *flowl.FunctionNode) {
		batchFuncs := 0
		ch := make(chan *FunctionResult, 10)

		for p := batch; p != nil; p = p.Parallel {
			batchFuncs += 1
			go func(fn *flowl.FunctionNode, r *FunctionResult) {
				r.beginTime = time.Now()
				r.returnValues, r.err = fn.Driver.Run(ctx)
				r.endTime = time.Now()
				select {
				case ch <- r:
				case <-ctx.Done():
				}
			}(p, flow.result[p.Name])
		}
		// waiting the batch functions to finish running
		for i := 0; i < batchFuncs; i++ {
			select {
			case r := <-ch:
				if r.err != nil {
					fc.flowstore.ModifyEntity(ctx, flow, markFlowAsError)
				}
			case <-ctx.Done():
				// canced
				close(ch)
			}
		}
	})

	if err := fc.flowstore.ModifyEntity(ctx, flow, markFlowAsStopped); err != nil {
		return err
	}
	return nil
}

func markFlowAsRunning(ctx context.Context, f *Flow) error {
	if f.status == FLOW_RUNNING {
		return errors.New("function is running: " + f.ID.Value())
	}
	if f.status == FLOW_ADDED {
		return errors.New("function is not ready: " + f.ID.Value())
	}
	f.status = FLOW_RUNNING
	f.beginTime = time.Now()
	return nil
}

func markFlowAsStopped(ctx context.Context, f *Flow) error {
	if f.status == FLOW_ERROR {
		// todo
	}
	if f.status == FLOW_RUNNING {
		f.status = FLOW_STOPPED
	}
	f.endTime = time.Now()
	return nil
}

func markFlowAsError(ctx context.Context, f *Flow) error {
	f.status = FLOW_ERROR
	return nil
}

func (fc *FlowController) InspectFlow(ctx context.Context, fid feedbackid.ID) (*Flow, error) {
	return fc.flowstore.Get(fid.Value())
}

func (fc *FlowController) StopFlow(ctx context.Context, fid feedbackid.ID) error {
	return nil
}

func (fc *FlowController) DeleteFlow(ctx context.Context, fid feedbackid.ID) error {
	return nil
}
