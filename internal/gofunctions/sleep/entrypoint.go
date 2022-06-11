package sleep

import "github.com/cofunclabs/cofunc/pkg/manifest"

func New() manifest.Manifester {
	return &sleeper{}
}

// todo
type sleeper struct {
}

func (p *sleeper) Name() string {
	return "sleep"
}

func (p *sleeper) Manifest() manifest.Manifest {
	return manifest.Manifest{
		Driver:     "go",
		EntryPoint: "EntryPoint",
	}
}

func (p *sleeper) EntryPoint(args map[string]interface{}) {

}
