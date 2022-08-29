package command

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandFunction(t *testing.T) {
	mf, ep := New()
	assert.Equal(t, "go", mf.Driver)
	_, err := ep(context.Background(), os.Stdout, "latest", map[string]string{
		"script": "echo hello cofunc && sleep 2 && echo hello cofunc2",
	})
	assert.NoError(t, err)
}
