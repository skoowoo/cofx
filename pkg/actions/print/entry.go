package print

import "github.com/autoflowlabs/autoflow/pkg/actions"

func init() {
	actions.Register("print", &printer{})
}

// todo
type printer struct {
}

func (p *printer) Manifest() *actions.Manifest {
	return nil
}
