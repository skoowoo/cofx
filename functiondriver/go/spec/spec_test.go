package spec

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func IsAFunction(ctx context.Context, bundle EntrypointBundle, args EntrypointArgs) (map[string]string, error) {
	return nil, nil
}

func TestFunc2Name(t *testing.T) {
	f := IsAFunction
	name := Func2Name(f)
	assert.Equal(t, "github.com/cofunclabs/cofunc/functiondriver/go/spec.IsAFunction", name)
}
