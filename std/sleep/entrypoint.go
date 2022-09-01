package sleep

import (
	"context"
	"time"

	"github.com/cofunclabs/cofunc/functiondriver/go/spec"
	"github.com/cofunclabs/cofunc/manifest"
)

var durationArg = manifest.UsageDesc{
	Name: "duration",
	Desc: "Specify a duration to sleep, default 1s",
}

var _manifest = manifest.Manifest{
	Name:        "sleep",
	Description: "Used to pause the program for a period of time",
	Driver:      "go",
	Args: map[string]string{
		"duration": "1s",
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
		return nil, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
