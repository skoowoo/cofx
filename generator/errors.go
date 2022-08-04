package generator

import (
	"errors"
	"fmt"
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
	args = append(args, err)
	return fmt.Errorf(format+": %w", args)
}
