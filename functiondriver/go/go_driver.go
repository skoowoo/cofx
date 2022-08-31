package godriver

import (
	"context"
	"errors"
	"io"

	"github.com/cofunclabs/cofunc/functiondriver/go/spec"
	"github.com/cofunclabs/cofunc/manifest"
	"github.com/cofunclabs/cofunc/std"
)

const Name = "go"

// GoDriver
type GoDriver struct {
	path       string
	fname      string
	version    string
	manifest   *manifest.Manifest
	entrypoint spec.EntrypointFunc
	logger     io.Writer
	bound      spec.Custom
}

func New(fname, fpath, version string) *GoDriver {
	return &GoDriver{
		path:    fpath,
		fname:   fname,
		version: version,
	}
}

// load go://function
func (d *GoDriver) Load(ctx context.Context, logger io.Writer) error {
	mf, ep, create := std.Lookup(d.fname)
	if mf == nil || ep == nil {
		return errors.New("in std, not found function's manifest or entrypoint: " + d.path)
	}
	d.manifest = mf
	d.entrypoint = ep
	d.logger = logger
	if create != nil {
		d.bound = create()
	}
	return nil
}

func (d *GoDriver) Run(ctx context.Context, args map[string]string) (map[string]string, error) {
	merged := d.mergeArgs(args)
	bundle := spec.EntrypointBundle{
		Version: d.version,
		Logger:  d.logger,
		Bound:   d.bound,
	}
	out, err := d.entrypoint(ctx, bundle, spec.EntrypointArgs(merged))
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (d *GoDriver) FunctionName() string {
	return d.fname
}

func (d *GoDriver) Name() string {
	return Name
}

func (d *GoDriver) Manifest() manifest.Manifest {
	return *d.manifest
}

func (d *GoDriver) mergeArgs(args map[string]string) map[string]string {
	merged := make(map[string]string)
	for k, v := range d.manifest.Args {
		merged[k] = v
	}
	for k, v := range args {
		merged[k] = v
	}
	return merged
}
