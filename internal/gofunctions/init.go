package gofunctions

import (
	"github.com/cofunclabs/cofunc/internal/gofunctions/print"
	"github.com/cofunclabs/cofunc/internal/gofunctions/sleep"
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
}
