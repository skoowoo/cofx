//go:generate stringer -type=FlowStatus
package flow

import (
	"errors"
	"sync"
	"time"

	"github.com/cofunclabs/cofunc/internal/flowl"
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
	FID          feedbackid.ID
	FNode        *flowl.FunctionNode
	ReturnValues map[string]string
	BeginTime    time.Time
	EndTime      time.Time
	Err          error
}

// Flow
//
type Flow struct {
	sync.RWMutex
	Body
}

type Body struct {
	ID           feedbackid.ID
	Status       FlowStatus
	BeginTime    time.Time
	EndTime      time.Time
	FnTotal      int
	ReadyFnCount int
	Results      map[string]*FunctionResult

	runq       *flowl.RunQueue
	blockstore *flowl.BlockStore
}

func (b *Body) Runq() *flowl.RunQueue {
	return b.runq
}

func (b *Body) BlockStore() *flowl.BlockStore {
	return b.blockstore
}

func New(id feedbackid.ID, runq *flowl.RunQueue, bs *flowl.BlockStore) *Flow {
	return &Flow{
		Body: Body{
			ID:         id,
			runq:       runq,
			blockstore: bs,
		},
	}
}

func (f *Flow) ReadField(read ...func(Body) error) error {
	f.RLock()
	defer f.RUnlock()
	for _, rd := range read {
		if err := rd(f.Body); err != nil {
			return err
		}
	}
	return nil
}

func (f *Flow) UpdateField(update ...func(*Body) error) error {
	f.Lock()
	defer f.Unlock()
	for _, up := range update {
		if err := up(&f.Body); err != nil {
			return err
		}
	}
	return nil
}

func UpdateBegineTime(b *Body) error {
	b.BeginTime = time.Now()
	return nil
}

func UpdateEndTime(b *Body) error {
	b.EndTime = time.Now()
	return nil
}

func ToRunning(b *Body) error {
	if b.Status == FLOW_RUNNING {
		return errors.New("function is running: " + b.ID.Value())
	}
	if b.Status == FLOW_ADDED {
		return errors.New("function is not ready: " + b.ID.Value())
	}
	b.Status = FLOW_RUNNING
	return nil
}

func ToStopped(b *Body) error {
	if b.Status == FLOW_ERROR {
		// todo
	}
	if b.Status == FLOW_RUNNING {
		b.Status = FLOW_STOPPED
	}
	return nil
}

func ToError(b *Body) error {
	b.Status = FLOW_ERROR
	return nil
}
