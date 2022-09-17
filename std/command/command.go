package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/manifest"
	"github.com/cofxlabs/cofx/pkg/output"
)

var cmdArg = manifest.UsageDesc{
	Name: "cmd",
	Desc: "Specify a command to run",
}

var envArg = manifest.UsageDesc{
	Name: "env",
	Desc: "Specify environment variables for the command",
}

var workingDirArg = manifest.UsageDesc{
	Name: "working_dir",
	Desc: "Specify working directory for the command",
}

var splitArg = manifest.UsageDesc{
	Name: "split",
	Desc: "Specify a separator to split",
}

var extractArg = manifest.UsageDesc{
	Name: "extract",
	Desc: "Specify one column or more to extract, .e.g 0,1,2",
}

var queryArg = manifest.UsageDesc{
	Name: "query",
	Desc: "Specify a SQL to qeury",
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
		Args: []manifest.UsageDesc{cmdArg, envArg, workingDirArg, splitArg, queryArg},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	// cmd
	cmdstr := args.GetString(cmdArg.Name)
	if cmdstr == "" {
		return nil, errors.New("command function miss argument: " + cmdArg.Name)
	}
	argv := strings.Fields(cmdstr)
	if len(args) == 0 {
		return nil, nil
	}
	name := argv[0]
	argv = argv[1:]

	// env
	env := args.GetStringSlice(envArg.Name)
	// split
	splitSep := args.GetString(splitArg.Name)
	columns, err := args.GetIntSlice(extractArg.Name)
	if err != nil {
		return nil, fmt.Errorf("%w: extract %s", err, args.GetString(extractArg.Name))
	}

	// working dir
	workingDir := args.GetString(workingDirArg.Name)
	if workingDir == "" {
		dir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		workingDir = dir
	}

	var outTable [][]string
	out := &output.Output{
		W: nil,
		HandleFunc: output.ColumnFunc(splitSep, func(columns []string) {
			// insert sqlite db
			outTable = append(outTable, columns)
		}, columns...),
	}
	defer out.Close()

	cmd := exec.CommandContext(ctx, name, argv...)
	cmd.Env = append(cmd.Env, env...)
	cmd.Dir = workingDir
	opipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	epipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		io.Copy(out, opipe)
		io.Copy(out, epipe)
	}()
	wg.Wait()
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	fmt.Fprintf(bundle.Resources.Logwriter, "---> %s\n", cmd.String())

	// query
	for i, row := range outTable {
		fmt.Fprintf(bundle.Resources.Logwriter, "%d: %v\n", i, row)
	}
	return nil, nil
}
