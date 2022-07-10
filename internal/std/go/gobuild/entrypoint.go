package gobuild

import (
	"context"

	"github.com/cofunclabs/cofunc/pkg/manifest"
)

var _manifest = manifest.Manifest{
	Description:    "A tool for building go project that based on 'go mod'",
	Driver:         "go",
	EntryPoint:     "",
	EntrypointFunc: nil,
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
				Name: "mainpkg",
				Desc: `Specifies the location of main package, if there are more than one, separated by ','.
 If not specified, the mainpkg is automatically parsed`,
			},
			{
				Name:           "generate",
				OptionalValues: []string{"true", "false"},
				Desc:           "",
			},
		},
		ReturnValues: []manifest.UsageDesc{},
	},
}

func New() manifest.Manifester {
	return &_gobuild{}
}

type _gobuild struct{}

func (f *_gobuild) Name() string {
	return "gobuild"
}

func (f *_gobuild) Manifest() manifest.Manifest {
	_manifest.EntrypointFunc = f.Entrypoint
	return _manifest
}

func (f *_gobuild) Entrypoint(ctx context.Context, args map[string]string) (map[string]string, error) {
	return nil, nil
}
