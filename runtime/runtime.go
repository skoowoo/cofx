package runtime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/cofunclabs/cofunc/config"
	"github.com/cofunclabs/cofunc/pkg/logfile"
	"github.com/cofunclabs/cofunc/pkg/nameid"
	"github.com/cofunclabs/cofunc/runtime/actuator"
)

type GetLogger func(actuator.Node) (*logfile.Logfile, error)

// GetStdoutLogger returns a stdout logger to use
func GetStdoutLogger(node actuator.Node) (*logfile.Logfile, error) {
	return logfile.Stdout(), nil
}

// GetDefaultLogger returns a default logger to use, the defualt logger is output to a file
func GetDefaultLogger(id nameid.ID) GetLogger {
	return func(node actuator.Node) (*logfile.Logfile, error) {
		seq := node.(actuator.Task).Seq()
		// Initialize the local logdir directory for the function/node in the flow
		logdir, err := config.LogFunctionDir(id.ID(), seq)
		if err != nil {
			return nil, fmt.Errorf("%w: create function's log directory", err)
		}
		logger, err := logfile.TruncFile(config.LogFunctionFile(logdir))
		if err != nil {
			return nil, fmt.Errorf("%w: create function's logger", err)
		}
		return logger, nil
	}
}

// Event is from the event trigger, it will be used to make the flow run
type Event struct {
	id      nameid.ID
	result  chan error
	execute func(id nameid.ID) error
}

// Runtime
//
type Runtime struct {
	store  *flowstore
	events chan Event
}

func New() *Runtime {
	r := &Runtime{
		events: make(chan Event, 64),
	}
	r.store = &flowstore{
		entity: make(map[string]*Flow),
	}

	// start a watcher goroutine to waiting and handling event from triggers. The watcher goroutine
	// don't finished forever.
	go r.startEventWatcher()
	return r
}

// ParseFlow parse one flowl source file, and add a flow into runtime, the argument 'rd' is a reader for
// a flow source file.
// After invoking this method, the flow's status is ADDED.
func (rt *Runtime) ParseFlow(ctx context.Context, id nameid.ID, rd io.Reader) error {
	rq, ast, err := actuator.New(rd)
	if err != nil {
		return err
	}
	flow := newflow(id, rq, ast)
	if err := rt.store.store(id.ID(), flow); err != nil {
		return err
	}
	flow.WithLock(func(b *FlowBody) error {
		b.status = StatusAdded
		return nil
	})
	return nil
}

