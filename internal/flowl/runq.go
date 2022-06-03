package flowl

import (
	"container/list"
	"errors"
	"path"
	"strings"
)

// RunQueue
//
type RunQueue struct {
	Functions map[string]*Function
	Queue     *list.List
}

func NewRunQueue() *RunQueue {
	return &RunQueue{
		Functions: make(map[string]*Function),
		Queue:     list.New(),
	}
}

func (rq *RunQueue) Generate(bs *BlockStore) error {
	return bs.Foreach(func(b *Block) error {
		switch b.GetKind() {
		case _block_load:
			if err := rq.processLoad(b); err != nil {
				return err
			}
		case _block_set:
			if err := rq.processSet(b); err != nil {
				return err
			}
		case _block_run:
			if err := rq.processRun(b); err != nil {
				return err
			}
		case _block_var:

		}
		return nil
	})
}

func (rq *RunQueue) processLoad(b *Block) error {
	location := b.directives[0].tokens[1]
	var (
		loader FunctionLoader
		name   string
	)
	if l := newGoLoader(location.value); l != nil {
		loader = l
		name = l.funcName
	} else if l := newCmdLoader(location.value); l != nil {
		loader = l
		name = l.funcName
	} else {
		return errors.New("not found loader: " + location.value)
	}

	_, ok := rq.Functions[name]
	if !ok {
		rq.Functions[name] = &Function{
			Name:   name,
			Loader: loader,
		}
	} else {
		return errors.New("repeat to load: " + name)
	}
	return nil
}

func (rq *RunQueue) processSet(b *Block) error {
	return nil
}

func (rq *RunQueue) processRun(b *Block) error {
	return nil
}

// Function
//

// Function is the unit of task running
type FunctionLoader interface {
	Load()
}

type FunctionRunner interface {
}

type Function struct {
	Name   string
	Loader FunctionLoader
	Runner FunctionRunner
	input  map[string]string
	output map[string]string
}

func NewFunction(name string, loader FunctionLoader, runner FunctionRunner) *Function {
	return &Function{
		Name:   name,
		Loader: loader,
		Runner: runner,
	}
}

func (a *Function) Input(k, v string) *Function {
	a.input[k] = v
	return a
}

func (a *Function) Output() map[string]string {
	return a.output
}

// Loader
//
// go
// load go://function
type GoLoader struct {
	location string
	funcName string
}

func newGoLoader(loc string) *GoLoader {
	if !strings.HasPrefix(loc, "go:") {
		return nil
	}
	name := strings.TrimPrefix(loc, "go:")
	return &GoLoader{
		location: name,
		funcName: name,
	}
}

func (l *GoLoader) Load() {
	// todo
}

// Cmd
type CmdLoader struct {
	location string
	funcName string
}

func newCmdLoader(loc string) *CmdLoader {
	if !strings.HasPrefix(loc, "cmd:") {
		return nil
	}
	p := strings.TrimPrefix(loc, "cmd:")
	name := path.Base(p)
	return &CmdLoader{
		funcName: name,
		location: p,
	}
}

func (l *CmdLoader) Load() {
	// todo
}

// Runner
