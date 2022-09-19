package gitgetremote

import (
	"context"
	"fmt"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/manifest"
	"github.com/cofxlabs/cofx/std/command"
)

var targetArg = manifest.UsageDesc{
	Name: "target",
	Desc: "Specify a target to get remote url, .e.g origin, upstream",
}

var _manifest = manifest.Manifest{
	Category:       "git",
	Name:           "git_get_remote",
	Description:    "Use 'git remote -v' to get remote url",
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
		return nil, nil
	}
	// upstream	https://github.com/cofxlabs/cofx.git (fetch)
	_args := spec.EntrypointArgs{
		"cmd":            "git remote -v",
		"split":          "",
		"extract_fields": "0,1,2",
		"query_columns":  "c1",
		"query_where":    fmt.Sprintf("c0 == '%s'", target) + " AND c2 like '%fetch%'",
	}
	_, ep, _ := command.New()
	rets, err := ep(ctx, bundle, _args)
	if err != nil {
		return nil, fmt.Errorf("%w: in git_get_remote function", err)
	}
	v, ok := rets["outcome_0"]
	if !ok || v == "" {
		return nil, fmt.Errorf("not found remote url for target %s", target)
	}
	rets["remote"] = v
	rets["outcome"] = v
	return rets, nil
}
