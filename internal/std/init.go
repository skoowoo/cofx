package std

import (
	"errors"

	"github.com/cofunclabs/cofunc/internal/std/command"
	"github.com/cofunclabs/cofunc/internal/std/print"
	"github.com/cofunclabs/cofunc/internal/std/sleep"
	cotime "github.com/cofunclabs/cofunc/internal/std/time"
	"github.com/cofunclabs/cofunc/pkg/manifest"
)

func Lookup(name string) manifest.Manifester {
	fc, ok := builtin[name]
	if ok {
		return fc
	}
	return nil
}

var builtin map[string]manifest.Manifester

func init() {
	builtin = make(map[string]manifest.Manifester)

	var stds = []func() manifest.Manifester{
		sleep.New,
		print.New,
		command.New,
		cotime.New,
	}

	for _, New := range stds {
		def := New()
		if err := register(def.Name(), def); err != nil {
			panic(err)
		}
	}
}

func register(name string, def manifest.Manifester) error {
	_, ok := builtin[name]
	if ok {
		return errors.New("repeat register the function name: " + name)
	}
	builtin[name] = def
	return nil
}
