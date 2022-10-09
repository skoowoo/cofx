package gitpull

import (
	"context"
	"fmt"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/manifest"
	"github.com/cofxlabs/cofx/std/command"
)

var _manifest = manifest.Manifest{
	Category:       "git",
	Name:           "git_pull",
	Description:    "Update the local repository with the 'git pull' command",
	Driver:         "go",
	Args:           map[string]string{},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{},
		ReturnValues: []manifest.UsageDesc{},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	_args := spec.EntrypointArgs{
		"cmd":            "git pull",
		"split":          "",
		"extract_fields": "",
		"query_columns":  "",
		"query_where":    "",
	}
	_, ep, _ := command.New()
	rets, err := ep(ctx, bundle, _args)
	if err != nil {
		return nil, fmt.Errorf("%w: in git_pull function", err)
	}
	return rets, nil
}
