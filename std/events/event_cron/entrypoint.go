package eventcron

import (
	"context"
	"time"

	"github.com/cofunclabs/cofunc/functiondriver/go/spec"
	"github.com/cofunclabs/cofunc/manifest"
	"github.com/cofunclabs/cofunc/service/resource"
)

var exprArg = manifest.UsageDesc{
	Name:           "expr",
	OptionalValues: []string{},
	Desc:           "A cron expression, e.g. 0 0 * * *, 0 15 10 ? * *",
}

var _manifest = manifest.Manifest{
	Category:       "event",
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
	return &_manifest, Entrypoint, func() spec.Customer { return &custom{waiting: make(chan time.Time)} }
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	expr := args.GetString(exprArg.Name)
	custom := bundle.Custom.(*custom)
	if custom.entity == nil && custom.cron == nil {
		cron := bundle.Resources.CronTrigger
		entity, err := cron.Add(expr, custom.waiting)
		if err != nil {
			return nil, err
		}
		custom.entity = entity
		custom.cron = cron
	}

	select {
	case <-custom.waiting:
		return map[string]string{"which": _manifest.Name}, nil
	case <-ctx.Done():
		custom.Close()
		return nil, ctx.Err()
	}
}

type custom struct {
	entity  interface{}
	cron    resource.CronTrigger
	waiting chan time.Time
}

func (c *custom) Close() error {
	close(c.waiting)
	c.cron.Remove(c.entity)
	c.entity = nil
	c.cron = nil
	return nil
}
