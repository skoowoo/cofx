package godriver

import (
	"context"
	"errors"
	"io"

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
	entrypoint manifest.EntrypointFunc
	logger     io.Writer
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
	mf, ep := std.Lookup(d.fname)
	if mf == nil || ep == nil {
		return errors.New("in std, not found function's manifest or entrypoint: " + d.path)
	}
	d.manifest = mf
	d.entrypoint = ep
	d.logger = logger
	return nil
}

func (d *GoDriver) Run(ctx context.Context, args map[string]string) (map[string]string, error) {
	out, err := d.entrypoint(ctx, d.logger, d.version, args)
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

func (d *GoDriver) MergeArgs(args map[string]string) map[string]string {
	return mergeArgs(d.manifest.Args, args)
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
