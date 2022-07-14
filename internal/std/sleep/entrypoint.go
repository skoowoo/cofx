package sleep

import (
	"context"

	"github.com/cofunclabs/cofunc/pkg/manifest"
)

var _manifest = manifest.Manifest{
	Name:           "sleep",
	Driver:         "go",
	EntrypointFunc: Entrypoint,
}

func New() *manifest.Manifest {
	return &_manifest
}

func Entrypoint(ctx context.Context, args map[string]string) (map[string]string, error) {
	return nil, nil
}
