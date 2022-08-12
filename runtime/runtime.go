package runtime

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/cofunclabs/cofunc/generator"
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
	rq, ast, err := generator.New(rd)
	if err != nil {
		return err
	}
	flow := newflow(fid, rq, ast)
	if err := sd.store.store(fid.Value(), flow); err != nil {
		return err
	}
	flow.WithLock(func(b *flowBody) error {
		b.status = _flow_added
		return nil
	})
	return nil
}

func (sd *Sched) ReadyFlow(ctx context.Context, fid feedbackid.ID) error {
	flow, err := sd.store.get(fid.Value())
	if err != nil {
		return err
	}

	ready := func(body *flowBody) error {
		if body.status == _flow_ready || body.status == _flow_running {
			return nil
		}
		body.results = make(map[int]*FunctionResult)
		err := body.runq.ForfuncNode(func(stage int, n generator.Node) error {
			if err := n.Init(ctx); err != nil {
				return err
			}
			body.results[n.(generator.NodeExtend).Seq()] = &FunctionResult{
				functionResultBody: functionResultBody{
					fid:    body.id,
					node:   n,
					status: _flow_ready,
				},
			}
			return nil
		})
		if err != nil {
			return err
		}
		body.status = _flow_ready
		body.progress.total = len(body.results)
		return nil
	}

	if err := flow.WithLock(ready); err != nil {
		return err
	}
	if err := flow.Refresh(); err != nil {
		return err
	}
	return nil
}

func (sd *Sched) StartFlow(ctx context.Context, fid feedbackid.ID) error {
	flow, err := sd.store.get(fid.Value())
	if err != nil {
		return err
	}

	execOneStep := func(batch []generator.Node) error {
		ch := make(chan *FunctionResult, len(batch))
		// parallel run functions at the step
		for _, node := range batch {
			fr := flow.GetResult(node.(generator.NodeExtend).Seq())
			fr.WithLock(func(body *functionResultBody) {
				body.begin = time.Now()
				body.status = _flow_running
			})

			go func(n generator.Node, fr *FunctionResult) {
				err := n.Exec(ctx)
				fr.WithLock(func(body *functionResultBody) {
					body.err = err
				})
				select {
				case ch <- fr:
				case <-ctx.Done():
				}
			}(node, fr)
		}

		// refresh flow
		flow.Refresh()

		// waiting functions at the step to finish running
		errResults := make([]*FunctionResult, 0)
		for i := 0; i < len(batch); i++ {
			select {
			case <-ctx.Done():
				// canced
				close(ch)
			case r := <-ch:
				r.WithLock(func(body *functionResultBody) {
					body.end = time.Now()
					body.status = _flow_stopped
					body.executed = true
					body.runs += 1

					if body.err != nil {
						if body.err == generator.ErrConditionIsFalse {
							body.err = nil
							body.executed = false
							body.runs -= 1
						} else {
							body.status = _flow_error
						}
					}
				})
				if r.err != nil {
					errResults = append(errResults, r)
				}
				flow.Refresh()
			}
		}
		flow.Refresh()

		if l := len(errResults); l != 0 {
			return errors.New("occurred error at step")
		}
		return nil
	}
	if err := flow.GetRunQ().ForstepAndExec(ctx, execOneStep); err != nil {
		return err
	}
	return nil
}

func (sd *Sched) InspectFlow(ctx context.Context, fid feedbackid.ID, read func(*flowBody) error) error {
	flow, err := sd.store.get(fid.Value())
	if err != nil {
		return err
	}
	return flow.WithLock(read)
}

func (sd *Sched) StopFlow(ctx context.Context, fid feedbackid.ID) error {
	return nil
}

func (sd *Sched) DeleteFlow(ctx context.Context, fid feedbackid.ID) error {
	return nil
}
