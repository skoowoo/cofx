package cofunc

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
)

func GeneratorErrorf(err error, format string, args ...interface{}) error {
	args = append(args, err)
	return fmt.Errorf(format+": %w", args)
}
