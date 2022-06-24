package functiondriver

import (
	"context"

	cmddriver "github.com/cofunclabs/cofunc/internal/functiondriver/cmd"
	godriver "github.com/cofunclabs/cofunc/internal/functiondriver/go"
)

type Driver interface {
	FunctionName() string
	Load(ctx context.Context, args map[string]string) error
	Run(ctx context.Context) (map[string]string, error)
}

func New(l string) Driver {
	if d := godriver.New(l); d != nil {
		return d
	}
	if d := cmddriver.New(l); d != nil {
		return d
	}
	return nil
}
