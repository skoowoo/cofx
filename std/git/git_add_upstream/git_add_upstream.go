package gitaddupstream

import (
	"context"
	"fmt"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/manifest"
	"github.com/cofxlabs/cofx/std/command"
)

var upstreamUrlArg = manifest.UsageDesc{
	Name: "upstream_url",
	Desc: "Specify the upstream repo url",
}

var _manifest = manifest.Manifest{
	Category:       "git",
	Name:           "git_add_upstream",
	Description:    "Use 'git remote add' to add a remote upstream repo.",
	Driver:         "go",
	Args:           map[string]string{},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{upstreamUrlArg},
		ReturnValues: []manifest.UsageDesc{},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	url := args.GetString(upstreamUrlArg.Name)
	if url == "" {
		return nil, fmt.Errorf("not specified upstream_url")
	}

	_args := spec.EntrypointArgs{
		"cmd":            "git remote add upstream " + url,
		"split":          "",
		"extract_fields": "",
		"query_columns":  "",
		"query_where":    "",
	}
	_, ep, _ := command.New()
	rets, err := ep(ctx, bundle, _args)
	if err != nil {
		return nil, fmt.Errorf("%w: in git_add_upstream function", err)
	}
	return rets, nil
}
