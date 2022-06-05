package scheduler

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/autoflowlabs/funcflow/internal/flowl"
	"github.com/autoflowlabs/funcflow/pkg/feedbackid"
	"github.com/sirupsen/logrus"
)

// Flow
//
type Flow struct {
	sync.RWMutex
	ID         feedbackid.ID
	runq       *flowl.RunQueue
	blockstore *flowl.BlockStore
	status     FlowStatus
}

func (f *Flow) SetStatus(set func(current FlowStatus) FlowStatus) {
	f.Lock()
	defer f.Unlock()
	f.status = set(f.status)
}

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

// FlowController
//
type FlowController struct {
	events    chan *Flow
	flowCount int
}

func (c *FlowController) Notify(f *Flow) error {
	select {
	case c.events <- f:
	case <-time.After(time.Millisecond * 500):
		return errors.New("can't notify flow runner: " + f.ID.Value())
	}
	return nil
}

func (c *FlowController) StartAndMonitoring(ctx context.Context) {
	for {
		select {
		case flow := <-c.events:
			_ = flow // todo
			c.flowCount += 1
			flow.ID.Feedback("Handled")
		case <-ctx.Done():
		}
	}
}

// Sched
//
type Sched struct {
	input chan struct {
		ID feedbackid.ID
		Rd io.Reader
	}
	flowstore  *FlowStore
	controller *FlowController
}

func NewSched() *Sched {
	s := &Sched{}
	s.input = make(chan struct {
		ID feedbackid.ID
		Rd io.Reader
	}, 5)
	s.flowstore = &FlowStore{
		entity: make(map[string]*Flow),
		notify: s,
	}
	s.controller = &FlowController{
		events: make(chan *Flow, 5),
	}
	return s
}

func (s *Sched) StartAndRun(ctx context.Context) error {
	// todo

	// goroutine is used to parse 'flowl' files, and then store it into flowstore
	go func(ctx context.Context) {
		for {
			select {
			case fr := <-s.input:
				rq, bs, err := flowl.Parse(fr.Rd)
				if err != nil {
					e := "failed to parse flowl, " + err.Error()
					logrus.Errorln(e)
					fr.ID.Feedback(e)
					continue
				}
				flow := &Flow{
					ID:         fr.ID,
					runq:       rq,
					blockstore: bs,
				}
				if err := s.flowstore.Store(flow.ID.Value(), flow); err != nil {
					e := "failed to store flow, " + err.Error()
					logrus.Errorln(e)
					flow.ID.Feedback(e)
					continue
				}
			case <-ctx.Done():
				logrus.Infoln("goroutine exit parsing flowl")
				return
			}
		}
	}(ctx)

	// goroutine is used to run and manage the all flows in flowstore
	go s.controller.StartAndMonitoring(ctx)

	return nil
}

func (s *Sched) Stop(ctx context.Context) {

}

func (s *Sched) ForceStop(ctx context.Context) {

}

func (s *Sched) InputFlowL(fid feedbackid.ID, rd io.Reader) {
	s.input <- struct {
		ID feedbackid.ID
		Rd io.Reader
	}{
		ID: fid,
		Rd: rd,
	}
	fid.Feedback("Inputed scheduler")
}

// Added is a notify function, a new flow is added to the flowstore
func (s *Sched) Added(f *Flow) {
	f.SetStatus(func(status FlowStatus) FlowStatus {
		return _FLOW_ADDED
	})

	if err := s.controller.Notify(f); err != nil {
		logrus.Error(err)
		f.ID.Feedback(err.Error())
		return
	}
	f.ID.Feedback("Notified controller")
}

// Updated is a notify function, a flow is updated
func (s *Sched) Updated(f *Flow) {

}

// Deleted is a notify function, a flow is deleted from the flowstore
func (s *Sched) Deleted(f *Flow) {

}
