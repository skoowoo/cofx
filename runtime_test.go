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
	load "go:print"
	load "go:sleep"

	fn p = print {
		args = {
			"k1": "v1"
			"k2": "v2"
		}
	}

	co p
	co	sleep
	`

	sd := New()

	ctx := context.Background()
	id := feedbackid.NewDefaultID("testingdata.flowl")

	{
		err := sd.AddFlow(ctx, id, strings.NewReader(testingdata))
		assert.NoError(t, err)

		var status FlowStatus
		err = sd.InspectFlow(ctx, id, func(b flowBody) error {
			status = b.status
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, _flow_added, status)
	}

	{
		err := sd.ReadyFlow(ctx, id)
		assert.NoError(t, err)

		var status FlowStatus
		err = sd.InspectFlow(ctx, id, func(b flowBody) error {
			status = b.status
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, _flow_ready, status)
	}

	{
		err := sd.StartFlow(ctx, id)
		assert.NoError(t, err)

		time.Sleep(time.Second * 5)

		var status FlowStatus
		err = sd.InspectFlow(ctx, id, func(b flowBody) error {
			status = b.status
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, _flow_stopped, status)
	}

	assert.Len(t, sd.store.entity, 1)
}
