package eventtick

import (
	"context"
	"time"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/manifest"
)

var durationArg = manifest.UsageDesc{
	Name: "duration",
	Desc: "A time duration, e.g. 1s, 1m, 1h, 1m10s",
}

var _manifest = manifest.Manifest{
	Category:    "event",
	Name:        "event_tick",
	Description: "Used to trigger an event every X seconds",
	Driver:      "go",
	Args: map[string]string{
		"duration": "10s",
	},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{durationArg},
		ReturnValues: []manifest.UsageDesc{},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	s := args.GetString(durationArg.Name)
	v, err := time.ParseDuration(s)
	if err != nil {
		return nil, err
	}
	ticker := time.NewTicker(v)
	select {
	case <-ticker.C:
		return map[string]string{"which": _manifest.Name}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
