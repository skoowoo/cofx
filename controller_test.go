package cofunc

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cofunclabs/cofunc/pkg/feedbackid"
	"github.com/stretchr/testify/assert"
)

func TestAddReadyStartFlow(t *testing.T) {
	const testingdata string = `
	load go:print
	load go:sleep

	fn p = print {
		args = {
			k1: v1
			k2: v2
		}
	}

	run p
	run	sleep
	`

	ctrl := NewController()

	ctx := context.Background()
	id := feedbackid.NewDefaultID("testingdata.flowl")

	{
		err := ctrl.AddFlow(ctx, id, strings.NewReader(testingdata))
		assert.NoError(t, err)

		var status FlowStatus
		err = ctrl.InspectFlow(ctx, id, func(b FlowBody) error {
			status = b.status
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, FLOW_ADDED, status)
	}

	{
		err := ctrl.ReadyFlow(ctx, id)
		assert.NoError(t, err)

		var status FlowStatus
		err = ctrl.InspectFlow(ctx, id, func(b FlowBody) error {
			status = b.status
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, FLOW_READY, status)
	}

	{
		err := ctrl.StartFlow(ctx, id)
		assert.NoError(t, err)

		time.Sleep(time.Second * 5)

		var status FlowStatus
		err = ctrl.InspectFlow(ctx, id, func(b FlowBody) error {
			status = b.status
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, FLOW_STOPPED, status)
	}

	assert.Len(t, ctrl.store.entity, 1)
}
