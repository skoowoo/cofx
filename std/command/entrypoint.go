package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"

	"github.com/cofunclabs/cofunc/manifest"
)

var _manifest = manifest.Manifest{
	Name:   "command",
	Driver: "go",
	Args: map[string]string{
		"script": "",
	},
}

func New() (*manifest.Manifest, manifest.EntrypointFunc) {
	return &_manifest, Entrypoint
}

func Entrypoint(ctx context.Context, out io.Writer, version string, args map[string]string) (map[string]string, error) {
	script := args["script"]
	if script == "" {
		return nil, errors.New("command function miss 'script' argument")
	}
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", script)
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	fmt.Fprintf(out, "---> %s\n", cmd.String())
	return nil, nil
}
