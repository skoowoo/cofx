package textparse

import "errors"

var (
	ErrNotfound = errors.New("not found")
	ErrTooMany  = errors.New("result is too many")
)
