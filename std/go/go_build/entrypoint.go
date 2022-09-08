package gobuild

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/cofunclabs/cofunc/functiondriver/go/spec"
	"github.com/cofunclabs/cofunc/manifest"
	"github.com/cofunclabs/cofunc/pkg/textline"
)

var (
	prefixArg = manifest.UsageDesc{
		Name: "prefix",
		Desc: "By default, the module field contents are read from the 'go.mod' file",
	}
	bindirArg = manifest.UsageDesc{
		Name: "bindir",
		Desc: "",
	}
	mainpkgArg = manifest.UsageDesc{
		Name: "mainpkg_path",
		Desc: `Specifies the path of main package, if there are more than one, separated by ','. If not specified, the mainpkg is automatically parsed`,
	}
)

var _manifest = manifest.Manifest{
	Name:        "go_build",
	Description: "For building go project that based on 'go mod'",
	Driver:      "go",
	Args: map[string]string{
		"bindir": "bin/",
	},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{prefixArg, bindirArg, mainpkgArg},
		ReturnValues: []manifest.UsageDesc{},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	bindir := args.GetString(bindirArg.Name)
	prefix := args.GetString(prefixArg.Name)
	if prefix == "" {
		var err error
		if prefix, err = textline.FindFileLine("go.mod", splitSpace, getPrefix); err != nil {
			return nil, err
		}
	}
	mainpkgPath := args.GetString(mainpkgArg.Name)
	if mainpkgPath == "" {
		// TODO:
		_ = mainpkgPath
	}

	// print args
	fmt.Fprintf(bundle.Resources.Logwriter, "===> prefix      : %s\n", prefix)
	fmt.Fprintf(bundle.Resources.Logwriter, "===> mainpkg_path: %s\n", mainpkgPath)
	fmt.Fprintf(bundle.Resources.Logwriter, "===> bindir      : %s\n", bindir)

	var (
		platforms = map[string][]string{
			"linux":   {"CGO_ENABLED=0", "GOOS=linux", "GOARCH=amd64"},
			"windows": {"CGO_ENABLED=0", "GOOS=windows", "GOARCH=amd64"},
			"darwin":  {"CGO_ENABLED=0", "GOOS=darwin", "GOARCH=amd64"},
		}
	)
	paths := strings.Split(mainpkgPath, ",")
	for _, path := range paths {
		for p, env := range platforms {
			dstdir := filepath.Join(bindir, p) + "/"
			cmd, err := buildCommand(ctx, prefix, dstdir, path, bundle.Resources.Logwriter)
			if err != nil {
				return nil, err
			}
			cmd.Env = append(os.Environ(), env...)
			if err := cmd.Start(); err != nil {
				return nil, err
			}
			if err := cmd.Wait(); err != nil {
				return nil, err
			}
			fmt.Fprintf(bundle.Resources.Logwriter, "---> %s\n", cmd.String())
		}
	}

	return nil, nil
}

func buildCommand(ctx context.Context, prefix, binpath, mainpath string, w io.Writer) (*exec.Cmd, error) {
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
