package shelldriver

import (
	"context"
	"io"

	"github.com/cofunclabs/cofunc/manifest"
)

const Name = "shell"

type ShellDriver struct {
	fpath   string
	fname   string
	version string
	logger  io.Writer
}

func New(fname, fpath, version string) *ShellDriver {
	return &ShellDriver{
		fname:   fname,
		fpath:   fpath,
		version: version,
	}
}

func (d *ShellDriver) Load(ctx context.Context, logger io.Writer) error {
	// todo
	return nil
}

func (d *ShellDriver) MergeArgs(args map[string]string) map[string]string {
	return nil
}

func (d *ShellDriver) Run(ctx context.Context, args map[string]string) (map[string]string, error) {
	return nil, nil
}

func (d *ShellDriver) StopAndRelease(ctx context.Context) error {
	return nil
}

func (d *ShellDriver) FunctionName() string {
	return d.fname
}

func (d *ShellDriver) Name() string {
	return Name
}
func (d *ShellDriver) Manifest() manifest.Manifest {
	return manifest.Manifest{}
}
