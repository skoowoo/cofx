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

func (d *CmdDriver) Load(args map[string]string) error {
	// todo
	return nil
}

func (d *CmdDriver) Run() (map[string]string, error) {
	return nil, nil
}

func (d *CmdDriver) Name() string {
	return d.funcName
}
