package gobuild

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cofunclabs/cofunc/pkg/manifest"
)

var _manifest = manifest.Manifest{
	Name:           "go_build",
	Description:    "For building go project that based on 'go mod'",
	Driver:         "go",
	EntrypointFunc: Entrypoint,
	Args:           map[string]string{},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args: []manifest.UsageDesc{
			{
				Name:           "prefix",
				Desc:           "By default, the module field contents are read from the 'go.mod' file",
				OptionalValues: nil,
			},
			{
				Name: "binpath",
				Desc: "",
			},
			{
				Name: "mainpkg_path",
				Desc: `Specifies the path of main package, if there are more than one, separated by ','.
 If not specified, the mainpkg is automatically parsed`,
			},
		},
		ReturnValues: []manifest.UsageDesc{},
	},
}

func New() *manifest.Manifest {
	return &_manifest
}

func Entrypoint(ctx context.Context, args map[string]string) (map[string]string, error) {
	prefix, ok := args["prefix"]
	if !ok {
		// TODO:
		_ = ok
	}
	binpath, ok := args["binpath"]
	if !ok {
		binpath = "bin"
	}
	mainpkg_path, ok := args["mainpkg_path"]
	if !ok {
		// TODO:
		_ = ok
	}
	paths := strings.Split(mainpkg_path, ",")
	for _, path := range paths {
		cmd, err := buildCommands(ctx, prefix, binpath, path)
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
	}

	return nil, nil
}

func buildCommands(ctx context.Context, prefix, binpath, mainpath string) (*exec.Cmd, error) {
	var args []string
	args = append(args, "build")
	args = append(args, "-o")
	args = append(args, binpath)
	args = append(args, filepath.Join(prefix, mainpath))

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, nil
}
