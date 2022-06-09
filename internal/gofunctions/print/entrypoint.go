package print

import "github.com/cofunclabs/cofunc/pkg/functiondefine"

func Function() functiondefine.Define {
	return &printer{}
}

// todo
type printer struct {
}

func (p *printer) Manifest() *functiondefine.Manifest {
	return nil
}

func (p *printer) Name() string {
	return "print"
}
