package gitinsight

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/manifest"
	"github.com/cofxlabs/cofx/pkg/runcmd"
	"github.com/cofxlabs/cofx/pkg/textparse"
	"github.com/cofxlabs/cofx/std/git"
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
	lastCommitMainRet = manifest.UsageDesc{
		Name: "last_commit_main",
		Desc: "Last commit of the main branch in the local repo",
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
			stashCountRet,
			commitCountRet,
			behindCountWithOriginRet,
			behindCountWithUpstreamRet,
			conflictWithUpstreamRet,
			lastCommitOriginRet,
			lastCommitUpstreamRet,
			lastCommitMainRet,
			lastCommitHeadRet,
		},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	m := make(map[string]string)

	{
		branches, err := getBranches(ctx, bundle)
		if err != nil {
			return nil, err
		}
		m[branchCountRet.Name] = strconv.Itoa(len(branches))
	}
	{
		stash, err := getStash(ctx, bundle)
		if err != nil {
			return nil, err
		}
		m[stashCountRet.Name] = strconv.Itoa(len(stash))
	}
	{
		commitCount, err := getCommitCount(ctx, bundle)
		if err != nil {
			return nil, err
		}
		m[commitCountRet.Name] = strconv.Itoa(commitCount)
	}
	{
		dir, err := git.GetGitDir(ctx)
		if err != nil {
			return nil, err
		}
		commits, err := getLastCommit(ctx, bundle, dir)
		if err != nil {
			return nil, err
		}
		for k, commit := range commits {
			m[k] = commit[1]
		}

		_, cm, err := getLastCommitHead(ctx, bundle, dir)
		if err != nil {
			return nil, err
		}
		m[lastCommitHeadRet.Name] = cm
	}
	return m, nil
}

// 'cat .git/refs/heads/main'
// 'cat .git/refs/remotes/origin/main'
// 'cat .git/refs/remotes/upstream/main'
func getLastCommit(ctx context.Context, bundle spec.EntrypointBundle, dir string) (map[string][]string, error) {
	commits := make(map[string][]string, 3)
	{
		p, s, err := textparse.ReadFile(
			filepath.Join(dir, ".git/refs/heads/main"),
			filepath.Join(dir, ".git/refs/heads/master")).String()
		if err != nil {
			return nil, err
		}
		p = strings.TrimPrefix(p, filepath.Join(dir, ".git/refs/heads/"))
		commits[lastCommitMainRet.Name] = []string{p, s}
	}
	{
		p, s, err := textparse.ReadFile(
			filepath.Join(dir, ".git/refs/remotes/origin/main"),
			filepath.Join(dir, ".git/refs/remotes/origin/master")).String()
		if err != nil {
			return nil, err
		}
		p = strings.TrimPrefix(p, filepath.Join(dir, ".git/refs/remotes/"))
		commits[lastCommitOriginRet.Name] = []string{p, s}
	}
	{
		p, s, err := textparse.ReadFile(
			filepath.Join(dir, ".git/refs/remotes/upstream/main"),
			filepath.Join(dir, ".git/refs/remotes/upstream/master")).String()
		if err != nil {
			return nil, err
		}
		p = strings.TrimPrefix(p, filepath.Join(dir, ".git/refs/remotes/"))
		commits[lastCommitUpstreamRet.Name] = []string{p, s}
	}
	return commits, nil
}

// 'cat .git/HEAD'
func getLastCommitHead(ctx context.Context, bundle spec.EntrypointBundle, dir string) (string, string, error) {
	var (
		path   string
		commit string
	)
	{
		nst, err := textparse.New(".git/HEAD", ":", []int{1})
		if err != nil {
			return "", "", err
		}
		if err := nst.ParseFile(ctx, filepath.Join(dir, ".git/HEAD")); err != nil {
			return "", "", err
		}
		if path, err = nst.String(ctx, "c0", ""); err != nil {
			return "", "", err
		}
	}
	{
		_, s, err := textparse.ReadFile(filepath.Join(dir, ".git", path)).String()
		if err != nil {
			return "", "", err
		}
		commit = s
	}
	return path, commit, nil
}

// 'git branch -al'
func getBranches(ctx context.Context, bundle spec.EntrypointBundle) ([]string, error) {
	wrap := runcmd.Wrap{
		Name:         "git",
		Args:         []string{"branch", "-al"},
		Split:        "",
		Extract:      []int{0},
		QueryColumns: []string{"c0"},
		QueryWhere:   "",
	}
	rows, err := wrap.Run(ctx)
	if err != nil {
		return nil, err
	}
	return rows.Column2Slice(0), nil
}

// 'git stash list'
func getStash(ctx context.Context, bundle spec.EntrypointBundle) ([]string, error) {
	wrap := runcmd.Wrap{
		Name:         "git",
		Args:         []string{"stash", "list"},
		Split:        "",
		Extract:      []int{0},
		QueryColumns: []string{"c0"},
		QueryWhere:   "",
	}
	rows, err := wrap.Run(ctx)
	if err != nil {
		return nil, err
	}
	return rows.Column2Slice(0), nil
}

// 'git rev-list --all --count'
func getCommitCount(ctx context.Context, bundle spec.EntrypointBundle) (int, error) {
	wrap := runcmd.Wrap{
		Name:         "git",
		Args:         []string{"rev-list", "--all", "--count"},
		Split:        "",
		Extract:      []int{0},
		QueryColumns: []string{"c0"},
		QueryWhere:   "",
	}
	rows, err := wrap.Run(ctx)
	if err != nil {
		return 0, err
	}
	return rows.Int(0, 0)
}
