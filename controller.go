package funcflow

import (
	"context"
	"io"

	"github.com/autoflowlabs/funcflow/internal/flowl"
	"github.com/autoflowlabs/funcflow/pkg/feedbackid"
)

// FlowController
//
type FlowController struct {
	flowstore *FlowStore
	toStop    context.CancelFunc
}

func NewFlowController() *FlowController {
	fc := &FlowController{}
	fc.flowstore = &FlowStore{
		entity: make(map[string]*Flow),
	}
	return fc
}

func (fc *FlowController) AddFlow(ctx context.Context, fid feedbackid.ID, rd io.Reader) error {
	rq, bs, err := flowl.Parse(rd)
	if err != nil {
		return err
	}
	flow := &Flow{
		ID:         fid,
		runq:       rq,
		blockstore: bs,
		status:     FLOW_ADDED,
	}
	if err := fc.flowstore.Store(flow.ID.Value(), flow); err != nil {
		return err
	}
	return nil
}

func (fc *FlowController) ReadyFlow(ctx context.Context, fid feedbackid.ID) error {
	flow, err := fc.flowstore.Get(fid.Value())
	if err != nil {
		return err
	}
	return flow.Ready(ctx)
}

// StartFlow need to call 'flow.Execute' on a groutine
func (fc *FlowController) StartFlow(ctx context.Context, fid feedbackid.ID) error {
	flow, err := fc.flowstore.Get(fid.Value())
	if err != nil {
		return err
	}
	go func() {
		nctx, cancel := context.WithCancel(context.Background())
		flow.SetWithLock(func(s *Flow) {
			s.cancel = cancel
		})
		flow.ExecuteAndWaitFunc(nctx)
	}()
	return nil
}

func (fc *FlowController) InspectFlow(ctx context.Context, fid feedbackid.ID) (*Flow, error) {
	return fc.flowstore.Get(fid.Value())
}

func (fc *FlowController) StopFlow(ctx context.Context, fid feedbackid.ID) error {
	return nil
}

func (fc *FlowController) DeleteFlow(ctx context.Context, fid feedbackid.ID) error {
	return nil
}
