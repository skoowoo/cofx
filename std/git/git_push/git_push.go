package gitpush

import (
	"context"
	"fmt"

	"github.com/skoowoo/cofx/functiondriver/go/spec"
	"github.com/skoowoo/cofx/manifest"
	"github.com/skoowoo/cofx/std/command"
)

var localBranchArg = manifest.UsageDesc{
	Name: "local_branch",
	Desc: "Specify the local branch name",
}

var remoteBranchArg = manifest.UsageDesc{
	Name: "remote_branch",
	Desc: "Specify the remote branch name",
}

var _manifest = manifest.Manifest{
	Category:       "git",
	Name:           "git_push",
	Description:    "Sync local branch to remote using 'git push' command",
	Driver:         "go",
	Args:           map[string]string{},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{localBranchArg, remoteBranchArg},
		ReturnValues: []manifest.UsageDesc{},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	remote := args.GetString(remoteBranchArg.Name)
	local := args.GetString(localBranchArg.Name)
	if local == "" {
		return nil, fmt.Errorf("not specified local_branch")
	}

	var cmd string
	if remote != "" {
		cmd = fmt.Sprintf("git push origin %s:%s", local, remote)
	} else {
		cmd = fmt.Sprintf("git push origin %s", local)
	}

	_args := spec.EntrypointArgs{
		"cmd":            cmd,
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
