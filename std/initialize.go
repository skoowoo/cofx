package std

import (
	"errors"
	"fmt"

	"github.com/cofunclabs/cofunc/functiondriver/go/spec"
	"github.com/cofunclabs/cofunc/manifest"
	"github.com/cofunclabs/cofunc/std/command"
	eventcron "github.com/cofunclabs/cofunc/std/events/event_cron"
	eventtick "github.com/cofunclabs/cofunc/std/events/event_tick"
	syncupstream "github.com/cofunclabs/cofunc/std/git/sync_upstream"
	gobuild "github.com/cofunclabs/cofunc/std/go/go_build"
	gogenerate "github.com/cofunclabs/cofunc/std/go/go_generate"
	"github.com/cofunclabs/cofunc/std/go/gotest"
	"github.com/cofunclabs/cofunc/std/outcome"
	"github.com/cofunclabs/cofunc/std/print"
	"github.com/cofunclabs/cofunc/std/sleep"
	stdtime "github.com/cofunclabs/cofunc/std/time"
)

// Lookup returns the manifest object and entrypoint method of the given function name.
func Lookup(name string) (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	fc, ok := builtin[name]
	if ok {
		return fc, entrypoints[fc.Entrypoint], factories[name]
	}
	return nil, nil, nil
}

// ListAll returns all the function manifest of the standard library.
func ListAll() []manifest.Manifest {
	var mfs []manifest.Manifest
	for _, m := range builtin {
		mfs = append(mfs, *m)
	}
	return mfs
}

func register(name string, mf *manifest.Manifest, ep spec.EntrypointFunc, cr spec.CreateCustomFunc) error {
	_, ok := builtin[name]
	if ok {
		return errors.New("repeat register the function name: " + name)
	}
	builtin[name] = mf
	entrypoints[mf.Entrypoint] = ep
	factories[name] = cr
	return nil
}

var (
	// builtin store kvs of function name -> manifest.
	builtin map[string]*manifest.Manifest
	// entrypoints store kvs of entrypoint name -> entrypoint func.
	entrypoints map[string]spec.EntrypointFunc
	// factories store kvs of function name -> a func that be used to create a custom object.
	factories map[string]spec.CreateCustomFunc
)

func init() {
	builtin = make(map[string]*manifest.Manifest)
	entrypoints = make(map[string]spec.EntrypointFunc)
	factories = make(map[string]spec.CreateCustomFunc)

	var stds = []func() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc){
		sleep.New,
		print.New,
		command.New,
		stdtime.New,
		gobuild.New,
		gogenerate.New,
		gotest.New,
		outcome.New,
		syncupstream.New,
		// event trigger function
		eventtick.New,
		eventcron.New,
	}

	for i, New := range stds {
		mf, ep, cr := New()
		// Get entrypoint name from entrypointfunc, then auto register the entrypoint field of the manifest.
		// NOTE: Automatically getted the entrypoint name is unique
		mf.Entrypoint = spec.Func2Name(ep)

		if mf.Name == "" {
			panic(fmt.Errorf("name is empty in manifest %d", i))
		}
		if mf.Entrypoint == "" {
			panic(fmt.Errorf("entrypoint is empty in manifest %d", i))
		}
		if ep == nil {
			panic(fmt.Errorf("entrypoint is nil in %d", i))
		}
		if err := register(mf.Name, mf, ep, cr); err != nil {
			panic(err)
		}
	}
}
