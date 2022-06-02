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
	Actions map[string]*Action
	Queue   *list.List
}

func NewRunQueue() *RunQueue {
	return &RunQueue{
		Actions: make(map[string]*Action),
		Queue:   list.New(),
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
		loader ActionLoader
		name   string
	)
	if l := newGoLoader(location.value); l != nil {
		loader = l
		name = l.actionName
	} else if l := newCmdLoader(location.value); l != nil {
		loader = l
		name = l.actionName
	} else {
		return errors.New("not found loader: " + location.value)
	}

	_, ok := rq.Actions[name]
	if !ok {
		rq.Actions[name] = &Action{
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

// Action
//

// Action is the unit of task running
type ActionLoader interface {
	Load()
}

type ActionRunner interface {
}

type Action struct {
	Name   string
	Loader ActionLoader
	Runner ActionRunner
	input  map[string]string
	output map[string]string
}

func NewAction(name string, loader ActionLoader, runner ActionRunner) *Action {
	return &Action{
		Name:   name,
		Loader: loader,
		Runner: runner,
	}
}

func (a *Action) Input(k, v string) *Action {
	a.input[k] = v
	return a
}

func (a *Action) Output() map[string]string {
	return a.output
}

// Loader
//
// go
// load go://action
type GoLoader struct {
	location   string
	actionName string
}

func newGoLoader(loc string) *GoLoader {
	if !strings.HasPrefix(loc, "go:") {
		return nil
	}
	name := strings.TrimPrefix(loc, "go:")
	return &GoLoader{
		location:   name,
		actionName: name,
	}
}

func (l *GoLoader) Load() {

}

// Cmd
type CmdLoader struct {
	location   string
	actionName string
}

func newCmdLoader(loc string) *CmdLoader {
	if !strings.HasPrefix(loc, "cmd:") {
		return nil
	}
	p := strings.TrimPrefix(loc, "cmd:")
	name := path.Base(p)
	return &CmdLoader{
		actionName: name,
		location:   p,
	}
}

func (l *CmdLoader) Load() {

}

// Runner
