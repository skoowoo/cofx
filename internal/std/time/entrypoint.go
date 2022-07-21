package time

import (
	"context"
	"time"

	"github.com/cofunclabs/cofunc/pkg/manifest"
	"github.com/pkg/errors"
)

var _manifest = manifest.Manifest{
	Name:           "time",
	Description:    "",
	Driver:         "go",
	EntryPoint:     "",
	EntrypointFunc: Entrypoint,
	Args:           map[string]string{},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args: []manifest.UsageDesc{
			{
				Name: "format",
				OptionalValues: []string{
					"YYYY-MM-DD hh:mm:ss",
					"YYYY/MM/DD hh:mm:ss",
					"MM-DD-YYYY hh:mm:ss",
					"MM/DD/YYYY hh:mm:ss",
				},
				Desc: `Specifies the format for getting the current time`,
			},
		},
		ReturnValues: []manifest.UsageDesc{
			{
				Name: "Now",
				Desc: "Current time",
			},
		},
	},
}

func New() *manifest.Manifest {
	return &_manifest
}

func Entrypoint(ctx context.Context, args map[string]string) (map[string]string, error) {
	format, ok := args["format"]
	if !ok {
		format = "YYYY-MM-DD hh:mm:ss"
	}

	var now string
	switch format {
	case "YYYY-MM-DD hh:mm:ss":
		now = time.Now().Format("2006-01-02 15:04:05")
	case "YYYY/MM/DD hh:mm:ss":
		now = time.Now().Format("2006/01/02 15:04:05")
	case "MM-DD-YYYY hh:mm:ss":
		now = time.Now().Format("01-02-2006 15:04:05")
	case "MM/DD/YYYY hh:mm:ss":
		now = time.Now().Format("01/02/2006 15:04:05")
	default:
		return nil, errors.New("invalid format argument: " + format)
	}

	return map[string]string{
		"now": now,
	}, nil
}
