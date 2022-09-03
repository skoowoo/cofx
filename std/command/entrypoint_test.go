package command

import (
	"context"
	"os"
	"testing"

	"github.com/cofunclabs/cofunc/functiondriver/go/spec"
	"github.com/cofunclabs/cofunc/service/resource"
	"github.com/stretchr/testify/assert"
)

func TestCommandFunction(t *testing.T) {
	mf, ep, _ := New()
	assert.Equal(t, "go", mf.Driver)
	bundle := spec.EntrypointBundle{
		Version: "latest",
		Resources: resource.Resources{
			Logwriter: os.Stdout,
		},
	}
	_, err := ep(context.Background(), bundle, map[string]string{
		"cmd": "echo hello cofunc && sleep 2 && echo hello cofunc2",
	})
	assert.NoError(t, err)
}
