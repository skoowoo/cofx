package runtime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/cofunclabs/cofunc/pkg/nameid"
	"github.com/cofunclabs/cofunc/runtime/actuator"
)

// Event is from the event trigger, it will be used to make the flow run
type Event struct {
	id      nameid.ID
	result  chan error
	execute func(id nameid.ID) error
}

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
	return nil
}

// InitFlow initialize the flow and make it into READY status.
func (rt *Runtime) InitFlow(ctx context.Context, id nameid.ID, opts ...FlowOption) error {
	flow, err := rt.store.get(id.ID())
	if err != nil {
		return err
	}
	if !flow.IsAdded() {
		return fmt.Errorf("not added: flow %s", id.ID())
	}

	ready := func(fb *FlowBody) error {
		// Initialize options of the flow
		for _, opt := range opts {
			opt(fb)
		}

		// Initialize all task nodes
		err := fb.runq.WalkNode(func(node actuator.Node) error {
			seq := node.(actuator.Task).Seq()

			fb.statistics[seq] = &functionStatistics{
				functionStatisticsBody: functionStatisticsBody{
					fid:    fb.id,
					node:   node,
					status: StatusReady,
				},
			}
			fb.progress.nodes = append(fb.progress.nodes, seq)

			// Initialize the function node, it will Load&Init the function driver
			logwriter, err := fb.createLogwriter(strconv.Itoa(seq))
			if err != nil {
				return err
			}
			resources := fb.copyResources()
			resources.Logwriter = logwriter
			return node.Init(ctx, actuator.WithResources(resources))
		})
		if err != nil {
			return err
		}

		// Initialize all triggers
		triggers := fb.runq.GetTriggers()
		for _, tg := range triggers {
			logwriter, err := fb.createLogwriter(strconv.Itoa(tg.(actuator.Task).Seq()))
			if err != nil {
				return err
			}
			resources := fb.copyResources()
			resources.Logwriter = logwriter
			if err := tg.Init(ctx, actuator.WithResources(resources)); err != nil {
				return err
			}
		}

		fb.status = StatusReady
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

// MustReay is a thin wrapper of Stopped2Ready
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
	triggers := flow.RunQ().GetTriggers()
	return len(triggers) != 0, nil
}

// StartEventTrigger start the event trigger of a flow, every event trigger function will run in a goroutine
// When a event triggeer returned without an error, will create and send a event to runtime
func (rt *Runtime) StartEventTrigger(ctx context.Context, id nameid.ID) error {
	flow, err := rt.store.get(id.ID())
	if err != nil {
		return err
	}
	triggers := flow.RunQ().GetTriggers()
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
func (rt *Runtime) ExecFlow(ctx context.Context, id nameid.ID) (err0 error) {
	flow, err := rt.store.get(id.ID())
	if err != nil {
		return err
	}
	if !flow.IsReady() {
		return fmt.Errorf("not ready: flow %s", id.ID())
	}

	flow.ToRuning()
	if err := flow.beforeFunc(id); err != nil {
		return err
	}
	defer func() {
		flow.ToStopped()
		if err := flow.afterFunc(id); err != nil {
			err0 = err
		}
	}()
	err = flow.RunQ().WalkAndExec(ctx, rt.execStepFunc(ctx, flow))
	if err != nil {
		return err
	}
	return nil
}

func (rt *Runtime) execStepFunc(ctx context.Context, f *Flow) func([]actuator.Node) error {
	return func(batch []actuator.Node) error {
		ch := make(chan *functionStatistics, len(batch))
		nodes := len(batch)

		// parallel run functions at the step
		for _, n := range batch {
			f.GetStatistics(n.(actuator.Task).Seq()).ToRuning()
			f.Refresh()

			go func(node actuator.Node) {
				fs := f.GetStatistics(node.(actuator.Task).Seq())
				retries := node.(actuator.Task).RetryOnFailure() + 1
				// Start to execute the function node, it will call the function driver to execute the function code
				for i := 0; i < retries; i++ {
					err := node.Exec(ctx)
					fs.ToStopped(err)

					// Retry if have an error
					if err != nil && err != actuator.ErrConditionIsFalse {
						continue
					} else {
						break
					}
				}
				// Send the result of the function execution to make it stopped really
				ch <- fs
			}(n)
		} // End of start batch

		// Waiting functions at the step to finish
		abortErr := make([]error, 0)
		for i := 0; i < nodes; i++ {
			fs := <-ch
			// Find the function node that executes with an error
			fs.WithLock(func(body *functionStatisticsBody) {
				ignore := body.node.(actuator.Task).IgnoreFailure()
				if body.err != nil && !ignore && !errors.Is(body.err, context.Canceled) {
					abortErr = append(abortErr, body.err)
				}
			})
			f.Refresh()
		}

		// Have an error at the step, so abort the flow
		if l := len(abortErr); l != 0 {
			return errors.New("occurred error at step: " + fmt.Sprintf("%+v", abortErr))
		}
		return nil
	}
}
