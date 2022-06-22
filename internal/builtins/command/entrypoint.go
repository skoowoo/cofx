package command

import (
	"context"
	"errors"
	"os"
	"os/exec"

	"github.com/cofunclabs/cofunc/pkg/manifest"
)

func New() manifest.Manifester {
	return &Command{}
}

type Command struct{}

func (c *Command) Name() string {
	return "command"
}

func (c *Command) Manifest() manifest.Manifest {
	return manifest.Manifest{
		Driver:         "go",
		EntrypointFunc: c.Entrypoint,
		Args: map[string]string{
			"script": "",
		},
	}
}

func (c *Command) Entrypoint(ctx context.Context, args map[string]string) (map[string]string, error) {
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
	return nil, nil
}
