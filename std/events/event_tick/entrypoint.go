package eventtick

import (
	"context"
	"io"
	"time"

	"github.com/cofunclabs/cofunc/manifest"
)

var _manifest = manifest.Manifest{
	Name:        "event_tick",
	Description: "Used to trigger an event every X seconds",
	Driver:      "go",
	Args: map[string]string{
		"duration": "10s",
	},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{},
		ReturnValues: []manifest.UsageDesc{},
	},
}

func New() (*manifest.Manifest, manifest.EntrypointFunc) {
	return &_manifest, Entrypoint
}

func Entrypoint(ctx context.Context, out io.Writer, version string, args map[string]string) (map[string]string, error) {
	s := args["duration"]
	v, err := time.ParseDuration(s)
	if err != nil {
		return nil, err
	}
	ticker := time.NewTicker(v)
	select {
	case <-ticker.C:
		return nil, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
