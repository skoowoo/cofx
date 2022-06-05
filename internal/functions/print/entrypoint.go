package print

import "github.com/autoflowlabs/funcflow/internal/functions"

func init() {
	functions.Register("print", &printer{})
}

// todo
type printer struct {
}

func (p *printer) Manifest() *functions.Manifest {
	return nil
}
