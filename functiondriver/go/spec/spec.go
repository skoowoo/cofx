package spec

import (
	"context"
	"errors"
	"io"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

const (
	isString ArgValType = iota
	isInt
	isBool
)

type ArgValType int

// EntrypointArgs is the 'Args' argument of the entrypoint
type EntrypointArgs map[string]string

func (e EntrypointArgs) GetString(name string) string {
	v, _ := e.Get(name, isString)
	if v == nil {
		return ""
	}
	return v.(string)
}

func (e EntrypointArgs) GetInt(name string) (int, error) {
	v, err := e.Get(name, isInt)
	if err != nil {
		return 0, err
	}
	return int(v.(int64)), nil
}

func (e EntrypointArgs) GetBool(name string) (bool, error) {
	v, err := e.Get(name, isBool)
	if err != nil {
		return false, err
	}
	return v.(bool), nil
}

// Get returns the value of the key 'name' in map, if the 'name' not existed, return nil.
func (e EntrypointArgs) Get(name string, typ ArgValType) (interface{}, error) {
	v, ok := e[name]
	if !ok {
		return nil, errors.New("not found")
	}
	switch typ {
	case isString:
		return v, nil
	case isInt:
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return n, nil
		}
	case isBool:
		if s := strings.ToLower(v); s == "true" || s == "yes" || s == "1" {
			return true, nil
		} else if s := strings.ToLower(v); s == "false" || s == "no" || s == "0" {
			return false, nil
		} else {
			return nil, errors.New("invalid bool value")
		}
	}
	return nil, nil
}

// EntrypointBundle
type EntrypointBundle struct {
	Version   string
	Logwriter io.Writer
	Custom    Customer
}

// EntrypointFunc defines the entrypoint type of the function
type EntrypointFunc func(context.Context, EntrypointBundle, EntrypointArgs) (map[string]string, error)

// CreateCustomFunc can be used to create a custom object for the function
// The custom object must implement the 'Close' method, godriver can use this method to
// close or release the custom object.
type CreateCustomFunc func() Customer
type Customer interface {
	Close() error
}

// Func2Name returns the name of the function 'f', it contains the full package name.
func Func2Name(f EntrypointFunc) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
