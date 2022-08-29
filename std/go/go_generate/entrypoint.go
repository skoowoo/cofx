package gogenerate

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/cofunclabs/cofunc/manifest"
)

var _manifest = manifest.Manifest{
	Name:           "go_generate",
	Description:    "A function that packaging 'go generate' command",
	Driver:         "go",
	Args:           map[string]string{},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{},
		ReturnValues: []manifest.UsageDesc{},
	},
}

func New() (*manifest.Manifest, manifest.EntrypointFunc) {
	return &_manifest, Entrypoint
}

func Entrypoint(ctx context.Context, out io.Writer, version string, args map[string]string) (map[string]string, error) {
	cmd, err := buildCommands(ctx, out)
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	fmt.Fprintf(out, "---> %s\n", cmd.String())

	return nil, nil
}

func buildCommands(ctx context.Context, w io.Writer) (*exec.Cmd, error) {
	var args []string
	args = append(args, "generate")
	args = append(args, "./...")

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Stdout = w
	cmd.Stderr = w
	return cmd, nil
}
