package gofunctions

import (
	"github.com/cofunclabs/cofunc/internal/gofunctions/print"
	"github.com/cofunclabs/cofunc/internal/gofunctions/sleep"
	"github.com/cofunclabs/cofunc/pkg/functiondefine"
)

func init() {
	builtin = make(map[string]functiondefine.Define)

	{
		def := sleep.Function()
		if err := register(def.Name(), def); err != nil {
			panic(err)
		}
	}

	{
		def := print.Function()
		if err := register(def.Name(), def); err != nil {
			panic(err)
		}
	}
}
