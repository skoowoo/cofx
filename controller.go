package cofunc

import (
	"context"
	"io"
	"time"

	"github.com/cofunclabs/cofunc/internal/flow"
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
		entity: make(map[string]*flow.Flow),
	}
	return fc
}

func (ctrl *FlowController) AddFlow(ctx context.Context, fid feedbackid.ID, rd io.Reader) error {
	rq, bs, err := flowl.Parse(rd)
	if err != nil {
		return err
	}
	fw := flow.New(fid, rq, bs)
	if err := ctrl.flowstore.Store(fid.Value(), fw); err != nil {
		return err
	}
	fw.UpdateField(func(b *flow.Body) error {
		b.Status = flow.FLOW_ADDED
		return nil
	})
	return nil
}

func (ctrl *FlowController) ReadyFlow(ctx context.Context, fid feedbackid.ID) error {
	fw, err := ctrl.flowstore.Get(fid.Value())
	if err != nil {
		return err
	}

	ready := func(body *flow.Body) error {
		if body.Status == flow.FLOW_READY || body.Status == flow.FLOW_RUNNING {
			return nil
		}
		body.FnTotal = body.Runq().NodeNum()
		body.Results = make(map[string]*flow.FunctionResult)

		body.Runq().Foreach(func(stage int, n *flowl.Node) error {
			if err := n.Driver.Load(ctx, n.Args); err != nil {
				return err
			}
			body.ReadyFnCount += 1

			body.Results[n.Name] = &flow.FunctionResult{
				FID:          body.ID,
				Node:         n,
				ReturnValues: make(map[string]string),
			}
			return nil
		})
		body.Status = flow.FLOW_READY
		return nil
	}

	return fw.UpdateField(ready)
}

func (ctrl *FlowController) StartFlow(ctx context.Context, fid feedbackid.ID) error {
	fw, err := ctrl.flowstore.Get(fid.Value())
	if err != nil {
		return err
	}

	if err := fw.UpdateField(flow.ToRunning, flow.UpdateBegineTime); err != nil {
		return err
	}

	fw.Runq().Stage(func(stage int, node *flowl.Node) {
		batchFuncs := 0
		ch := make(chan *flow.FunctionResult, 10)

		for p := node; p != nil; p = p.Parallel {
			batchFuncs += 1
			go func(n *flowl.Node, fr *flow.FunctionResult) {
				fr.BeginTime = time.Now()
				fr.ReturnValues, fr.Err = n.Driver.Run(ctx)
				fr.EndTime = time.Now()
				select {
				case ch <- fr:
				case <-ctx.Done():
				}
			}(p, fw.Results[p.Name])
		}
		// waiting the batch functions to finish running
		for i := 0; i < batchFuncs; i++ {
			select {
			case r := <-ch:
				if r.Err != nil {
					fw.UpdateField(flow.ToError)
				}
			case <-ctx.Done():
				// canced
				close(ch)
			}
		}
	})

	if err := fw.UpdateField(flow.ToStopped, flow.UpdateEndTime); err != nil {
		return err
	}
	return nil
}

func (ctrl *FlowController) InspectFlow(ctx context.Context, fid feedbackid.ID, read func(flow.Body) error) error {
	flow, err := ctrl.flowstore.Get(fid.Value())
	if err != nil {
		return err
	}
	if err := flow.ReadField(read); err != nil {
		return err
	}
	return nil
}

func (ctrl *FlowController) StopFlow(ctx context.Context, fid feedbackid.ID) error {
	return nil
}

func (ctrl *FlowController) DeleteFlow(ctx context.Context, fid feedbackid.ID) error {
	return nil
}
