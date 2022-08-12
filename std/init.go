package std

import (
	"errors"

	"github.com/cofunclabs/cofunc/manifest"
	"github.com/cofunclabs/cofunc/std/command"
	gobuild "github.com/cofunclabs/cofunc/std/go/go_build"
	gogenerate "github.com/cofunclabs/cofunc/std/go/go_generate"
	"github.com/cofunclabs/cofunc/std/print"
	"github.com/cofunclabs/cofunc/std/sleep"
	cotime "github.com/cofunclabs/cofunc/std/time"
)

func Lookup(name string) *manifest.Manifest {
	fc, ok := builtin[name]
	if ok {
		return fc
	}
	return nil
}

var builtin map[string]*manifest.Manifest

func init() {
	builtin = make(map[string]*manifest.Manifest)

	var stds = []func() *manifest.Manifest{
		sleep.New,
		print.New,
		command.New,
		cotime.New,
		gobuild.New,
		gogenerate.New,
	}

	for _, New := range stds {
		mf := New()
		if err := register(mf.Name, mf); err != nil {
			panic(err)
		}
	}
}

func register(name string, mf *manifest.Manifest) error {
	_, ok := builtin[name]
	if ok {
		return errors.New("repeat register the function name: " + name)
	}
	builtin[name] = mf
	return nil
}