package gitinsight

import (
	"context"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/manifest"
)

var (
	branchCountRet = manifest.UsageDesc{
		Name: "branch_count",
		Desc: "Branch Count in the git local repo",
	}
	mergedBranchRet = manifest.UsageDesc{
		Name: "merged_branch",
		Desc: "List of merged branch in the git local repo, separated by comma",
	}
	unmergedBranchRet = manifest.UsageDesc{
		Name: "unmerged_branch",
		Desc: "List of unmerged branch in the git local repo, separated by comma",
	}
	stashCountRet = manifest.UsageDesc{
		Name: "stash_count",
		Desc: "Stash Count in the git local repo",
	}
	commitCountRet = manifest.UsageDesc{
		Name: "commit_count",
		Desc: "Commit Count of the current branch in the git local repo",
	}
	behindCountWithOriginRet = manifest.UsageDesc{
		Name: "behind_count_with_origin",
		Desc: "Behind Count of the current branch with origin in the git local repo",
	}
	behindCountWithUpstreamRet = manifest.UsageDesc{
		Name: "behind_count_with_upstream",
		Desc: "Behind Count of the current branch with upstream in the git local repo",
	}
	conflictWithUpstreamRet = manifest.UsageDesc{
		Name: "conflict_with_upstream",
		Desc: "Conflict or not with upstream in the git local repo",
	}
	lastCommitOriginRet = manifest.UsageDesc{
		Name: "last_commit_origin",
		Desc: "Last commit of the main branch in the origin repo",
	}
	lastCommitUpstreamRet = manifest.UsageDesc{
		Name: "last_commit_upstream",
		Desc: "Last commit of the main branch in the upstream repo",
	}
	lastCommitHeadRet = manifest.UsageDesc{
		Name: "last_commit_head",
		Desc: "Last commit of the HEAD in the local repo",
	}
)

var _manifest = manifest.Manifest{
	Category:       "git",
	Name:           "git_insight",
	Description:    "Analyze status data of the local git repository",
	Driver:         "go",
	Args:           map[string]string{},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args: []manifest.UsageDesc{},
		ReturnValues: []manifest.UsageDesc{
			branchCountRet,
			mergedBranchRet,
			unmergedBranchRet,
			stashCountRet,
			commitCountRet,
			behindCountWithOriginRet,
			behindCountWithUpstreamRet,
			conflictWithUpstreamRet,
			lastCommitOriginRet,
			lastCommitUpstreamRet,
			lastCommitHeadRet,
		},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	m := make(map[string]string)
	return m, nil
}
