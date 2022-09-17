package syncupstream

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/manifest"
	"github.com/cofxlabs/cofx/pkg/output"
)

var branchArg = manifest.UsageDesc{
	Name: "branch",
	Desc: "Specify branches to sync, multiple branches are separated by ',', default main and master",
}

var upstreamArg = manifest.UsageDesc{
	Name: "upstream",
	Desc: "Specify upstream to sync, it not set, will try to find out it from 'git remote -v'",
}

var _manifest = manifest.Manifest{
	Category:    "git",
	Name:        "git_sync_upstream",
	Description: "Sync git branch from upstream",
	Driver:      "go",
	Args: map[string]string{
		branchArg.Name: "main,master",
	},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{branchArg, upstreamArg},
		ReturnValues: []manifest.UsageDesc{},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	branches := args.GetStringSlice(branchArg.Name)
	currentBranch, cmd, err := getCurrentBranch(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(bundle.Resources.Logwriter, "---> %s ➜ %s\n", cmd, currentBranch)

	var found bool
	for _, branch := range branches {
		if branch == currentBranch {
			found = true
			break
		}
	}
	if !found {
		return map[string]string{"outcome": "no sync: not sync this branch"}, nil
	}

	upstream, cmd, err := getUpstreamAddr(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(bundle.Resources.Logwriter, "---> %s ➜ %s\n", cmd, upstream)

	if upstream == "" {
		if upstream = args.GetString(upstreamArg.Name); upstream == "" {
			return map[string]string{"outcome": "no sync: not found upstream"}, nil
		}
		if cmd, err := addUpstream(ctx, upstream); err != nil {
			return nil, err
		} else {
			fmt.Fprintf(bundle.Resources.Logwriter, "---> %s\n", cmd)
		}
	}
	// git fetch --all
	if cmd, err := fetchRemotes(ctx); err != nil {
		return nil, err
	} else {
		fmt.Fprintf(bundle.Resources.Logwriter, "---> %s\n", cmd)
	}
	// git merge-base
	// git merge-tree
	state1, err := checkUpstreamDiff(ctx, currentBranch)
	if err != nil {
		return nil, err
	}
	state2, err := checkOriginDiff(ctx, currentBranch)
	if err != nil {
		return nil, err
	}
	switch state1 {
	case consistent:
		if state2 == consistent {
			return map[string]string{"outcome": "no sync: three branches are consistent"}, nil
		} else {
			if cmd, err := pushOrigin(ctx, currentBranch); err != nil {
				return nil, err
			} else {
				fmt.Fprintf(bundle.Resources.Logwriter, "---> %s\n", cmd)
			}
		}
		return nil, nil
	case conflict:
		return map[string]string{"outcome": "no sync: branches are conflict"}, nil
	case noConflict:
	}

	// git rebase upstream/branch
	if err := rebaseUpstream(ctx, currentBranch); err != nil {
		return nil, err
	}

	// git push origin branch
	if cmd, err := pushOrigin(ctx, currentBranch); err != nil {
		return nil, err
	} else {
		fmt.Fprintf(bundle.Resources.Logwriter, "---> %s\n", cmd)
	}

	return map[string]string{"outcome": "synced"}, nil
}

func addUpstream(ctx context.Context, upstream string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "remote", "add", "upstream", upstream)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, string(out))
	}
	return cmd.String(), nil
}

func pushOrigin(ctx context.Context, branch string) (string, error) {
	var (
		lasterr error
		cmdstr  string
	)
	for i := 0; i < 3; i++ {
		cmd := exec.CommandContext(ctx, "git", "push", "origin", branch)
		out, err := cmd.CombinedOutput()
		cmdstr = cmd.String()
		if err != nil {
			lasterr = fmt.Errorf("%w: %s", err, string(out))
			continue
		}
		return cmdstr, nil
	}
	return cmdstr, lasterr
}

func rebaseUpstream(ctx context.Context, branch string) error {
	cmd := exec.CommandContext(ctx, "git", "rebase", "upstream/"+branch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(out))
	}
	return nil
}

type mergeBaseState int

const (
	unknow mergeBaseState = iota
	consistent
	conflict
	noConflict
)

func checkUpstreamDiff(ctx context.Context, branch string) (mergeBaseState, error) {
	return checkMergeBase(ctx, branch, "upstream/"+branch)
}

func checkOriginDiff(ctx context.Context, branch string) (mergeBaseState, error) {
	return checkMergeBase(ctx, "origin/"+branch, branch)
}

func checkMergeBase(ctx context.Context, toBranch, fromBranch string) (mergeBaseState, error) {
	var commitId string
	{
		cmd := exec.CommandContext(ctx, "git", "merge-base", toBranch, fromBranch)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return unknow, fmt.Errorf("%w: %s", err, string(out))
		}
		commitId = strings.TrimSpace(string(out))
	}
	{
		cmd := exec.CommandContext(ctx, "git", "merge-tree", commitId, toBranch, fromBranch)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return unknow, fmt.Errorf("%w: %s", err, string(out))
		}
		if len(out) == 0 {
			return consistent, nil
		}
		if bytes.Contains(out, []byte("\nchanged in both")) {
			return conflict, nil
		}
	}
	return noConflict, nil
}

func fetchRemotes(ctx context.Context) (string, error) {
	var (
		lasterr error
		cmdstr  string
	)
	for i := 0; i < 3; i++ {
		cmd := exec.CommandContext(ctx, "git", "fetch", "--all")
		out, err := cmd.CombinedOutput()
		cmdstr = cmd.String()
		if err != nil {
			lasterr = fmt.Errorf("%w: %s", err, string(out))
			continue
		}
		return cmdstr, nil
	}
	return cmdstr, lasterr
}

func getCurrentBranch(ctx context.Context) (string, string, error) {
	cmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", cmd.String(), fmt.Errorf("%w: %s", err, string(out))
	}
	return strings.TrimSpace(string(out)), cmd.String(), nil
}

func getUpstreamAddr(ctx context.Context) (string, string, error) {
	var (
		rows [][]string
		sep  string
	)
	out := &output.Output{
		W: nil,
		HandleFunc: output.ColumnFunc(sep, func(fields []string) {
			if fields[0] == "upstream" && strings.Contains(fields[2], "fetch") {
				rows = append(rows, fields)
			}
		}, 0, 1, 2),
	}

	cmd := exec.CommandContext(ctx, "git", "remote", "-v")
	cmd.Stderr = out
	cmd.Stdout = out
	if err := cmd.Run(); err != nil {
		return "", cmd.String(), err
	}
	if len(rows) != 0 {
		return rows[0][1], cmd.String(), nil
	} else {
		return "", cmd.String(), nil
	}
}
