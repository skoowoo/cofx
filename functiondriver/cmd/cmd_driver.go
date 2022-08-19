package cmddriver

import (
	"context"

	"github.com/cofunclabs/cofunc/pkg/output"
)

// Cmd
type CmdDriver struct {
	fpath   string
	fname   string
	version string
	output  *output.Output
}

func New(fname, fpath, version string) *CmdDriver {
	return &CmdDriver{
		fname:   fname,
		fpath:   fpath,
		version: version,
		output:  output.New(),
	}
}

func (d *CmdDriver) Load(ctx context.Context) error {
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
