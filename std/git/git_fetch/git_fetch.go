package gitfetch

import (
	"context"
	"fmt"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/manifest"
	"github.com/cofxlabs/cofx/std/command"
)

var targetArg = manifest.UsageDesc{
	Name: "target",
	Desc: "Specify a target to get remote url, e.g. origin, upstream; if not specified, fetch all remotes",
}

var _manifest = manifest.Manifest{
	Category:       "git",
	Name:           "git_fetch",
	Description:    "Use the 'git fetch' command to update the local repository",
	Driver:         "go",
	Args:           map[string]string{},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{targetArg},
		ReturnValues: []manifest.UsageDesc{},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	target := args.GetString(targetArg.Name)
	if target == "" {
		target = "--all"
	}
	_args := spec.EntrypointArgs{
		"cmd":            fmt.Sprintf("git fetch %s", target),
		"split":          "",
		"extract_fields": "",
		"query_columns":  "",
		"query_where":    "",
	}
	_, ep, _ := command.New()
	rets, err := ep(ctx, bundle, _args)
	if err != nil {
		return nil, fmt.Errorf("%w: in git_fetch function", err)
	}
	return rets, nil
}
