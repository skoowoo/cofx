package godriver

import (
	"context"
	"errors"
	"strings"

	"github.com/cofunclabs/cofunc/internal/builtins"
	"github.com/cofunclabs/cofunc/pkg/manifest"
)

// GoDriver
type GoDriver struct {
	path       string
	fname      string
	manifest   *manifest.Manifest
	mergedArgs map[string]string
}

func New(loc string) *GoDriver {
	if !strings.HasPrefix(loc, "go:") {
		return nil
	}
	name := strings.TrimPrefix(loc, "go:")
	return &GoDriver{
		path:  name,
		fname: name,
	}
}

// load go://function
func (d *GoDriver) Load(ctx context.Context) error {
	fn := builtins.Lookup(d.fname)
	if fn == nil {
		return errors.New("in builtins package, not found function: " + d.path)
	}
	mf := fn.Manifest()
	d.manifest = &mf
	return nil
}

func (d *GoDriver) MergeArgs(args map[string]string) error {
	d.mergedArgs = mergeArgs(d.manifest.Args, args)
	return nil
}

func (d *GoDriver) Run(ctx context.Context) (map[string]string, error) {
	entrypoint := d.manifest.EntrypointFunc
	if entrypoint == nil {
		return nil, errors.New("in function, not found the entrypoint: " + d.path)
	}
	out, err := entrypoint(ctx, d.mergedArgs)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (d *GoDriver) FunctionName() string {
	return d.fname
}

func mergeArgs(base, prior map[string]string) map[string]string {
	merged := make(map[string]string)
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range prior {
		merged[k] = v
	}
	return merged
}
