package outcome

import (
	"context"
	"fmt"

	"github.com/cofunclabs/cofunc/functiondriver/go/spec"
	"github.com/cofunclabs/cofunc/manifest"
	"github.com/cofunclabs/cofunc/pkg/stringutil"
)

var _manifest = manifest.Manifest{
	Name:           "outcome",
	Description:    "Used to collect and output the result of all functions",
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
	for k, v := range args {
		fmt.Fprintf(bundle.Resources.Logwriter, "%s\n", k)
		slice := stringutil.String2Slice(v)
		for _, s := range slice {
			fmt.Fprintf(bundle.Resources.Logwriter, "  ➜ %s\n", s)
		}
	}
	return nil, nil
}
