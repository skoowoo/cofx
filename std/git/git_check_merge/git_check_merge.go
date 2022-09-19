package gitcheckmerge

import (
	"context"
	"fmt"
	"strings"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/manifest"
	"github.com/cofxlabs/cofx/std/command"
)

var toBranchArg = manifest.UsageDesc{
	Name: "to_branch",
	Desc: "Specify the to branch name",
}

var fromBranchArg = manifest.UsageDesc{
	Name: "from_branch",
	Desc: "Specify the from branch name",
}

var _manifest = manifest.Manifest{
	Category:       "git",
	Name:           "git_check_merge",
	Description:    "Use 'git merge-base/merge-tree' to check two branches are conflict or not",
	Driver:         "go",
	Args:           map[string]string{},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{toBranchArg, fromBranchArg},
		ReturnValues: []manifest.UsageDesc{},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	from := args.GetString(fromBranchArg.Name)
	to := args.GetString(toBranchArg.Name)
	if to == "" || from == "" {
		return nil, fmt.Errorf("not specified to_branch or from_branch")
	}

	var baseId string
	{
		_args1 := spec.EntrypointArgs{
			"cmd":            fmt.Sprintf("git merge-base %s %s", to, from),
			"split":          "",
			"extract_fields": "0",
			"query_columns":  "c0",
			"query_where":    "",
		}
		_, ep, _ := command.New()
		rets, err := ep(ctx, bundle, _args1)
		if err != nil {
			return nil, fmt.Errorf("%w: in git_check_merge function", err)
		}
		if len(rets) == 0 {
			return nil, fmt.Errorf("not found merge-base when check merge")
		}
		v, ok := rets["outcome"]
		if !ok {
			v = rets["outcome_0"]
		}
		baseId = v
	}

	{
		_args2 := spec.EntrypointArgs{
			"cmd":            fmt.Sprintf("git merge-tree %s %s %s", baseId, to, from),
			"split":          "",
			"extract_fields": "0",
			"query_columns":  "c0",
			"query_where":    "",
		}
		_, ep, _ := command.New()
		rets, err := ep(ctx, bundle, _args2)
		if err != nil {
			return nil, fmt.Errorf("%w: in git_check_merge function", err)
		}
		outcome := "outcome"
		m := map[string]string{
			outcome: "no-conflict",
		}
		if len(rets) == 0 {
			m[outcome] = "no-content-to-merge"
			return m, nil
		}
		for _, v := range rets {
			if strings.Contains(v, "changed in both") {
				m[outcome] = "conflict"
				break
			}
		}
		return m, nil
	}
}
