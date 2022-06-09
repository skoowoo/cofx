package flowl

import (
	"container/list"
	"errors"
	"path"
	"strings"

	"github.com/cofunclabs/cofunc/internal/gofunctions"
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
	// First directive and it's second token is 'load location'
	location := b.directives[0].tokens[1].text
	var (
		loader FunctionLoader
		runner FunctionRunner
	)
	if l := newGoLoader(location); l != nil {
		loader = l
		runner = newGoRunner()
	} else if l := newCmdLoader(location); l != nil {
		loader = l
		runner = nil //todo
	} else {
		return errors.New("not found loader: " + location)
	}

	_, ok := rq.Functions[loader.Name()]
	if !ok {
		rq.Functions[loader.Name()] = NewFunction(loader.Name(), loader, runner)
	} else {
		return errors.New("repeat to load: " + loader.Name())
	}
	return nil
}

func (rq *RunQueue) processSet(b *Block) error {
	// First directive and it's second token is function's name
	fname := b.directives[0].tokens[1].text
	fc, ok := rq.Functions[fname]
	if !ok {
		return errors.New("in loaded functions, not found: " + fname)
	}
	for _, dir := range b.directives {
		if dir.name == Keyword("input") {
			k, v := dir.tokens[1].text, dir.tokens[2].text
			fc.InputArg(k, v)
		}
		// todo, handle others
	}
	return nil
}

func (rq *RunQueue) processRun(b *Block) error {
	if len(b.directives) == 1 {
		// First directive and it's second token is function's name with prefix '@'
		fname := b.directives[0].tokens[1].subtext[1]
		fc, ok := rq.Functions[fname]
		if !ok {
			return errors.New("in loaded functions, not found: " + fname)
		}
		rq.Queue.PushBack(fc)
		return nil
	}

	// parallel run
	var last *Function
	for _, dir := range b.directives {
		if dir.name != Keyword("@") {
			continue
		}
		fname := dir.tokens[0].subtext[1]
		fc, ok := rq.Functions[fname]
		if !ok {
			return errors.New("in loaded functions, not found: " + fname)
		}
		if last == nil {
			rq.Queue.PushBack(fc) // it's first
		} else {
			last.Parallel = fc
		}
		last = fc
	}
	return nil
}

func (rq *RunQueue) Step(batch func(*Function)) {
	for e := rq.Queue.Front(); e != nil; e = e.Next() {
		batch(e.Value.(*Function))
	}
}

// Function
//

// Function is the unit of task running
type FunctionLoader interface {
	Load() error
	Name() string
}

type FunctionRunner interface {
	Run() error
}

type Function struct {
	Name   string
	Loader FunctionLoader
	Runner FunctionRunner
	args   map[string]string

	Parallel *Function
}

func NewFunction(name string, loader FunctionLoader, runner FunctionRunner) *Function {
	return &Function{
		Name:   name,
		Loader: loader,
		Runner: runner,
		args:   make(map[string]string),
	}
}

func (f *Function) InputArg(k, v string) *Function {
	f.args[k] = v
	return f
}

func (f *Function) Args() map[string]string {
	return f.args
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

func (l *GoLoader) Load() error {
	def := gofunctions.Lookup(l.location)
	if def == nil {
		return errors.New("in gofunctions package, not found function: " + l.location)
	}
	manifest := def.Manifest()
	// todo
	_ = manifest
	return nil
}

func (l *GoLoader) Name() string {
	return l.funcName
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

func (l *CmdLoader) Load() error {
	// todo
	return nil
}

func (l *CmdLoader) Name() string {
	return l.funcName
}

// Runner
type GoRunner struct{}

func newGoRunner() *GoRunner {
	return &GoRunner{}
}

func (r *GoRunner) Run() error {
	return nil
}
