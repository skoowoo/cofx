package gitcurrentbranch

import (
	"context"
	"fmt"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/manifest"
	"github.com/cofxlabs/cofx/std/command"
)

var _manifest = manifest.Manifest{
	Category:       "git",
	Name:           "git_current_branch",
	Description:    "Use 'git branch --show-current' to get the current branch",
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
		"cmd":            "git branch --show-current",
		"split":          "",
		"extract_fields": "0",
		"query_columns":  "c0",
		"query_where":    "",
	}
	_, ep, _ := command.New()
	rets, err := ep(ctx, bundle, _args)
	if err != nil {
		return nil, fmt.Errorf("%w: in git_current_branch function", err)
	}
	v, ok := rets["outcome_0"]
	if !ok || v == "" {
		return nil, fmt.Errorf("not found current branch")
	}
	rets["current_branch"] = v
	rets["outcome"] = v

	return rets, nil
}
