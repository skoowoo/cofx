package shelldriver

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cofxlabs/cofx/service/resource"
	"github.com/stretchr/testify/assert"
)

func TestShellDriver(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	os.Setenv("COFX_HOME", filepath.Join(wd, "testdata"))
	defer os.Unsetenv("COFX_HOME")

	var buf bytes.Buffer
	ctx := context.Background()

	driver := New("echo", "echo", "lastest")
	if err := driver.Load(ctx, resource.Resources{
		Logwriter: &buf,
	}); err != nil {
		assert.FailNow(t, err.Error())
	}
	rets, err := driver.Run(ctx, map[string]string{"message": "testing shell driver"})
	_ = rets
	assert.NoError(t, err)
	assert.Equal(t, "testing shell driver", strings.TrimSpace(buf.String()))
}
