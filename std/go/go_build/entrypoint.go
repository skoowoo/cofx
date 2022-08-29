package gobuild

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/cofunclabs/cofunc/manifest"
	"github.com/cofunclabs/cofunc/pkg/textline"
)

var _manifest = manifest.Manifest{
	Name:        "go_build",
	Description: "For building go project that based on 'go mod'",
	Driver:      "go",
	Entrypoint:  "Entrypoint",
	Args: map[string]string{
		"bindir": "bin/",
	},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args: []manifest.UsageDesc{
			{
				Name:           "prefix",
				Desc:           "By default, the module field contents are read from the 'go.mod' file",
				OptionalValues: nil,
			},
			{
				Name: "bindir",
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

func New() (*manifest.Manifest, manifest.EntrypointFunc) {
	return &_manifest, Entrypoint
}

func Entrypoint(ctx context.Context, out io.Writer, version string, args map[string]string) (map[string]string, error) {
	bindir := args["bindir"]
	prefix, ok := args["prefix"]
	if !ok {
		var err error
		if prefix, err = textline.FindFileLine("go.mod", splitSpace, getPrefix); err != nil {
			return nil, err
		}
	}
	mainpkgPath, ok := args["mainpkg_path"]
	if !ok {
		// TODO:
		_ = ok
	}

	// print args
	fmt.Fprintf(out, "===> prefix      : %s\n", prefix)
	fmt.Fprintf(out, "===> mainpkg_path: %s\n", mainpkgPath)
	fmt.Fprintf(out, "===> bindir      : %s\n", bindir)

	paths := strings.Split(mainpkgPath, ",")
	for _, path := range paths {
		cmd, err := buildCommands(ctx, prefix, bindir, path, out)
		if err != nil {
			return nil, err
		}
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		if err := cmd.Wait(); err != nil {
			return nil, err
		}
		fmt.Fprintf(out, "---> %s\n", cmd.String())
	}

	return nil, nil
}

func buildCommands(ctx context.Context, prefix, binpath, mainpath string, w io.Writer) (*exec.Cmd, error) {
	var args []string
	args = append(args, "build")
	args = append(args, "-o")
	args = append(args, binpath)
	args = append(args, filepath.Join(prefix, mainpath))

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Stdout = w
	cmd.Stderr = w
	return cmd, nil
}

func splitSpace(c rune) bool {
	return unicode.IsSpace(c)
}

// getPrefix read 'module' field from go.mod file
func getPrefix(fields []string) (string, bool) {
	if len(fields) == 2 && fields[0] == "module" {
		return fields[1], true
	}
	return "", false
}

// getMainpkgPath search the main package
func getMainpkgPath(fields []string) (string, bool) {
	return "", false
}
