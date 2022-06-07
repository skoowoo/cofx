package funcflow

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/autoflowlabs/funcflow/pkg/feedbackid"
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

		flow, err := ctrl.InspectFlow(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, FLOW_ADDED, flow.status)
	}

	{
		err := ctrl.ReadyFlow(ctx, id)
		assert.NoError(t, err)

		flow, err := ctrl.InspectFlow(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, FLOW_READY, flow.status)
	}

	{
		err := ctrl.StartFlow(ctx, id)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		flow, err := ctrl.InspectFlow(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, FLOW_STOPPED, flow.status)
	}

	assert.Len(t, ctrl.flowstore.entity, 1)
}
