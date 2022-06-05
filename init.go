package funcflow

import (
	"github.com/autoflowlabs/funcflow/internal/gofunctions/print"
	"github.com/autoflowlabs/funcflow/internal/gofunctions/sleep"
)

func init() {
	_ = print.Fake
	_ = sleep.Fake
}
