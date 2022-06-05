package funcflow

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/autoflowlabs/funcflow/pkg/feedbackid"
	"github.com/stretchr/testify/assert"
)

func TestAddFlow(t *testing.T) {
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
	sched := NewSched()
	err := sched.StartAndRun(context.TODO())
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	id := feedbackid.NewDefaultID("testingdata.flowl")
	sched.InputFlowL(id, strings.NewReader(testingdata))

	t.Log("feedbackid", id.String())
	time.Sleep(time.Second)
	t.Log("feedbackid", id.String())

	flow, err := sched.flowstore.Get(id.Value())
	assert.NoError(t, err)
	assert.Equal(t, _FLOW_STOPPED, flow.status)

	assert.Len(t, sched.flowstore.entity, 1)

	assert.Equal(t, 1, sched.controller.flowCount)
}
