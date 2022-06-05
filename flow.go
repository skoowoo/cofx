package funcflow

import (
	"sync"

	"github.com/autoflowlabs/funcflow/internal/flowl"
	"github.com/autoflowlabs/funcflow/pkg/feedbackid"
)

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
