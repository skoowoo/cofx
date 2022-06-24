package cofunc

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/cofunclabs/cofunc/pkg/feedbackid"
)

// Controller
//
type Controller struct {
	store *FlowStore
}

func NewController() *Controller {
	c := &Controller{}
	c.store = &FlowStore{
		entity: make(map[string]*Flow),
	}
	return c
}

func (ctl *Controller) AddFlow(ctx context.Context, fid feedbackid.ID, rd io.Reader) error {
	rq, ast, err := ParseFlowl(rd)
	if err != nil {
		return err
	}
	fw := NewFlow(fid, rq, ast)
	if err := ctl.store.Store(fid.Value(), fw); err != nil {
		return err
	}
	fw.UpdateField(func(b *FlowBody) error {
		b.status = FLOW_ADDED
		return nil
	})
	return nil
}

func (ctl *Controller) ReadyFlow(ctx context.Context, fid feedbackid.ID) error {
	fw, err := ctl.store.Get(fid.Value())
	if err != nil {
		return err
	}

	ready := func(body *FlowBody) error {
		if body.status == FLOW_READY || body.status == FLOW_RUNNING {
			return nil
		}
		body.total = body.Runq().NodeNum()
		body.results = make(map[string]*FunctionResult)

		body.Runq().Foreach(func(stage int, n *Node) error {
			if err := n.Driver.Load(ctx, n.args); err != nil {
				return err
			}
			body.ready += 1

			body.results[n.Name] = &FunctionResult{
				fid:     body.id,
				node:    n,
				returns: make(map[string]string),
			}
			return nil
		})
		body.status = FLOW_READY
		return nil
	}

	return fw.UpdateField(ready)
}

func (ctl *Controller) StartFlow(ctx context.Context, fid feedbackid.ID) error {
	fw, err := ctl.store.Get(fid.Value())
	if err != nil {
		return err
	}

	if err := fw.UpdateField(toRunning, updateBeginTime); err != nil {
		return err
	}

	fw.Runq().Forstage(func(stage int, node *Node) error {
		// find out functions at the stage
		errResults := make([]*FunctionResult, 0)
		results := make([]*FunctionResult, 0)
		for p := node; p != nil; p = p.Parallel {
			results = append(results, fw.results[p.Name])
		}
		ch := make(chan *FunctionResult, len(results))

		// parallel run functions at the stage
		for p := node; p != nil; p = p.Parallel {
			go func(n *Node, fr *FunctionResult) {
				fr.begin = time.Now()
				fr.returns, fr.err = n.Driver.Run(ctx)
				fr.end = time.Now()
				select {
				case ch <- fr:
				case <-ctx.Done():
				}
			}(p, fw.results[p.Name])
		}

		// waiting functions at the stage to finish running
		for i := 0; i < len(results); i++ {
			select {
			case r := <-ch:
				if r.err != nil {
					fw.UpdateField(toError)
					errResults = append(errResults, r)
				}
			case <-ctx.Done():
				// canced
				close(ch)
			}
		}

		if l := len(errResults); l != 0 {
			return errors.New("occurred error at stage")
		}
		return nil
	})

	if err := fw.UpdateField(toStopped, updateEndTime); err != nil {
		return err
	}
	return nil
}

func (ctl *Controller) InspectFlow(ctx context.Context, fid feedbackid.ID, read func(FlowBody) error) error {
	flow, err := ctl.store.Get(fid.Value())
	if err != nil {
		return err
	}
	if err := flow.ReadField(read); err != nil {
		return err
	}
	return nil
}

func (ctl *Controller) StopFlow(ctx context.Context, fid feedbackid.ID) error {
	return nil
}

func (ctl *Controller) DeleteFlow(ctx context.Context, fid feedbackid.ID) error {
	return nil
}
