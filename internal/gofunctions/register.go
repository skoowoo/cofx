package gofunctions

import "errors"

var builtin map[string]FunctionDefine

func init() {
	builtin = make(map[string]FunctionDefine)
}

type FunctionDefine interface {
	Manifest() *Manifest
}

func Lookup(name string) FunctionDefine {
	fc, ok := builtin[name]
	if ok {
		return fc
	}
	return nil
}

func Register(name string, def FunctionDefine) error {
	_, ok := builtin[name]
	if ok {
		return errors.New("repeat register the function name: " + name)
	}
	builtin[name] = def
	return nil
}
