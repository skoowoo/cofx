package runtime

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cofxlabs/cofx/pkg/nameid"
	"github.com/cofxlabs/cofx/service/resource"
	"github.com/stretchr/testify/assert"
)

func TestCancelFlow(t *testing.T) {
	const testingdata string = `
load "go:print"
for {
	sleep "5s"
	co print{
		"_": "sleep 5s"
	}
}
	`

	rt := New()
	id := nameid.New("testingdata.flowl")
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		e := expect{
			nodes:      1,
			output:     "",
			hasExecErr: true,
		}
		newTestingFlowCase(t, rt, testingdata, id, e)
	}()
	time.Sleep(2 * time.Second)
	err := rt.CancelFlow(context.Background(), id)
	assert.NoError(t, err)
	wg.Wait()
}

func TestHelloWorld(t *testing.T) {
	const testingdata string = `
// An example of Hello World
load "go:print"

var s = "hello world!!!"

co print {
    "_" : "$(s)"
}
	`

	rt := New()
	id := nameid.New("testingdata.flowl")
	e := expect{
		nodes:  1,
		output: "hello world!!!",
	}
	newTestingFlowCase(t, rt, testingdata, id, e)
}

func TestAddReadyStartFlow(t *testing.T) {
	const testingdata string = `
	load "go:print"

	fn p = print {
		args = {
			"k1": "v1"
			"k2": "v2"
		}
	}

	co p
	`

	rt := New()

	ctx := context.Background()
	id := nameid.New("testingdata.flowl")

	{
		err := rt.ParseFlow(ctx, id, strings.NewReader(testingdata))
		assert.NoError(t, err)

		var status StatusType
		err = rt.FetchFlow(ctx, id, func(b *FlowBody) error {
			status = b.status
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, StatusAdded, status)
	}

	{
		err := rt.InitFlow(ctx, id)
		assert.NoError(t, err)

		var status StatusType
		err = rt.FetchFlow(ctx, id, func(b *FlowBody) error {
			status = b.status
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, StatusReady, status)
	}

	{
		err := rt.ExecFlow(ctx, id)
		assert.NoError(t, err)

		var status StatusType
		err = rt.FetchFlow(ctx, id, func(b *FlowBody) error {
			status = b.status
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, StatusStopped, status)
	}

	assert.Len(t, rt.store.entity, 1)
}

type expect struct {
	nodes      int
	output     string
	hasExecErr bool
}

func newTestingFlowCase(t *testing.T, rt *Runtime, testingdata string, id nameid.ID, e expect) {
	ctx := context.Background()
	var out bytes.Buffer

	{
		if err := rt.ParseFlow(ctx, id, strings.NewReader(testingdata)); err != nil {
			assert.FailNow(t, err.Error())
		}

		if flow, err := rt.store.get(id.ID()); err != nil {
			assert.FailNow(t, err.Error())
		} else {
			assert.Equal(t, StatusAdded, flow.status)
			assert.Equal(t, id.Name(), flow.id.Name())
			assert.Equal(t, id.ID(), flow.id.ID())
			assert.NotNil(t, flow.beforeFunc)
			assert.NotNil(t, flow.afterFunc)
			assert.NotNil(t, flow.createLogwriter)
			assert.NotNil(t, flow.copyResources)

			assert.Nil(t, flow.cancel)
		}
	}

	{
		createLogWriter := func(writerid string) (io.Writer, error) {
			return &out, nil
		}
		beforeExec := func(id nameid.ID) error {
			return nil
		}
		afterExec := func(id nameid.ID) error {
			return nil
		}
		copy := func() resource.Resources {
			return resource.Resources{}
		}
		var opts = []FlowOption{
			WithBeforeFunc(beforeExec),
			WithAfterFunc(afterExec),
			WithCopyResources(copy),
			WithCreateLogwriter(createLogWriter),
		}
		if err := rt.InitFlow(ctx, id, opts...); err != nil {
			assert.FailNow(t, err.Error())
		} else {
			if flow, err := rt.store.get(id.ID()); err != nil {
				assert.FailNow(t, err.Error())
			} else {
				assert.Equal(t, StatusReady, flow.status)
				assert.NotNil(t, flow.beforeFunc)
				assert.NotNil(t, flow.afterFunc)
				assert.NotNil(t, flow.createLogwriter)
				assert.NotNil(t, flow.copyResources)

				assert.Len(t, flow.statistics, e.nodes)
				assert.Len(t, flow.progress.nodes, e.nodes)

				assert.Nil(t, flow.cancel)
			}
		}
	}

	{
		ctx, cancel := context.WithCancel(context.Background())
		rt.FetchFlow(ctx, id, func(b *FlowBody) error {
			b.SetCancel(cancel)
			return nil
		})
		if err := rt.ExecFlow(ctx, id); err != nil {
			if !e.hasExecErr {
				assert.FailNow(t, err.Error())
			}
		} else {
			if flow, err := rt.store.get(id.ID()); err != nil {
				assert.FailNow(t, err.Error())
			} else {
				assert.Equal(t, StatusStopped, flow.status)
				assert.NotNil(t, flow.beforeFunc)
				assert.NotNil(t, flow.afterFunc)
				assert.NotNil(t, flow.createLogwriter)
				assert.NotNil(t, flow.copyResources)

				assert.Len(t, flow.statistics, e.nodes)
				assert.Len(t, flow.progress.nodes, e.nodes)
				assert.Len(t, flow.progress.done, e.nodes)
				assert.Len(t, flow.progress.running, 0)

				// assert.Equal(t, StatusStopped, flow.statistics[1000].status)
				// assert.Equal(t, nil, flow.statistics[1000].err)

				if e.output != "" {
					assert.Equal(t, e.output, strings.TrimSpace(out.String()))
				}
			}
		}
	}
}
