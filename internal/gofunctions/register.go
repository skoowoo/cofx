package gofunctions

import (
	"errors"

	"github.com/autoflowlabs/funcflow/pkg/functiondefine"
)

var builtin map[string]functiondefine.Define

func init() {
	builtin = make(map[string]functiondefine.Define)
}

func Lookup(name string) functiondefine.Define {
	fc, ok := builtin[name]
	if ok {
		return fc
	}
	return nil
}

func Register(name string, def functiondefine.Define) error {
	_, ok := builtin[name]
	if ok {
		return errors.New("repeat register the function name: " + name)
	}
	builtin[name] = def
	return nil
}
