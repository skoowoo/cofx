package gobuild

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cofunclabs/cofunc/functiondriver/go/spec"
	"github.com/cofunclabs/cofunc/manifest"
)

var (
	prefixArg = manifest.UsageDesc{
		Name: "prefix",
		Desc: "By default, the module field contents are read from the 'go.mod' file",
	}
	binFormatArg = manifest.UsageDesc{
		Name: "bin_format",
		Desc: "Specifies the format of the binary file that to be built",
	}
	mainpkgArg = manifest.UsageDesc{
		Name: "find_mainpkg_dirs",
		Desc: `Specifies the dirs to find main package, if there are more than one, separated by ','. If not specified, it will find it from current dir.`,
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
		binFormatArg.Name: "bin/",
		mainpkgArg.Name:   ".",
	},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{prefixArg, binFormatArg, mainpkgArg},
		ReturnValues: []manifest.UsageDesc{outcomeRet},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	bins, err := parseBinFormats(args.GetStringSlice(binFormatArg.Name))
	if err != nil {
		return nil, err
	}

	mods, err := findMods(".", args.GetStringSlice(mainpkgArg.Name))
	if err != nil {
		return nil, err
	}
	for _, m := range mods {
		fmt.Fprintf(bundle.Resources.Logwriter, "mod %s in %s\n", m.name, m.dir)
	}

	var outcomes []string
	for _, mod := range mods {
		for _, pkg := range mod.mainpkgs {
			for _, bin := range bins {
				dstbin := bin.fullBinPath(filepath.Base(pkg))
				cmd, err := buildCommand(ctx, dstbin, pkg, bundle.Resources.Logwriter)
				if err != nil {
					return nil, err
				}
				cmd.Env = append(os.Environ(), bin.envs()...)
				cmd.Dir = mod.dir

				fmt.Fprintf(bundle.Resources.Logwriter, "---> %s\n", cmd.String())

				if err := cmd.Start(); err != nil {
					return nil, err
				}
				if err := cmd.Wait(); err != nil {
					return nil, err
				}
				outcomes = append(outcomes, filepath.Join(mod.dir, dstbin))
			}
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
