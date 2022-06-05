package sleep

import "github.com/autoflowlabs/funcflow/internal/gofunctions"

var Fake struct{}

func init() {
	gofunctions.Register("sleep", &sleeper{})
}

// todo
type sleeper struct {
}

func (p *sleeper) Manifest() *gofunctions.Manifest {
	return nil
}
