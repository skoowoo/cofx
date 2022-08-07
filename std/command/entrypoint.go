package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/cofunclabs/cofunc/manifest"
)

var _manifest = manifest.Manifest{
	Name:           "command",
	Driver:         "go",
	EntrypointFunc: Entrypoint,
	Args: map[string]string{
		"script": "",
	},
}

func New() *manifest.Manifest {
	return &_manifest
}

func Entrypoint(ctx context.Context, version string, args map[string]string) (map[string]string, error) {
	script := args["script"]
	if script == "" {
		return nil, errors.New("command function miss 'script' argument")
	}
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	fmt.Printf("---> %s\n", cmd.String())
	return nil, nil
}
