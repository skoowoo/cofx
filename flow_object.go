//go:generate stringer -type=FlowStatus
package cofunc

import (
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
	fid          feedbackid.ID
	fnode        *flowl.FunctionNode
	returnValues map[string]string
	beginTime    time.Time
	endTime      time.Time
	err          error
}

// Flow
//
type Flow struct {
	sync.RWMutex
	ID           feedbackid.ID
	runq         *flowl.RunQueue
	blockstore   *flowl.BlockStore
	status       FlowStatus
	beginTime    time.Time
	endTime      time.Time
	fnTotal      int
	readyFnCount int
	successCount int
	result       map[string]*FunctionResult
}
