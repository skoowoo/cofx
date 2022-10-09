package gotest

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

var _manifest = manifest.Manifest{
	Category:       "go",
	Name:           "go_test",
	Description:    "Wraps the 'go test' unit testing command",
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
	fails, oks, nos, cmd, err := execGoTestCommand(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(bundle.Resources.Logwriter, "---> %s\n", cmd)

	for _, fail := range fails {
		fmt.Fprintf(bundle.Resources.Logwriter, "%s\n", strings.Join(fail, "  "))
	}
	for _, ok := range oks {
		fmt.Fprintf(bundle.Resources.Logwriter, "%s\n", strings.Join(ok, "  "))
	}

	return map[string]string{"outcome": fmt.Sprintf("pkg fail:%d ok:%d no test:%d", len(fails), len(oks), len(nos))}, nil
}

func execGoTestCommand(ctx context.Context) ([][]string, [][]string, [][]string, string, error) {
	var (
		fails [][]string
		oks   [][]string
		nos   [][]string
		sep   string
		buff  bytes.Buffer
	)
	out := &output.Output{
		W: &buff,
		HandleFunc: output.ColumnFunc(sep, func(fields []string) {
			if fields[0] != "?" && fields[0] != "ok" {
				fails = append(fails, fields)
				return
			}
			if fields[0] == "ok" {
				oks = append(oks, fields)
				return
			}
			if fields[0] == "?" {
				nos = append(nos, fields)
				return
			}
		}, 0, 1, 2, 3, 4),
	}
	cmd := exec.CommandContext(ctx, "go", "test", "-covermode=count", "-coverprofile=/tmp/cover.out", "./...")
	cmd.Stdout = out
	cmd.Stderr = out
	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("%w: %s", err, buff.String())
	}
	return fails, oks, nos, cmd.String(), err
}
