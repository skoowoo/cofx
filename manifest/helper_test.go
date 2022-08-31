package manifest

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func IsAFunction(ctx context.Context, w io.Writer, version string, args EntrypointArgs) (map[string]string, error) {
	return nil, nil
}

func TestFunc2Name(t *testing.T) {
	f := IsAFunction
	name := Func2Name(f)
	assert.Equal(t, "github.com/cofunclabs/cofunc/manifest.IsAFunction", name)
}
