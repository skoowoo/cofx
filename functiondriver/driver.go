package functiondriver

import (
	"context"
	"errors"
	"path"
	"strings"

	cmddriver "github.com/cofunclabs/cofunc/functiondriver/cmd"
	godriver "github.com/cofunclabs/cofunc/functiondriver/go"
)

type Driver interface {
	FunctionName() string
	Load(context.Context) error
	MergeArgs(map[string]string) map[string]string
	Run(context.Context, map[string]string) (map[string]string, error)
}

func New(l Location) Driver {
	switch l.DriverName {
	case "go":
		if d := godriver.New(l.FuncName, l.FuncPath, l.Version); d != nil {
			return d
		}

	case "cmd":
		if d := cmddriver.New(l.FuncName, l.FuncPath, l.Version); d != nil {
			return d
		}
	}
	return nil
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
