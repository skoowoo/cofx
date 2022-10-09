package gitrebase

import (
	"context"
	"fmt"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/manifest"
	"github.com/cofxlabs/cofx/std/command"
)

var branchArg = manifest.UsageDesc{
	Name: "branch",
	Desc: "Specify the branch name",
}

var _manifest = manifest.Manifest{
	Category:       "git",
	Name:           "git_rebase",
	Description:    "Merge branches using 'git rebase' command",
	Driver:         "go",
	Args:           map[string]string{},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{branchArg},
		ReturnValues: []manifest.UsageDesc{},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	branch := args.GetString(branchArg.Name)
	if branch == "" {
		return nil, fmt.Errorf("not specified local_branch")
	}

	_args := spec.EntrypointArgs{
		"cmd":            "git rebase " + branch,
		"split":          "",
		"extract_fields": "",
		"query_columns":  "",
		"query_where":    "",
	}
	_, ep, _ := command.New()
	rets, err := ep(ctx, bundle, _args)
	if err != nil {
		return nil, fmt.Errorf("%w: in git_rebase function", err)
	}
	return rets, nil
}
