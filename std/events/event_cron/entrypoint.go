package eventcron

import (
	"context"
	"io"

	"github.com/cofunclabs/cofunc/manifest"
)

var exprArg = manifest.UsageDesc{
	Name:           "expr",
	OptionalValues: []string{},
	Desc:           "A cron expression, e.g. 0 0 * * *, 0 15 10 ? * *",
}

var _manifest = manifest.Manifest{
	Name:           "event_cron",
	Description:    "Used to trigger an event based on a cron expression",
	Driver:         "go",
	Entrypoint:     "",
	Args:           map[string]string{},
	RetryOnFailure: 0,
	IgnoreFailure:  false,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{exprArg},
		ReturnValues: []manifest.UsageDesc{}},
}

func New() (*manifest.Manifest, manifest.EntrypointFunc) {
	return &_manifest, Entrypoint
}

func Entrypoint(ctx context.Context, out io.Writer, version string, args manifest.EntrypointArgs) (map[string]string, error) {
	s := args.GetString(exprArg.Name)
	_ = s
	return nil, nil
}
