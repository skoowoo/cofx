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
	binoutArg = manifest.UsageDesc{
		Name: "bin_out",
		Desc: "Specifies the format of the binary file that to be built",
	}
	mainpkgArg = manifest.UsageDesc{
		Name: "mainpkg_path",
		Desc: `Specifies the path of main package, if there are more than one, separated by ','. If not specified, the mainpkg is automatically parsed`,
	}
)

var (
	outcomeRet = manifest.UsageDesc{
		Name: "outcome",
	}
)

var _manifest = manifest.Manifest{
	Name:        "go_build",
	Description: "For building go project that based on 'go mod'",
	Driver:      "go",
	Args: map[string]string{
		"bin_out": "bin/",
	},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{prefixArg, binoutArg, mainpkgArg},
		ReturnValues: []manifest.UsageDesc{outcomeRet},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	binouts := strings.FieldsFunc(args.GetString(binoutArg.Name), func(r rune) bool {
		return r == ',' || r == '\n'
	})
	var bins []binaryOutFormat
	for _, binout := range binouts {
		bf, err := parseBinout(strings.TrimSpace(binout))
		if err != nil {
			return nil, fmt.Errorf("%w: %s", err, binout)
		}
		bins = append(bins, bf)
	}

	prefix := args.GetString(prefixArg.Name)
	if prefix == "" {
		var err error
		if prefix, err = textline.FindFileLine("go.mod", splitSpace, getPrefix); err != nil {
			return nil, err
		}
	}

	var mainpkgs []string
	mainpkgPath := args.GetString(mainpkgArg.Name)
	if mainpkgPath != "" {
		mainpkgs = strings.Split(mainpkgPath, ",")
		for i, mainpkg := range mainpkgs {
			mainpkgs[i] = filepath.Join(prefix, strings.TrimSpace(mainpkg))
		}
	} else {
		var err error
		mainpkgs, err = findMainPkg(prefix, ".")
		if err != nil {
			return nil, err
		}
	}

	var outcomes []string
	for _, mainpkg := range mainpkgs {
		for _, bin := range bins {
			dstbin := bin.fullBinPath(filepath.Base(mainpkg))
			cmd, err := buildCommand(ctx, dstbin, mainpkg, bundle.Resources.Logwriter)
			if err != nil {
				return nil, err
			}
			cmd.Env = append(os.Environ(), bin.envs()...)
			if err := cmd.Start(); err != nil {
				return nil, err
			}
			if err := cmd.Wait(); err != nil {
				return nil, err
			}
			fmt.Fprintf(bundle.Resources.Logwriter, "---> %s\n", cmd.String())
			outcomes = append(outcomes, dstbin)
		}
	}

	return map[string]string{
		outcomeRet.Name: strings.Join(outcomes, ","),
	}, nil
}

func buildCommand(ctx context.Context, binpath, mainpath string, w io.Writer) (*exec.Cmd, error) {
	var args []string
	args = append(args, "build")
	args = append(args, "-o")
	args = append(args, binpath)
	args = append(args, mainpath)

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
