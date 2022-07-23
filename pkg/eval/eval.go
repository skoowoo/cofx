package eval

import (
	"fmt"

	"github.com/Knetic/govaluate"
)

func New(s string) (interface{}, error) {
	exp, err := govaluate.NewEvaluableExpression(s)
	if err != nil {
		return nil, err
	}
	result, err := exp.Evaluate(nil)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func Float64(s string) (float64, error) {
	res, err := New(s)
	if err != nil {
		return 0, err
	}
	v, ok := res.(float64)
	if !ok {
		return 0, fmt.Errorf("can't convert to float64: '%s'", s)
	}
	return v, nil
}

func Bool(s string) (bool, error) {
	res, err := New(s)
	if err != nil {
		return false, err
	}
	v, ok := res.(bool)
	if !ok {
		return false, fmt.Errorf("can't convert to bool: '%s'", s)
	}
	return v, nil
}

func String(s string) (string, error) {
	res, err := New(s)
	if err != nil {
		return "", err
	}
	switch v := res.(type) {
	case float64:
		return fmt.Sprint(v), nil
	case bool:
		if v {
			return "true", nil
		} else {
			return "false", nil
		}
	default:
		return "", fmt.Errorf("invalid eval type: '%s'", s)
	}
}
