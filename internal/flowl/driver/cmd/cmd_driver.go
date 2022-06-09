package cmddriver

import (
	"path"
	"strings"
)

// Cmd
type CmdDriver struct {
	location string
	funcName string
}

func New(loc string) *CmdDriver {
	if !strings.HasPrefix(loc, "cmd:") {
		return nil
	}
	p := strings.TrimPrefix(loc, "cmd:")
	name := path.Base(p)
	return &CmdDriver{
		funcName: name,
		location: p,
	}
}

func (d *CmdDriver) Load() error {
	// todo
	return nil
}

func (d *CmdDriver) Run() error {
	return nil
}

func (d *CmdDriver) Name() string {
	return d.funcName
}
