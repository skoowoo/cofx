package flowl

import (
	"errors"
	"io"
	"path"
	"strings"

	"github.com/cofunclabs/cofunc/internal/functiondriver"
)

func ParseFlowl(rd io.Reader) (runq *RunQueue, ast *AST, err error) {
	if ast, err = ParseAST(rd); err != nil {
		return
	}
	runq, err = NewRunQueue(ast)
	if err != nil {
		return
	}
	return
}

// LoadedLocation
//
type LoadedLocation struct {
	DriverName   string
	FunctionName string
	Path         string
}

// Node
//
type Node struct {
	Name     string
	Driver   functiondriver.FunctionDriver
	Parallel *Node
	Args     map[string]string
	// TODO:
	RunBlock *Block
	FnBlock  *Block
}

// RunQueue
//
type RunQueue struct {
	Locations       map[string]LoadedLocation
	ConfiguredNodes map[string]*Node
	Queue           []*Node
	ast             *AST
}

func NewRunQueue(ast *AST) (*RunQueue, error) {
	q := &RunQueue{
		Locations:       make(map[string]LoadedLocation),
		ConfiguredNodes: make(map[string]*Node),
		Queue:           make([]*Node, 0),
		ast:             ast,
	}
	if err := q.processLoad(ast); err != nil {
		return nil, err
	}
	if err := q.processFn(ast); err != nil {
		return nil, err
	}
	if err := q.processRun(ast); err != nil {
		return nil, err
	}
	return q, nil
}

func (rq *RunQueue) createNode(nodeName, fName string) (*Node, error) {
	loc, ok := rq.Locations[fName]
	if !ok {
		return nil, errors.New("not load function: " + fName)
	}
	loadTarget := loc.DriverName + ":" + loc.Path
	driver := functiondriver.New(loadTarget)
	if driver == nil {
		return nil, errors.New("not found driver: " + loadTarget)
	}
	return &Node{
		Name:   nodeName,
		Driver: driver,
		Args:   make(map[string]string),
	}, nil
}

func (rq *RunQueue) processLoad(ast *AST) error {
	return ast.Foreach(func(b *Block) error {
		if b.Kind.Value != "load" {
			return nil
		}
		s := b.Target.Value
		fields := strings.Split(s, ":")
		dname, p, fname := fields[0], fields[1], path.Base(fields[1])
		if _, ok := rq.Locations[fname]; ok {
			return errors.New("repeat to load function: " + fname)
		}
		rq.Locations[fname] = LoadedLocation{
			DriverName:   dname,
			Path:         p,
			FunctionName: fname,
		}
		return nil
	})
}

func (rq *RunQueue) processFn(ast *AST) error {
	return ast.Foreach(func(b *Block) error {
		if b.Kind.Value != "fn" {
			return nil
		}
		nodeName, fName := b.Target.Value, b.TypeOrValue.Value
		if nodeName == fName {
			return errors.New("node and function name are the same: " + nodeName)
		}
		node, err := rq.createNode(nodeName, fName)
		if err != nil {
			return err
		}
		if _, ok := rq.ConfiguredNodes[node.Name]; ok {
			return errors.New("repeat to configure function:" + node.Name)
		}
		rq.ConfiguredNodes[node.Name] = node
		for _, child := range b.Child {
			if child.Kind.Value == "args" {
				node.Args = child.BlockBody.(*FlMap).ToMap()
			}
		}
		return nil
	})
}

func (rq *RunQueue) processRun(ast *AST) error {
	return ast.Foreach(func(b *Block) error {
		if b.Kind.Value != "run" {
			return nil
		}
		if name := b.Target.Value; name != "" {
			// here is the serial run function
			//
			node, ok := rq.ConfiguredNodes[name]
			if !ok {
				// not configured function, so run directly with default function name
				var err error
				if node, err = rq.createNode(name, name); err != nil {
					return err
				}
				if b.BlockBody != nil {
					node.Args = b.BlockBody.(*FlMap).ToMap()
				}
			} else {
				// the function is configured
				if b.BlockBody != nil {
					node.Args = b.BlockBody.(*FlMap).ToMap()
				}
			}
			rq.Queue = append(rq.Queue, node)
			return nil
		}

		// Here is the parallel run function
		//
		var last *Node
		names := b.BlockBody.(*FlList).ToSlice()
		for _, name := range names {
			node, ok := rq.ConfiguredNodes[name]
			if !ok {
				// not configured function, so run directly with default function name
				var err error
				if node, err = rq.createNode(name, name); err != nil {
					return err
				}
			}
			if last == nil {
				rq.Queue = append(rq.Queue, node)
			} else {
				last.Parallel = node
			}
			last = node
		}
		return nil
	})
}

func (rq *RunQueue) Forstage(do func(int, *Node) error) error {
	for i, e := range rq.Queue {
		if err := do(i+1, e); err != nil {
			return err
		}
	}
	return nil
}

func (rq *RunQueue) Foreach(do func(int, *Node) error) error {
	for i, e := range rq.Queue {
		for p := e; p != nil; p = p.Parallel {
			if err := do(i+1, p); err != nil {
				return err
			}
		}
	}
	return nil
}

func (rq *RunQueue) NodeNum() int {
	var n int
	for _, e := range rq.Queue {
		for p := e; p != nil; p = p.Parallel {
			n += 1
		}
	}
	return n
}
