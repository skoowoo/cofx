package git

import (
	"context"

	"github.com/skoowoo/cofx/pkg/runcmd"
)

// GetGitDir returns the root directory of the current git project.
// 'git rev-parse --show-toplevel'
func GetGitDir(ctx context.Context) (string, error) {
	wrap := runcmd.Wrap{
		Name:         "git",
		Args:         []string{"rev-parse", "--show-toplevel"},
		Split:        "",
		Extract:      []int{0},
		QueryColumns: []string{"c0"},
		QueryWhere:   "",
	}
	rows, err := wrap.Run(ctx)
	if err != nil {
		return "", err
	}
	return rows.String(0, 0), nil
}
