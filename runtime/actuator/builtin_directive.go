package actuator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
)

func lookupBuiltinDirective(name string) func(context.Context, ...string) error {
	f, ok := builtindirectives[name]
	if ok {
		return f
	}
	return nil
}

var builtindirectives = map[string]func(context.Context, ...string) error{
	"sleep":        sleep,
	"println":      _println,
	"exit":         exit,
	"if_none_exit": ifNoneExit,
}

func ifNoneExit(ctx context.Context, args ...string) error {
	if len(args) < 1 {
		return fmt.Errorf("if_none_exit: invalid argument count")
	}
	for i, arg := range args {
		if arg != "" {
			return nil
		} else {
			return fmt.Errorf("none exit: arg is empty at index %d", i)
		}
	}
	return nil
}

func sleep(ctx context.Context, args ...string) error {
	if len(args) != 1 {
		return fmt.Errorf("sleep: invalid argument count")
	}
	s := args[0]
	v, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("%w: sleep", err)
	}
	ticker := time.NewTicker(v)
	select {
	case <-ticker.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func _println(ctx context.Context, args ...string) error {
	for _, arg := range args {
		fmt.Fprintf(os.Stdout, "%s\n", arg)
	}
	return nil
}

func exit(ctx context.Context, args ...string) error {
	if len(args) == 0 {
		return ErrExitWithSuccess
	}
	if len(args) != 1 {
		return fmt.Errorf("exit: invalid argument count")
	}
	e := args[0]
	return errors.New(e)
}
