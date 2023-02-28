package gitinsight

import (
	"context"
	"os"
	"testing"

	"github.com/skoowoo/cofx/functiondriver/go/spec"
	"github.com/skoowoo/cofx/service/resource"

	"github.com/stretchr/testify/assert"
)

func TestEntrypoint(t *testing.T) {
	_, call, _ := New()

	bundle := spec.EntrypointBundle{
		Version: "latest",
		Resources: resource.Resources{
			Logwriter: os.Stdout,
		},
	}
	returns, err := call(context.Background(), bundle, map[string]string{})
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Len(t, returns, 7)

	assert.NotEmpty(t, returns[branchCountRet.Name])
	assert.NotEmpty(t, returns[commitCountRet.Name])
	assert.NotEmpty(t, returns[lastCommitHeadRet.Name])
	assert.NotEmpty(t, returns[lastCommitMainRet.Name])
	assert.NotEmpty(t, returns[lastCommitOriginRet.Name])
	assert.NotEmpty(t, returns[lastCommitUpstreamRet.Name])
}
