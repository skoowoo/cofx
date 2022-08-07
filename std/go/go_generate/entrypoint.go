package gogenerate

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/cofunclabs/cofunc/manifest"
)

var _manifest = manifest.Manifest{
	Name:           "go_generate",
	Description:    "A function that packaging 'go generate' command",
	Driver:         "go",
	EntrypointFunc: Entrypoint,
	Args:           map[string]string{},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{},
		ReturnValues: []manifest.UsageDesc{},
	},
}

func New() *manifest.Manifest {
	return &_manifest
}

func Entrypoint(ctx context.Context, version string, args map[string]string) (map[string]string, error) {
	cmd, err := buildCommands(ctx)
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	fmt.Printf("---> %s\n", cmd.String())

	return nil, nil
}

func buildCommands(ctx context.Context) (*exec.Cmd, error) {
	var args []string
	args = append(args, "generate")
	args = append(args, "./...")

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, nil
}
