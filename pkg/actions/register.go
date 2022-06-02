package actions

import "errors"

var builtin map[string]ActionDefine

func init() {
	builtin = make(map[string]ActionDefine)
}

type ActionDefine interface {
	Manifest() *Manifest
}

func Lookup(name string) ActionDefine {
	action, ok := builtin[name]
	if ok {
		return action
	}
	return nil
}

func Register(name string, def ActionDefine) error {
	_, ok := builtin[name]
	if ok {
		return errors.New("repeat register the action name: " + name)
	}
	builtin[name] = def
	return nil
}
