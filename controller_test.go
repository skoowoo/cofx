package cofunc

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cofunclabs/cofunc/internal/flow"
	"github.com/cofunclabs/cofunc/pkg/feedbackid"
	"github.com/stretchr/testify/assert"
)

func TestAddReadyStartFlow(t *testing.T) {
	const testingdata string = `
	load go:print
	load go:sleep

	set print {
		input k1 v1
		input k2 v2
	}

	run @print
	run	@sleep
	`

	ctrl := NewFlowController()

	ctx := context.Background()
	id := feedbackid.NewDefaultID("testingdata.flowl")

	{
		err := ctrl.AddFlow(ctx, id, strings.NewReader(testingdata))
		assert.NoError(t, err)

		var status flow.FlowStatus
		err = ctrl.InspectFlow(ctx, id, func(b flow.Body) error {
			status = b.Status
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, flow.FLOW_ADDED, status)
	}

	{
		err := ctrl.ReadyFlow(ctx, id)
		assert.NoError(t, err)

		var status flow.FlowStatus
		err = ctrl.InspectFlow(ctx, id, func(b flow.Body) error {
			status = b.Status
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, flow.FLOW_READY, status)
	}

	{
		err := ctrl.StartFlow(ctx, id)
		assert.NoError(t, err)

		time.Sleep(time.Second * 5)

		var status flow.FlowStatus
		err = ctrl.InspectFlow(ctx, id, func(b flow.Body) error {
			status = b.Status
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, flow.FLOW_STOPPED, status)
	}

	assert.Len(t, ctrl.flowstore.entity, 1)
}
