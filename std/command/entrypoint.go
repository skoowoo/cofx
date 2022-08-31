package command

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/cofunclabs/cofunc/functiondriver/go/spec"
	"github.com/cofunclabs/cofunc/manifest"
)

var _manifest = manifest.Manifest{
	Name:   "command",
	Driver: "go",
	Args: map[string]string{
		"script": "",
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	script := args["script"]
	if script == "" {
		return nil, errors.New("command function miss 'script' argument")
	}
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", script)
	cmd.Stdout = bundle.Logger
	cmd.Stderr = bundle.Logger
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	fmt.Fprintf(bundle.Logger, "---> %s\n", cmd.String())
	return nil, nil
}
