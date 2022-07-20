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
	_flow_unknown FlowStatus = iota
	_flow_stopped
	_flow_running
	_flow_ready
	_flow_error
	_flow_added
	_flow_updated
	_flow_running_and_updated
	_flow_deleted
)

type FunctionResult struct {
	fid     feedbackid.ID
	node    *FuncNode
	returns map[string]string
	begin   time.Time
	end     time.Time
	err     error
}

// Flow
//
type Flow struct {
	sync.RWMutex
	flowBody
}

type flowBody struct {
	id      feedbackid.ID
	status  FlowStatus
	begin   time.Time
	end     time.Time
	total   int
	ready   int
	results map[string]*FunctionResult

	runq *RunQ
	ast  *AST
}

func (b *flowBody) GetRunQ() *RunQ {
	return b.runq
}

func (b *flowBody) GetAST() *AST {
	return b.ast
}

func newflow(id feedbackid.ID, runq *RunQ, ast *AST) *Flow {
	return &Flow{
		flowBody: flowBody{
			id:   id,
			runq: runq,
			ast:  ast,
		},
	}
}

func (f *Flow) readField(read ...func(flowBody) error) error {
	f.RLock()
	defer f.RUnlock()
	for _, rd := range read {
		if err := rd(f.flowBody); err != nil {
			return err
		}
	}
	return nil
}

func (f *Flow) updateField(update ...func(*flowBody) error) error {
	f.Lock()
	defer f.Unlock()
	for _, up := range update {
		if err := up(&f.flowBody); err != nil {
			return err
		}
	}
	return nil
}

func updateBeginTime(b *flowBody) error {
	b.begin = time.Now()
	return nil
}

func updateEndTime(b *flowBody) error {
	b.end = time.Now()
	return nil
}

func toRunning(b *flowBody) error {
	if b.status == _flow_running {
		return errors.New("function is running: " + b.id.Value())
	}
	if b.status == _flow_added {
		return errors.New("function is not ready: " + b.id.Value())
	}
	b.status = _flow_running
	return nil
}

func toStopped(b *flowBody) error {
	if b.status == _flow_error {
		// TODO:
	}
	if b.status == _flow_running {
		b.status = _flow_stopped
	}
	return nil
}

func toError(b *flowBody) error {
	b.status = _flow_error
	return nil
}
