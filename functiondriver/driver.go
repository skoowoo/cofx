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
		if d := godriver.New(l.FuncName, l.FuncPath); d != nil {
			return d
		}

	case "cmd":
		if d := cmddriver.New(l.FuncName, l.FuncPath); d != nil {
			return d
		}
	}
	return nil
}

type Location struct {
	DriverName string
	FuncName   string
	FuncPath   string
}

func (l Location) String() string {
	return l.DriverName + ":" + l.FuncPath
}

type LocationStore map[string]Location

func NewLocationStore() LocationStore {
	return LocationStore(make(map[string]Location))
}

func (l LocationStore) Add(s string) (Location, error) {
	fields := strings.Split(s, ":")
	loc := Location{
		DriverName: fields[0],
		FuncName:   path.Base(fields[1]),
		FuncPath:   fields[1],
	}
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
