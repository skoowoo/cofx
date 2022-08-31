package eventtick

import (
	"context"
	"io"
	"strconv"
	"time"

	"github.com/cofunclabs/cofunc/manifest"
)

var _manifest = manifest.Manifest{
	Name:        "event_tick",
	Description: "Used to trigger an event every X seconds",
	Driver:      "go",
	Args: map[string]string{
		"seconds": "10",
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
	var secs int
	v := args["seconds"]
	secs, err := strconv.Atoi(v)
	if err != nil {
		return nil, err
	}
	ticker := time.NewTicker(time.Duration(secs) * time.Second)
	select {
	case <-ticker.C:
		return nil, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
