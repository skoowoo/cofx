package gitbasic

import (
	"context"
	"fmt"
	"strings"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/manifest"
	"github.com/cofxlabs/cofx/std/command"
	"github.com/cofxlabs/cofx/std/git"
)

var (
	originRet = manifest.UsageDesc{
		Name: "origin",
		Desc: "The origin url of the git local repo",
	}
	upstreamRet = manifest.UsageDesc{
		Name: "upstream",
		Desc: "The upstream url of the git local repo",
	}
	localLocationRet = manifest.UsageDesc{
		Name: "local_location",
		Desc: "Repo directory of the git local repo",
	}
	branchRet = manifest.UsageDesc{
		Name: "current_branch",
		Desc: "The current branch name of the git local repo",
	}
	orgRet = manifest.UsageDesc{
		Name: "github_org",
		Desc: "The github org name that the git local repo forked from",
	}
	repoRet = manifest.UsageDesc{
		Name: "github_repo",
		Desc: "The github repo name that the git local repo forked from",
	}
)

var _manifest = manifest.Manifest{
	Category:       "git",
	Name:           "git_basic",
	Description:    "Read common basic information of local git repository",
	Driver:         "go",
	Args:           map[string]string{},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args: []manifest.UsageDesc{},
		ReturnValues: []manifest.UsageDesc{
			originRet,
			upstreamRet,
			localLocationRet,
			branchRet,
			orgRet,
			repoRet,
		},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	m := make(map[string]string)
	// Get remotes
	// upstream	https://github.com/cofxlabs/cofx.git (fetch)
	{
		_args := spec.EntrypointArgs{
			"cmd":            "git remote -v",
			"split":          "",
			"extract_fields": "0,1,2",
			"query_columns":  "c0,c1",
			"query_where":    "c2 like '%fetch%'",
		}
		_, ep, _ := command.New()
		rets, err := ep(ctx, bundle, _args)
		if err != nil {
			return nil, fmt.Errorf("%w: in git_basic function", err)
		}
		for _, v := range rets {
			fields := strings.Fields(v)
			if len(fields) == 2 {
				m[fields[0]] = fields[1]
			}
		}
	}

	// Get github org and repo name
	// e.g. https://github.com/skoo87/cofx.git
	origin, ok := m["origin"]
	if ok {
		if strings.Contains(origin, "https://github.com") {
			fields := strings.Split(origin, "/")
			if len(fields) == 5 {
				m[orgRet.Name] = fields[3]
				m[repoRet.Name] = strings.TrimSuffix(fields[4], ".git")
			}
		}
	}

	// Get current branch
	{
		_args := spec.EntrypointArgs{
			// "cmd":            "git branch --show-current",
			"cmd":            "git rev-parse --abbrev-ref HEAD",
			"split":          "",
			"extract_fields": "0",
			"query_columns":  "c0",
			"query_where":    "",
		}
		_, ep, _ := command.New()
		rets, err := ep(ctx, bundle, _args)
		if err != nil {
			return nil, fmt.Errorf("%w: in git_basic function", err)
		}
		for _, v := range rets {
			m[branchRet.Name] = v
			break
		}
	}
	// Get local location (directory)
	{
		dir, err := git.GetGitDir(ctx)
		if err != nil {
			return nil, fmt.Errorf("%w: in git_basic function", err)
		}
		m[localLocationRet.Name] = dir
	}
	return m, nil
}
