package flowl

import (
	"container/list"
	"errors"

	"github.com/cofunclabs/cofunc/internal/flowl/driver"
	cmddriver "github.com/cofunclabs/cofunc/internal/flowl/driver/cmd"
	godriver "github.com/cofunclabs/cofunc/internal/flowl/driver/go"
)

// Function
//
type Function struct {
	Name   string
	Driver driver.FunctionDriver
	args   map[string]string

	Parallel *Function
}

func NewFunction(name string, driver driver.FunctionDriver) *Function {
	return &Function{
		Name:   name,
		Driver: driver,
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
		dv driver.FunctionDriver
	)
	if d := godriver.New(location); d != nil {
		dv = d
	} else if d := cmddriver.New(location); d != nil {
		dv = d
	} else {
		return errors.New("not found loader: " + location)
	}

	_, ok := rq.Functions[dv.Name()]
	if !ok {
		rq.Functions[dv.Name()] = NewFunction(dv.Name(), dv)
	} else {
		return errors.New("repeat to load: " + dv.Name())
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
