package eventcron

import (
	"context"

	"github.com/cofunclabs/cofunc/functiondriver/go/spec"
	"github.com/cofunclabs/cofunc/manifest"
	cron "github.com/robfig/cron/v3"
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

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, func() spec.Customer { return &custom{waiting: make(chan struct{})} }
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	expr := args.GetString(exprArg.Name)
	custom := bundle.Custom.(*custom)
	if custom.c == nil {
		c := cron.New(cron.WithSeconds())
		id, err := c.AddFunc(expr, func() {
			custom.waiting <- struct{}{}
		})
		if err != nil {
			return nil, err
		}
		custom.c = c
		custom.id = id
	}
	custom.c.Start()

	select {
	case <-custom.waiting:
		custom.c.Stop()
		return map[string]string{"which": _manifest.Name}, nil
	case <-ctx.Done():
		custom.Close()
		return nil, ctx.Err()
	}
}

type custom struct {
	id      cron.EntryID
	c       *cron.Cron
	waiting chan struct{}
}

func (c *custom) Close() error {
	c.c.Stop()
	c.c.Remove(c.id)
	c.c = nil
	close(c.waiting)
	return nil
}
