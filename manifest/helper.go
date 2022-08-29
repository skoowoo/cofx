package manifest

import (
	"context"
	"io"
)

type EntrypointFunc func(context.Context, io.Writer, string, map[string]string) (map[string]string, error)
