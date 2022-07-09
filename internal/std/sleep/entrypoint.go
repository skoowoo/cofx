package sleep

import (
	"context"

	"github.com/cofunclabs/cofunc/pkg/manifest"
)

func New() manifest.Manifester {
	return &sleeper{}
}

type sleeper struct{}

func (p *sleeper) Name() string {
	return "sleep"
}

func (p *sleeper) Manifest() manifest.Manifest {
	return manifest.Manifest{
		Driver:         "go",
		EntrypointFunc: p.Entrypoint,
	}
}

func (p *sleeper) Entrypoint(ctx context.Context, args map[string]string) (map[string]string, error) {
	return nil, nil
}
