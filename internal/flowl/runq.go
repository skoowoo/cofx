package flowl

import (
	"container/list"
	"errors"
	"strings"
)

// RunQueue
//
type RunQueue struct {
	Actions map[string]*Action
	Queue   *list.List
}

func NewRunQueue() *RunQueue {
	return nil
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
		loader     ActionLoader
		actionName string
	)
	if l := newGoLoad(location.value); l != nil {
		loader = l
		actionName = l.actionName
	} else if l := newCommandLoad(location.value); l != nil {
		loader = l
		actionName = l.actionName
	} else {
		return errors.New("not found loader: " + location.value)
	}

	action, ok := rq.Actions[actionName]
	if !ok {
		action = &Action{
			Loader: loader,
		}
		rq.Actions[actionName] = action
	}
	if action.Loader != nil {
		return errors.New("repeat to load: " + actionName)
	}
	action.Loader = loader
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
type GoLoad struct {
	location   string
	actionName string
}

func newGoLoad(loc string) *GoLoad {
	if !strings.HasPrefix(loc, "go://") {
		return nil
	}
	name := strings.TrimPrefix(loc, "go://")
	return &GoLoad{
		location:   loc,
		actionName: name,
	}
}

func (l *GoLoad) Load() {

}

// Command
type CommandLoad struct {
	location   string
	actionName string
}

func newCommandLoad(loc string) *CommandLoad {
	// todo
	return &CommandLoad{
		location: loc,
	}
}

func (l *CommandLoad) Load() {

}

// Runner
