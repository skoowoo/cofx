package functiondriver

import (
	cmddriver "github.com/cofunclabs/cofunc/internal/functiondriver/cmd"
	godriver "github.com/cofunclabs/cofunc/internal/functiondriver/go"
)

type FunctionDriver interface {
	Name() string
	Load(args map[string]string) error
	Run() (map[string]string, error)
}

func New(location string) FunctionDriver {
	if d := godriver.New(location); d != nil {
		return d
	}
	if d := cmddriver.New(location); d != nil {
		return d
	}
	return nil
}
