package gogenerate

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/cofunclabs/cofunc/functiondriver/go/spec"
	"github.com/cofunclabs/cofunc/manifest"
)

var _manifest = manifest.Manifest{
	Category:       "go",
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

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	cmd, err := buildCommands(ctx, bundle.Resources.Logwriter)
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	fmt.Fprintf(bundle.Resources.Logwriter, "---> %s\n", cmd.String())

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
