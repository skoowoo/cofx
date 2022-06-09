package sleep

import "github.com/cofunclabs/cofunc/pkg/functiondefine"

func Function() functiondefine.Define {
	return &sleeper{}
}

// todo
type sleeper struct {
}

func (p *sleeper) Name() string {
	return "sleep"
}

func (p *sleeper) Manifest() *functiondefine.Manifest {
	return nil
}
