package print

import "github.com/autoflowlabs/funcflow/internal/gofunctions"

var Fake struct{}

func init() {
	gofunctions.Register("print", &printer{})
}

// todo
type printer struct {
}

func (p *printer) Manifest() *gofunctions.Manifest {
	return nil
}
