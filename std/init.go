package std

import (
	"errors"
	"fmt"

	"github.com/cofunclabs/cofunc/manifest"
	"github.com/cofunclabs/cofunc/std/command"
	eventtick "github.com/cofunclabs/cofunc/std/events/event_tick"
	gobuild "github.com/cofunclabs/cofunc/std/go/go_build"
	gogenerate "github.com/cofunclabs/cofunc/std/go/go_generate"
	"github.com/cofunclabs/cofunc/std/print"
	"github.com/cofunclabs/cofunc/std/sleep"
	stdtime "github.com/cofunclabs/cofunc/std/time"
)

// Lookup returns the manifest object and entrypoint method of the given function name.
func Lookup(name string) (*manifest.Manifest, manifest.EntrypointFunc) {
	fc, ok := builtin[name]
	if ok {
		return fc, lookupEntrypoint(fc)
	}
	return nil, nil
}

func lookupEntrypoint(mf *manifest.Manifest) manifest.EntrypointFunc {
	return entrypoints[mf.Name+mf.Entrypoint]
}

func register(name string, mf *manifest.Manifest, ep manifest.EntrypointFunc) error {
	_, ok := builtin[name]
	if ok {
		return errors.New("repeat register the function name: " + name)
	}
	builtin[name] = mf
	entrypoints[mf.Name+mf.Entrypoint] = ep
	return nil
}

var (
	// builtin store kvs of function name -> manifest.
	builtin map[string]*manifest.Manifest
	// entrypoints store kvs of function name + entrypoint name -> entrypoint func.
	entrypoints map[string]manifest.EntrypointFunc
)

func init() {
	builtin = make(map[string]*manifest.Manifest)
	entrypoints = make(map[string]manifest.EntrypointFunc)

	var stds = []func() (*manifest.Manifest, manifest.EntrypointFunc){
		sleep.New,
		print.New,
		command.New,
		stdtime.New,
		gobuild.New,
		gogenerate.New,

		// event trigger function
		eventtick.New,
	}

	for i, New := range stds {
		mf, ep := New()
		// Get entrypoint name from entrypointfunc, then auto register the entrypoint field of the manifest.
		mf.Entrypoint = manifest.Func2Name(ep)

		if mf.Name == "" {
			panic(fmt.Errorf("name is empty in manifest %d", i))
		}
		if mf.Entrypoint == "" {
			panic(fmt.Errorf("entrypoint is empty in manifest %d", i))
		}
		if ep == nil {
			panic(fmt.Errorf("entrypoint is nil in %d", i))
		}
		if err := register(mf.Name, mf, ep); err != nil {
			panic(err)
		}
	}
}
