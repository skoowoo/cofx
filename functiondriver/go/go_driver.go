package godriver

import (
	"context"
	"errors"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/skoowoo/cofx/functiondriver/go/spec"
	"github.com/skoowoo/cofx/manifest"
	"github.com/skoowoo/cofx/service/resource"
	"github.com/skoowoo/cofx/std"
)

const Name = "go"

// GoDriver
type GoDriver struct {
	// path is got from 'load' statement in flowl
	path string
	// fname is the function name, got from 'load' statement in flowl.
	fname string
	// version is the function version that be expected, got from 'load' statement in flowl.
	version string
	// manifest be defined by function
	manifest *manifest.Manifest
	// entrypoint is a function that's the entry of the function, it's defined by function.
	entrypoint spec.EntrypointFunc
	// custom is a custom object for function, used to keep some states, it's created by function.
	custom spec.Customer
	// resources contains some services that can be used by function.
	resources resource.Resources
}

// New create a new GoDriver instance, the arguments are got from 'load' statement in flowl.
func New(fname, fpath, version string) *GoDriver {
	return &GoDriver{
		path:    fpath,
		fname:   fname,
		version: version,
	}
}

// Load loads the expected function into the driver.
func (d *GoDriver) Load(ctx context.Context, resources resource.Resources) error {
	mf, ep, create := std.Lookup(d.fname)
	if mf == nil || ep == nil {
		return errors.New("in std, not found function's manifest or entrypoint: " + d.path)
	}
	d.manifest = mf
	d.entrypoint = ep
	if create != nil {
		d.custom = create()
	}
	d.resources = resources
	return nil
}

// Run make the driver to run, then execute the function through the entrypoint.
func (d *GoDriver) Run(ctx context.Context, args map[string]string) (map[string]string, error) {
	merged := d.mergeArgs(args)
	bundle := spec.EntrypointBundle{
		Version:   d.version,
		Custom:    d.custom,
		Resources: d.resources,
	}
	pretty, ok := d.resources.Logwriter.(resource.OutPrettyPrinter)
	if ok {
		defer func() {
			pretty.Reset()
		}()
		pretty.WriteTitle(d.resources.Labels.Get("node_name"), d.Name()+":"+d.FunctionName())
	}
	out, err := d.entrypoint(ctx, bundle, spec.EntrypointArgs(merged))
	if err != nil {
		return nil, err
	}
	if ok {
		var keys []string
		var maxLen int
		for k := range out {
			if len(k) > maxLen {
				maxLen = len(k)
			}
			keys = append(keys, k)
		}
		var lines []string
		sort.Strings(keys)
		for _, k := range keys {
			l := lipgloss.NewStyle().Width(maxLen).Render(k) + ": " + out[k]
			lines = append(lines, l)
		}
		pretty.WriteSummary(lines)
	}
	return out, nil
}

// StopAndRelease closes the custom object of the function.
func (d *GoDriver) StopAndRelease(ctx context.Context) error {
	if d.custom != nil {
		return d.custom.Close()
	}
	return nil
}

// FunctionName returns the function name.
func (d *GoDriver) FunctionName() string {
	return d.fname
}

// Name returns the driver name.
func (d *GoDriver) Name() string {
	return Name
}

// Manifest returns the function manifest.
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
