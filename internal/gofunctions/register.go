package gofunctions

import (
	"errors"

	"github.com/cofunclabs/cofunc/pkg/functiondefine"
)

var builtin map[string]functiondefine.Define

func Lookup(name string) functiondefine.Define {
	fc, ok := builtin[name]
	if ok {
		return fc
	}
	return nil
}

func register(name string, def functiondefine.Define) error {
	_, ok := builtin[name]
	if ok {
		return errors.New("repeat register the function name: " + name)
	}
	builtin[name] = def
	return nil
}
