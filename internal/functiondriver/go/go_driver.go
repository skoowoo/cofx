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
	location string
	funcName string
	fn       manifest.Manifester
	manifest *manifest.Manifest
}

func New(loc string) *GoDriver {
	if !strings.HasPrefix(loc, "go:") {
		return nil
	}
	name := strings.TrimPrefix(loc, "go:")
	return &GoDriver{
		location: name,
		funcName: name,
	}
}

// load go://function
func (d *GoDriver) Load(ctx context.Context, args map[string]string) error {
	fn := builtins.Lookup(d.location)
	if fn == nil {
		return errors.New("in builtins package, not found function: " + d.location)
	}
	mf := fn.Manifest()
	if mf.Args == nil {
		mf.Args = make(map[string]string)
	}
	for k, v := range args {
		mf.Args[k] = v
	}
	d.fn = fn
	d.manifest = &mf
	return nil
}

func (d *GoDriver) Run(ctx context.Context) (map[string]string, error) {
	entrypoint := d.manifest.EntryPointFunc
	if entrypoint == nil {
		return nil, errors.New("in function, not found the entrypoint: " + d.location)
	}
	out, err := entrypoint(ctx, d.manifest.Args)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (d *GoDriver) Name() string {
	return d.funcName
}
