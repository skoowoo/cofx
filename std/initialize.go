package std

import (
	"errors"
	"fmt"

	"github.com/skoowoo/cofx/functiondriver/go/spec"
	"github.com/skoowoo/cofx/manifest"
	"github.com/skoowoo/cofx/std/command"
	eventcron "github.com/skoowoo/cofx/std/events/event_cron"
	eventtick "github.com/skoowoo/cofx/std/events/event_tick"
	gitaddupstream "github.com/skoowoo/cofx/std/git/git_add_upstream"
	gitbasic "github.com/skoowoo/cofx/std/git/git_basic"
	gitcheckmerge "github.com/skoowoo/cofx/std/git/git_check_merge"
	gitfetch "github.com/skoowoo/cofx/std/git/git_fetch"
	gitinsight "github.com/skoowoo/cofx/std/git/git_insight"
	gitpull "github.com/skoowoo/cofx/std/git/git_pull"
	gitpush "github.com/skoowoo/cofx/std/git/git_push"
	gitrebase "github.com/skoowoo/cofx/std/git/git_rebase"
	ghcreatepr "github.com/skoowoo/cofx/std/github/gh_create_pr"
	gobuild "github.com/skoowoo/cofx/std/go/go_build"
	gogenerate "github.com/skoowoo/cofx/std/go/go_generate"
	"github.com/skoowoo/cofx/std/go/gotest"
	httpget "github.com/skoowoo/cofx/std/http/http_get"
	httppost "github.com/skoowoo/cofx/std/http/http_post"
	"github.com/skoowoo/cofx/std/outcome"
	"github.com/skoowoo/cofx/std/print"
	stdtime "github.com/skoowoo/cofx/std/time"
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
		print.New,
		command.New,
		stdtime.New,
		outcome.New,
		// go
		gobuild.New,
		gogenerate.New,
		gotest.New,
		// http
		httpget.New,
		httppost.New,
		// git
		gitfetch.New,
		gitpush.New,
		gitcheckmerge.New,
		gitrebase.New,
		gitpull.New,
		gitaddupstream.New,
		gitbasic.New,
		gitinsight.New,
		// github
		ghcreatepr.New,
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