// InitFlow initialize the flow and make it into READY status.
func (rt *Runtime) InitFlow(ctx context.Context, id nameid.ID, getlogger GetLogger) error {
	flow, err := rt.store.get(id.ID())
	if err != nil {
		return err
	}
	if !flow.IsAdded() {
		return fmt.Errorf("not added: flow %s", id.ID())
	}

	ready := func(body *FlowBody) error {
		body.status = StatusReady
		body.metrics = make(map[int]*functionMetrics)

		// Initialize all task nodes
		err := body.runq.WalkNode(func(node actuator.Node) error {
			seq := node.(actuator.Task).Seq()
			body.metrics[seq] = &functionMetrics{
				functionMetricsBody: functionMetricsBody{
					fid:    body.id,
					node:   node,
					status: StatusReady,
				},
			}
			body.progress.nodes = append(body.progress.nodes, seq)

			logger, err := getlogger(node)
			if err != nil {
				return err
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

		// Initialize all triggers
		triggers := body.runq.GetTriggers()
		for _, tg := range triggers {
			logger, err := getlogger(tg)
			if err != nil {
				return err
			}
			if err := tg.Init(ctx, actuator.WithLoad(logger)); err != nil {
				return err
			}
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

// Stopped2Ready will reset the status of the flow and all nodes to ready, but only when all nodes are stopped
// When re-executing the flow, You need to call this method
func (rt *Runtime) Stopped2Ready(ctx context.Context, id nameid.ID) error {
	flow, err := rt.store.get(id.ID())
	if err != nil {
		return err
	}
	return flow.ToReady()
}

// MustReay is a simple wrapper of Stopped2Ready
func (rt *Runtime) MustReady(ctx context.Context, id nameid.ID) error {
	return rt.Stopped2Ready(ctx, id)
}

// FetchFlow get a flow, then access or handle it safety by the callback function
func (rt *Runtime) FetchFlow(ctx context.Context, id nameid.ID, do func(*FlowBody) error) error {
	flow, err := rt.store.get(id.ID())
	if err != nil {
		return err
	}
	return flow.WithLock(do)
}

// CancelFlow cancel the flow and make it into CANCELED status.
func (rt *Runtime) CancelFlow(ctx context.Context, id nameid.ID) error {
	return nil
}

func (rt *Runtime) DeleteFlow(ctx context.Context, id nameid.ID) error {
	return nil
}

// HasTrigger check the flow has a trigger or not
func (rt *Runtime) HasTrigger(id nameid.ID) (bool, error) {
	flow, err := rt.store.get(id.ID())
	if err != nil {
		return false, err
	}
	triggers := flow.GetRunQ().GetTriggers()
	return len(triggers) != 0, nil
}

// StartEventTrigger start the event trigger of a flow, every event trigger function will run in a goroutine
// When a event triggeer returned without an error, will create and send a event to runtime
func (rt *Runtime) StartEventTrigger(ctx context.Context, id nameid.ID) error {
	flow, err := rt.store.get(id.ID())
	if err != nil {
		return err
	}
	triggers := flow.GetRunQ().GetTriggers()
	n := len(triggers)
	if n == 0 {
		return nil
	}
	var wg sync.WaitGroup
	wg.Add(n)
	for _, tg := range triggers {
		go func(trigger actuator.Trigger) {
			defer wg.Done()
			errNum := 0
			for {
				err := trigger.Exec(ctx)
				if errors.Is(err, context.Canceled) {
					return
				}
				if err != nil {
					// TODO: log
					log.Println(err)
					errNum++
					if errNum <= 5 {
						time.Sleep(time.Second * time.Duration(errNum))
					} else if errNum <= 10 {
						time.Sleep(time.Second * time.Duration(errNum*2))
					} else if errNum <= 15 {
						time.Sleep(time.Second * time.Duration(errNum*5))
					} else {
						time.Sleep(time.Second * 300)
					}
					continue
				}
				// trigger returns without an error, it's success
				errNum = 0
				ev := Event{
					id:     id,
					result: make(chan error, 1),
					execute: func(id nameid.ID) error {
						if err := rt.MustReady(ctx, id); err != nil {
							return err
						}
						if err := rt.ExecFlow(ctx, id); err != nil {
							return err
						}
						return nil
					},
				}
				select {
				case rt.events <- ev:
					// wait for the result of the flow execution, avoid to run flow repeated at
					// the same time
					err := <-ev.result
					if err != nil {
						// TODO:
						log.Println(err)
					}
				case <-ctx.Done():
					return
				}
			}
		}(tg)
	}
	wg.Wait()
	return nil
}

func (rt *Runtime) startEventWatcher() error {
	for {
		ev := <-rt.events
		// Use a goroutine to execute the flow, avoid to block the event trigger, because
		// the flow may be run for a long time.
		go func() {
			ev.result <- ev.execute(ev.id)
			close(ev.result)
		}()
	}
}

// ExecFlow execute a flow step by step.
func (rt *Runtime) ExecFlow(ctx context.Context, id nameid.ID) error {
	flow, err := rt.store.get(id.ID())
	if err != nil {
		return err
	}
	if !flow.IsReady() {
		return fmt.Errorf("not ready: flow %s", id.ID())
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
				retries := node.(actuator.Task).RetryOnFailure() + 1
				// Start to execute the function node, it will call the function driver to execute the function code
				for i := 0; i < retries; i++ {
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

					// Retry if have an error
					if err != nil && err != actuator.ErrConditionIsFalse {
						continue
					} else {
						break
					}
				}
				// Send the result of the function execution to make it stopped really
				ch <- fm
			}(n, metrics)
		} // End of start batch
		flow.Refresh()

		// Waiting functions at the step to finish
		abortErr := make([]*functionMetrics, 0)
		for i := 0; i < nodes; i++ {
			m := <-ch
			// Find the function node that executes with an error
			m.WithLock(func(body *functionMetricsBody) {
				ignore := body.node.(actuator.Task).IgnoreFailure()
				if body.err != nil && !ignore && !errors.Is(body.err, context.Canceled) {
					abortErr = append(abortErr, m)
				}
			})
			flow.Refresh()
		}
		flow.Refresh()

		// Have an error at the step, so abort the flow
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
	err = flow.GetRunQ().WalkAndExec(ctx, execOneStep)
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
