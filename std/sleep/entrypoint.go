package sleep

import (
	"context"
	"io"
	"time"

	"github.com/cofunclabs/cofunc/manifest"
)

var _manifest = manifest.Manifest{
	Name:           "sleep",
	Description:    "Used to pause the program for a period of time",
	Driver:         "go",
	EntryPoint:     "",
	EntrypointFunc: Entrypoint,
	Args: map[string]string{
		"time": "1s",
	},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{},
		ReturnValues: []manifest.UsageDesc{},
	},
}

func New() *manifest.Manifest {
	return &_manifest
}

func Entrypoint(ctx context.Context, out io.Writer, version string, args map[string]string) (map[string]string, error) {
	time.Sleep(time.Second)
	return nil, nil
}
