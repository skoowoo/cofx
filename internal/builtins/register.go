package builtins

import (
	"errors"

	"github.com/cofunclabs/cofunc/pkg/manifest"
)

var builtin map[string]manifest.Manifester

func Lookup(name string) manifest.Manifester {
	fc, ok := builtin[name]
	if ok {
		return fc
	}
	return nil
}

func register(name string, def manifest.Manifester) error {
	_, ok := builtin[name]
	if ok {
		return errors.New("repeat register the function name: " + name)
	}
	builtin[name] = def
	return nil
}
