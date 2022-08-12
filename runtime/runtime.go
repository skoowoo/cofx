package runtime

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/cofunclabs/cofunc/generator"
	"github.com/cofunclabs/cofunc/pkg/feedbackid"
)

// Runtime
//
type Runtime struct {
	store *flowstore
}

func New() *Runtime {
	r := &Runtime{}
	r.store = &flowstore{
		entity: make(map[string]*Flow),
	}
	return r
}

func (rt *Runtime) AddFlow(ctx context.Context, fid feedbackid.ID, rd io.Reader) error {
	rq, ast, err := generator.New(rd)
	if err != nil {
		return err
	}
	flow := newflow(fid, rq, ast)
	if err := rt.store.store(fid.Value(), flow); err != nil {
		return err
	}
	flow.WithLock(func(b *FlowBody) error {
		b.status = _flow_added
		return nil
	})
	return nil
}

func (rt *Runtime) ReadyFlow(ctx context.Context, fid feedbackid.ID) error {
	flow, err := rt.store.get(fid.Value())
	if err != nil {
		return err
	}

	ready := func(body *FlowBody) error {
		if body.status == _flow_ready || body.status == _flow_running {
			return nil
		}
		body.status = _flow_ready
		body.results = make(map[int]*functionResult)
		err := body.runq.ForfuncNode(func(stage int, n generator.Node) error {
			if err := n.Init(ctx); err != nil {
				return err
			}
			seq := n.(generator.NodeExtend).Seq()
			body.results[seq] = &functionResult{
				functionResultBody: functionResultBody{
					fid:    body.id,
					node:   n,
					status: _flow_ready,
				},
			}
			body.progress.nodes = append(body.progress.nodes, seq)
			return nil
		})
		if err != nil {
			return err
		}
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

func (rt *Runtime) StartFlow(ctx context.Context, fid feedbackid.ID) error {
	flow, err := rt.store.get(fid.Value())
	if err != nil {
		return err
	}

	execOneStep := func(batch []generator.Node) error {
		ch := make(chan *functionResult, len(batch))
		// parallel run functions at the step
		for _, node := range batch {
			fr := flow.GetResult(node.(generator.NodeExtend).Seq())
			fr.WithLock(func(body *functionResultBody) {
				body.begin = time.Now()
				body.status = _flow_running
			})

			go func(n generator.Node, fr *functionResult) {
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
		errResults := make([]*functionResult, 0)
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

func (rt *Runtime) InspectFlow(ctx context.Context, fid feedbackid.ID, read func(*FlowBody) error) error {
	flow, err := rt.store.get(fid.Value())
	if err != nil {
		return err
	}
	return flow.WithLock(read)
}

func (rt *Runtime) StopFlow(ctx context.Context, fid feedbackid.ID) error {
	return nil
}

func (rt *Runtime) DeleteFlow(ctx context.Context, fid feedbackid.ID) error {
	return nil
}
