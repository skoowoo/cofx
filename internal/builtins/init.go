package builtins

import (
	"github.com/cofunclabs/cofunc/internal/builtins/command"
	"github.com/cofunclabs/cofunc/internal/builtins/print"
	"github.com/cofunclabs/cofunc/internal/builtins/sleep"
	"github.com/cofunclabs/cofunc/pkg/manifest"
)

func init() {
	builtin = make(map[string]manifest.Manifester)

	{
		def := sleep.New()
		if err := register(def.Name(), def); err != nil {
			panic(err)
		}
	}

	{
		def := print.New()
		if err := register(def.Name(), def); err != nil {
			panic(err)
		}
	}

	{
		def := command.New()
		if err := register(def.Name(), def); err != nil {
			panic(err)
		}
	}
}
