package manifest

import (
	"context"
	"io"
	"reflect"
	"runtime"
)

type EntrypointFunc func(context.Context, io.Writer, string, map[string]string) (map[string]string, error)

// Func2Name returns the name of the function 'f', it contains the full package name.
func Func2Name(f EntrypointFunc) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
