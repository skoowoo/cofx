package gofunctions

import (
	"github.com/autoflowlabs/funcflow/internal/gofunctions/print"
	"github.com/autoflowlabs/funcflow/internal/gofunctions/sleep"
	"github.com/autoflowlabs/funcflow/pkg/functiondefine"
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
