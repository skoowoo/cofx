package exit

import (
	"context"
	"errors"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/manifest"
)

var errorArg = manifest.UsageDesc{
	Name: "error",
	Desc: "Specify a error message",
}

var _manifest = manifest.Manifest{
	Name:           "exit",
	Description:    "Run exit function to make flow stopped",
	Driver:         "go",
	Args:           map[string]string{},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{errorArg},
		ReturnValues: []manifest.UsageDesc{},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	s := args.GetString(errorArg.Name)
	if s != "" {
		return nil, errors.New(s)
	}
	return nil, nil
}
