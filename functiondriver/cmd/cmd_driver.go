package cmddriver

import (
	"context"
	"io"

	"github.com/cofunclabs/cofunc/manifest"
)

const Name = "cmd"

// Cmd
type CmdDriver struct {
	fpath   string
	fname   string
	version string
	logger  io.Writer
}

func New(fname, fpath, version string) *CmdDriver {
	return &CmdDriver{
		fname:   fname,
		fpath:   fpath,
		version: version,
	}
}

func (d *CmdDriver) Load(ctx context.Context, logger io.Writer) error {
	// todo
	return nil
}

func (d *CmdDriver) MergeArgs(args map[string]string) map[string]string {
	return nil
}

func (d *CmdDriver) Run(ctx context.Context, args map[string]string) (map[string]string, error) {
	return nil, nil
}

func (d *CmdDriver) FunctionName() string {
	return d.fname
}

func (d *CmdDriver) Name() string {
	return Name
}
func (d *CmdDriver) Manifest() manifest.Manifest {
	return manifest.Manifest{}
}
