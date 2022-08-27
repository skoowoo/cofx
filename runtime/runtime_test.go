package runtime

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cofunclabs/cofunc/pkg/nameid"
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
	id := nameid.New("testingdata.flowl")

	{
		err := sd.ParseFlow(ctx, id, strings.NewReader(testingdata))
		assert.NoError(t, err)

		var status StatusType
		err = sd.FetchFlow(ctx, id, func(b *FlowBody) error {
			status = b.status
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, StatusAdded, status)
	}

	{
		err := sd.InitFlow(ctx, id, GetStdoutLogger)
		assert.NoError(t, err)

		var status StatusType
		err = sd.FetchFlow(ctx, id, func(b *FlowBody) error {
			status = b.status
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, StatusReady, status)
	}

	{
		err := sd.ExecFlow(ctx, id)
		assert.NoError(t, err)

		time.Sleep(time.Second * 5)

		var status StatusType
		err = sd.FetchFlow(ctx, id, func(b *FlowBody) error {
			status = b.status
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, StatusStopped, status)
	}

	assert.Len(t, sd.store.entity, 1)
}
