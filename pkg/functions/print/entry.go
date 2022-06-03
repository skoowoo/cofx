package print

import "github.com/autoflowlabs/funcflow/pkg/functions"

func init() {
	functions.Register("print", &printer{})
}

// todo
type printer struct {
}

func (p *printer) Manifest() *functions.Manifest {
	return nil
}
