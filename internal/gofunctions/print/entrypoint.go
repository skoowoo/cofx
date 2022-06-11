package print

import (
	"github.com/cofunclabs/cofunc/pkg/manifest"
)

func New() manifest.Manifester {
	return &printer{}
}

// todo
type printer struct {
}

func (p *printer) Manifest() manifest.Manifest {
	return manifest.Manifest{}
}

func (p *printer) Name() string {
	return "print"
}
