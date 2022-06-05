package funcflow

import (
	"github.com/autoflowlabs/funcflow/internal/gofunctions"
	"github.com/autoflowlabs/funcflow/internal/gofunctions/print"
	"github.com/autoflowlabs/funcflow/internal/gofunctions/sleep"
)

func init() {
	{
		def := sleep.Function()
		if err := gofunctions.Register(def.Name(), def); err != nil {
			panic(err)
		}
	}

	{
		def := print.Function()
		if err := gofunctions.Register(def.Name(), def); err != nil {
			panic(err)
		}
	}
}
