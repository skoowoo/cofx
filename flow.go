//go:generate stringer -type=FlowStatus
package cofunc

import (
	"errors"
	"sync"
	"time"

	"github.com/cofunclabs/cofunc/pkg/feedbackid"
)

type FlowStatus int

const (
	FLOW_UNKNOWN FlowStatus = iota
	FLOW_STOPPED
	FLOW_RUNNING
	FLOW_READY
	FLOW_ERROR
	FLOW_ADDED
	FLOW_UPDATED
	FLOW_RUNNING_AND_UPDATED
	FLOW_DELETED
)

type FunctionResult struct {
	fid     feedbackid.ID
	node    *Node
	returns map[string]string
	begin   time.Time
	end     time.Time
	err     error
}

// Flow
//
type Flow struct {
	sync.RWMutex
	FlowBody
}

type FlowBody struct {
	id      feedbackid.ID
	status  FlowStatus
	begin   time.Time
	end     time.Time
	total   int
	ready   int
	results map[string]*FunctionResult

	runq *RunQueue
	ast  *AST
}

func (b *FlowBody) Runq() *RunQueue {
	return b.runq
}

func (b *FlowBody) BlockStore() *AST {
	return b.ast
}

func NewFlow(id feedbackid.ID, runq *RunQueue, ast *AST) *Flow {
	return &Flow{
		FlowBody: FlowBody{
			id:   id,
			runq: runq,
			ast:  ast,
		},
	}
}

func (f *Flow) ReadField(read ...func(FlowBody) error) error {
	f.RLock()
	defer f.RUnlock()
	for _, rd := range read {
		if err := rd(f.FlowBody); err != nil {
			return err
		}
	}
	return nil
}

func (f *Flow) UpdateField(update ...func(*FlowBody) error) error {
	f.Lock()
	defer f.Unlock()
	for _, up := range update {
		if err := up(&f.FlowBody); err != nil {
			return err
		}
	}
	return nil
}

func updateBeginTime(b *FlowBody) error {
	b.begin = time.Now()
	return nil
}

func updateEndTime(b *FlowBody) error {
	b.end = time.Now()
	return nil
}

func toRunning(b *FlowBody) error {
	if b.status == FLOW_RUNNING {
		return errors.New("function is running: " + b.id.Value())
	}
	if b.status == FLOW_ADDED {
		return errors.New("function is not ready: " + b.id.Value())
	}
	b.status = FLOW_RUNNING
	return nil
}

func toStopped(b *FlowBody) error {
	if b.status == FLOW_ERROR {
		// todo
	}
	if b.status == FLOW_RUNNING {
		b.status = FLOW_STOPPED
	}
	return nil
}

func toError(b *FlowBody) error {
	b.status = FLOW_ERROR
	return nil
}
