package functiondriver

import (
	"context"
	"errors"
	"path"
	"strings"

	godriver "github.com/cofxlabs/cofx/functiondriver/go"
	shelldriver "github.com/cofxlabs/cofx/functiondriver/shell"
	"github.com/cofxlabs/cofx/manifest"
	"github.com/cofxlabs/cofx/service/resource"
)

type Driver interface {
	// Name returns the name of the driver, e.g. "go", "cmd", etc.
	Name() string
	// FunctionName returns the name of the function associated with the driver
	FunctionName() string
	// Manifest returns the manifest of the function associated with the driver
	Manifest() manifest.Manifest
	// Load loads the function into the driver
	Load(context.Context, resource.Resources) error
	// Run calls the entrypoint of the function associated with the driver to execute the function code
	Run(context.Context, map[string]string) (map[string]string, error)
	// StopAndRelease stops the driver and releases the resources of the driver
	StopAndRelease(context.Context) error
}

// New creates a driver instance based on the 'load' information in flowl source file,
// A Driver instance contains two parts that's driver and function
func New(l Location) Driver {
	var dr Driver
	switch l.DriverName {
	case godriver.Name:
		if d := godriver.New(l.FuncName, l.FuncPath, l.Version); d == nil {
			return nil
		} else {
			dr = d
		}
	case shelldriver.Name:
		if d := shelldriver.New(l.FuncName, l.FuncPath, l.Version); d == nil {
			return nil
		} else {
			dr = d
		}
	}
	return dr
}

type Location struct {
	DriverName string
	FuncName   string
	FuncPath   string
	Version    string
}

func NewLocation(s string) Location {
	fields := strings.Split(s, ":")
	dname, fpath := fields[0], fields[1]

	fname := path.Base(fields[1])
	version := "latest"
	if names := strings.Split(fname, "@"); len(names) == 2 {
		fname = names[0]
		version = names[1]
	}

	loc := Location{
		DriverName: dname,
		FuncName:   fname,
		FuncPath:   fpath,
		Version:    version,
	}
	return loc
}

func (l Location) String() string {
	return l.DriverName + ":" + l.FuncPath
}

type LocationStore map[string]Location

func NewLocationStore() LocationStore {
	return LocationStore(make(map[string]Location))
}

func (l LocationStore) Add(s string) (Location, error) {
	loc := NewLocation(s)
	if _, ok := l[loc.FuncName]; ok {
		return loc, errors.New("name conflict")
	}
	l[loc.FuncName] = loc
	return loc, nil
}

func (l LocationStore) Get(fname string) (Location, bool) {
	loc, ok := l[fname]
	return loc, ok
}
