package godriver

import (
	"errors"
	"strings"

	"github.com/cofunclabs/cofunc/internal/gofunctions"
)

// GoDriver
type GoDriver struct {
	location string
	funcName string
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
func (d *GoDriver) Load() error {
	def := gofunctions.Lookup(d.location)
	if def == nil {
		return errors.New("in gofunctions package, not found function: " + d.location)
	}
	manifest := def.Manifest()
	// todo
	_ = manifest
	return nil
}

func (d *GoDriver) Run() error {
	return nil
}

func (d *GoDriver) Name() string {
	return d.funcName
}
