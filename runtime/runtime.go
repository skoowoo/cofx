package runtime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/cofunclabs/cofunc/config"
	"github.com/cofunclabs/cofunc/pkg/logfile"
	"github.com/cofunclabs/cofunc/pkg/nameid"
	"github.com/cofunclabs/cofunc/runtime/actuator"
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

func (rt *Runtime) ParseFlow(ctx context.Context, id nameid.ID, rd io.Reader) error {
	rq, ast, err := actuator.New(rd)
	if err != nil {
		return err
	}
	flow := newflow(id, rq, ast)
	if err := rt.store.store(id.Value(), flow); err != nil {
		return err
	}
	flow.WithLock(func(b *FlowBody) error {
		b.status = StatusAdded
		return nil
	})
	return nil
}

func (rt *Runtime) InitFlow(ctx context.Context, id nameid.ID) error {
	flow, err := rt.store.get(id.Value())
	if err != nil {
		return err
	}
	if !flow.IsAdded() {
		return fmt.Errorf("not added: flow %s", id.Value())
	}

	ready := func(body *FlowBody) error {
		body.status = StatusReady
		body.metrics = make(map[int]*functionMetrics)

		err := body.runq.ForfuncNode(func(node actuator.Node) error {
			seq := node.(actuator.Task).Seq()
			body.metrics[seq] = &functionMetrics{
				functionMetricsBody: functionMetricsBody{
					fid:    body.id,
					node:   node,
					status: StatusReady,
				},
			}
			body.progress.nodes = append(body.progress.nodes, seq)

			// Initialize the local logdir directory for the function/node in the flow
			logdir, err := config.LogFunctionDir(id.Value(), seq)
			if err != nil {
				return fmt.Errorf("%w: create function's log directory", err)
			}
			logger, err := logfile.TruncFile(config.LogFunctionFile(logdir))
			if err != nil {
				return fmt.Errorf("%w: create function's logger", err)
			}
			body.logger = logger

			// Initialize the function node, it will Load&Init the function driver
			if err := node.Init(ctx, actuator.WithLoad(logger)); err != nil {
				return err
			}
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

func (rt *Runtime) ExecFlow(ctx context.Context, id nameid.ID) error {
	flow, err := rt.store.get(id.Value())
	if err != nil {
		return err
	}
	if !flow.IsReady() {
		return fmt.Errorf("not ready: flow %s", id.Value())
	}

	execOneStep := func(batch []actuator.Node) error {
		ch := make(chan *functionMetrics, len(batch))
		nodes := len(batch)

		// parallel run functions at the step
		for _, n := range batch {
			metrics := flow.GetMetrics(n.(actuator.Task).Seq())
			metrics.WithLock(func(body *functionMetricsBody) {
				body.begin = time.Now()
				body.status = StatusRunning
			})
			flow.Refresh()

			go func(node actuator.Node, fm *functionMetrics) {
				// Start to execute the function node, it will call the function driver to execute the function code
				err := node.Exec(ctx)

				// Update the statistics of the function node execution
				fm.WithLock(func(body *functionMetricsBody) {
					body.err = err
					body.end = time.Now()
					body.duration = body.end.Sub(body.begin).Milliseconds()
					body.status = StatusStopped
					body.runs += 1

					if body.err != nil {
						if body.err == actuator.ErrConditionIsFalse {
							body.err = nil
							body.runs -= 1
						}
					}
				})
				select {
				case ch <- fm:
				case <-ctx.Done():
				}
			}(n, metrics)
		}
		flow.Refresh()

		// waiting functions at the step to finish running
		abortErr := make([]*functionMetrics, 0)
		for i := 0; i < nodes; i++ {
			select {
			case <-ctx.Done():
				// canced
				close(ch)
			case m := <-ch:
				// Find the function node that executes with an error
				m.WithLock(func(body *functionMetricsBody) {
					if body.err != nil {
						abortErr = append(abortErr, m)
					}
				})
				flow.Refresh()
			}
		}
		flow.Refresh()

		if l := len(abortErr); l != 0 {
			return errors.New("occurred error at step")
		}
		return nil
	}
	flow.WithLock(func(body *FlowBody) error {
		body.begin = time.Now()
		body.status = StatusRunning
		return nil
	})
	err = flow.GetRunQ().ForstepAndExec(ctx, execOneStep)
	flow.WithLock(func(body *FlowBody) error {
		body.status = StatusStopped
		body.duration = time.Since(body.begin).Milliseconds()
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// Stopped2Ready will reset the status of the flow and all nodes to ready, but only when all nodes are stopped
// When re-executing the flow, You need to call this method
func (rt *Runtime) Stopped2Ready(ctx context.Context, id nameid.ID) error {
	flow, err := rt.store.get(id.Value())
	if err != nil {
		return err
	}
	return flow.ToReady()
}

func (rt *Runtime) FetchFlow(ctx context.Context, id nameid.ID, do func(*FlowBody) error) error {
	flow, err := rt.store.get(id.Value())
	if err != nil {
		return err
	}
	return flow.WithLock(do)
}

func (rt *Runtime) StopFlow(ctx context.Context, id nameid.ID) error {
	return nil
}

func (rt *Runtime) DeleteFlow(ctx context.Context, id nameid.ID) error {
	return nil
}
