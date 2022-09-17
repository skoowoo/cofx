package command

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/manifest"
)

var cmdArg = manifest.UsageDesc{
	Name: "cmd",
	Desc: "specify a command to run",
}

var _manifest = manifest.Manifest{
	Name:           "command",
	Description:    "Used to run a command",
	Driver:         "go",
	Entrypoint:     "",
	Args:           map[string]string{},
	RetryOnFailure: 0,
	IgnoreFailure:  false,
	Usage: manifest.Usage{
		Args: []manifest.UsageDesc{cmdArg},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	s := args.GetString(cmdArg.Name)
	if s == "" {
		return nil, errors.New("command function miss argument: " + cmdArg.Name)
	}
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", s)
	cmd.Stdout = bundle.Resources.Logwriter
	cmd.Stderr = bundle.Resources.Logwriter
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	fmt.Fprintf(bundle.Resources.Logwriter, "---> %s\n", cmd.String())
	return nil, nil
}
