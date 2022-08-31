package actuator

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrFunctionNotLoaded          error = errors.New("function not loaded")
	ErrLoadedFunctionDuplicated   error = errors.New("loaded function duplicated")
	ErrConfigedFunctionDuplicated error = errors.New("configured function duplicated")
	ErrDriverNotFound             error = errors.New("driver not found")
	ErrNameConflict               error = errors.New("name conflict")
	ErrConditionIsFalse           error = errors.New("condition is false")
	ErrNodeReused                 error = errors.New("node reused")
)

func wrapErrorf(err error, format string, args ...interface{}) error {
	var builder strings.Builder
	builder.WriteString(err.Error())
	builder.WriteString(": ")
	return fmt.Errorf(builder.String()+format, args...)
}
