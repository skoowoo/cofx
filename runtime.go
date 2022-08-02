package cofunc

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/cofunclabs/cofunc/pkg/feedbackid"
)

// Sched
//
type Sched struct {
	store *flowstore
}

func New() *Sched {
	s := &Sched{}
	s.store = &flowstore{
		entity: make(map[string]*Flow),
	}
	return s
}

func (sd *Sched) AddFlow(ctx context.Context, fid feedbackid.ID, rd io.Reader) error {
	rq, ast, err := ParseFlowl(rd)
	if err != nil {
		return err
	}
	fw := newflow(fid, rq, ast)
	if err := sd.store.store(fid.Value(), fw); err != nil {
		return err
	}
	fw.updateField(func(b *flowBody) error {
		b.status = _flow_added
		return nil
	})
	return nil
}

func (sd *Sched) ReadyFlow(ctx context.Context, fid feedbackid.ID) error {
	fw, err := sd.store.get(fid.Value())
	if err != nil {
		return err
	}

	ready := func(body *flowBody) error {
		if body.status == _flow_ready || body.status == _flow_running {
			return nil
		}
		body.total = body.GetRunQ().FuncNodeNum()
		body.results = make(map[string]*FunctionResult)

		err := body.GetRunQ().ForfuncNode(func(stage int, n Node) error {
			if err := n.Init(ctx); err != nil {
				return err
			}
			body.ready += 1

			body.results[n.Name()] = &FunctionResult{
				fid:     body.id,
				node:    n,
				returns: make(map[string]string),
			}
			return nil
		})
		if err != nil {
			return err
		}
		body.status = _flow_ready
		return nil
	}

	return fw.updateField(ready)
}

func (sd *Sched) StartFlow(ctx context.Context, fid feedbackid.ID) error {
	fw, err := sd.store.get(fid.Value())
	if err != nil {
		return err
	}

	if err := fw.updateField(toRunning, updateBeginTime); err != nil {
		return err
	}

	fw.GetRunQ().ForstageAndExec(ctx, func(stage int, batch []Node) error {
		ch := make(chan *FunctionResult, len(batch))
		// parallel run functions at the stage
		for _, node := range batch {
			go func(n Node, fr *FunctionResult) {
				fr.begin = time.Now()
				fr.err = n.Exec(ctx)
				fr.end = time.Now()

				select {
				case ch <- fr:
				case <-ctx.Done():
				}
			}(node, fw.results[node.Name()])
		}

		// waiting functions at the stage to finish running
		errResults := make([]*FunctionResult, 0)
		for i := 0; i < len(batch); i++ {
			select {
			case r := <-ch:
				if r.err != nil {
					fw.updateField(toError)
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

	if err := fw.updateField(toStopped, updateEndTime); err != nil {
		return err
	}
	return nil
}

func (sd *Sched) InspectFlow(ctx context.Context, fid feedbackid.ID, read func(flowBody) error) error {
	flow, err := sd.store.get(fid.Value())
	if err != nil {
		return err
	}
	if err := flow.readField(read); err != nil {
		return err
	}
	return nil
}

func (sd *Sched) StopFlow(ctx context.Context, fid feedbackid.ID) error {
	return nil
}

func (sd *Sched) DeleteFlow(ctx context.Context, fid feedbackid.ID) error {
	return nil
}
