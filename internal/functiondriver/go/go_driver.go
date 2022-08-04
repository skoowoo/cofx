package godriver

import (
	"context"
	"errors"

	"github.com/cofunclabs/cofunc/internal/std"
	"github.com/cofunclabs/cofunc/pkg/manifest"
)

// GoDriver
type GoDriver struct {
	path     string
	fname    string
	manifest *manifest.Manifest
}

func New(fname, fpath string) *GoDriver {
	return &GoDriver{
		path:  fpath,
		fname: fname,
	}
}

// load go://function
func (d *GoDriver) Load(ctx context.Context) error {
	mf := std.Lookup(d.fname)
	if mf == nil {
		return errors.New("in builtins package, not found function's manifest: " + d.path)
	}
	d.manifest = mf
	return nil
}

func (d *GoDriver) MergeArgs(args map[string]string) map[string]string {
	return mergeArgs(d.manifest.Args, args)
}

func (d *GoDriver) Run(ctx context.Context, args map[string]string) (map[string]string, error) {
	entrypoint := d.manifest.EntrypointFunc
	if entrypoint == nil {
		return nil, errors.New("in function, not found the entrypoint: " + d.path)
	}
	out, err := entrypoint(ctx, args)
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
