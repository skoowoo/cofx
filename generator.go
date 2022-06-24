package cofunc

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
	driverName   string
	functionName string
	path         string
}

// Node
//
type Node struct {
	Name     string
	Driver   functiondriver.FunctionDriver
	Parallel *Node
	args     map[string]string
	// TODO:
	runBlock *Block
	fnBlock  *Block
}

// RunQueue
//
type RunQueue struct {
	locations       map[string]LoadedLocation
	configuredNodes map[string]*Node
	queue           []*Node
	ast             *AST
}

func NewRunQueue(ast *AST) (*RunQueue, error) {
	q := &RunQueue{
		locations:       make(map[string]LoadedLocation),
		configuredNodes: make(map[string]*Node),
		queue:           make([]*Node, 0),
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
	loc, ok := rq.locations[fName]
	if !ok {
		return nil, errors.New("not load function: " + fName)
	}
	loadTarget := loc.driverName + ":" + loc.path
	driver := functiondriver.New(loadTarget)
	if driver == nil {
		return nil, errors.New("not found driver: " + loadTarget)
	}
	return &Node{
		Name:   nodeName,
		Driver: driver,
		args:   make(map[string]string),
	}, nil
}

func (rq *RunQueue) processLoad(ast *AST) error {
	return ast.Foreach(func(b *Block) error {
		if b.kind.value != "load" {
			return nil
		}
		s := b.target.value
		fields := strings.Split(s, ":")
		dname, p, fname := fields[0], fields[1], path.Base(fields[1])
		if _, ok := rq.locations[fname]; ok {
			return errors.New("repeat to load function: " + fname)
		}
		rq.locations[fname] = LoadedLocation{
			driverName:   dname,
			path:         p,
			functionName: fname,
		}
		return nil
	})
}

func (rq *RunQueue) processFn(ast *AST) error {
	return ast.Foreach(func(b *Block) error {
		if b.kind.value != "fn" {
			return nil
		}
		nodeName, fName := b.target.value, b.typevalue.value
		if nodeName == fName {
			return errors.New("node and function name are the same: " + nodeName)
		}
		node, err := rq.createNode(nodeName, fName)
		if err != nil {
			return err
		}
		if _, ok := rq.configuredNodes[node.Name]; ok {
			return errors.New("repeat to configure function:" + node.Name)
		}
		rq.configuredNodes[node.Name] = node
		for _, child := range b.child {
			if child.kind.value == "args" {
				node.args = child.BlockBody.(*FMap).ToMap()
			}
		}
		return nil
	})
}

func (rq *RunQueue) processRun(ast *AST) error {
	return ast.Foreach(func(b *Block) error {
		if b.kind.value != "run" {
			return nil
		}
		if name := b.target.value; name != "" {
			// here is the serial run function
			//
			node, ok := rq.configuredNodes[name]
			if !ok {
				// not configured function, so run directly with default function name
				var err error
				if node, err = rq.createNode(name, name); err != nil {
					return err
				}
				if b.BlockBody != nil {
					node.args = b.BlockBody.(*FMap).ToMap()
				}
			} else {
				// the function is configured
				if b.BlockBody != nil {
					node.args = b.BlockBody.(*FMap).ToMap()
				}
			}
			rq.queue = append(rq.queue, node)
			return nil
		}

		// Here is the parallel run function
		//
		var last *Node
		names := b.BlockBody.(*FList).ToSlice()
		for _, name := range names {
			node, ok := rq.configuredNodes[name]
			if !ok {
				// not configured function, so run directly with default function name
				var err error
				if node, err = rq.createNode(name, name); err != nil {
					return err
				}
			}
			if last == nil {
				rq.queue = append(rq.queue, node)
			} else {
				last.Parallel = node
			}
			last = node
		}
		return nil
	})
}

func (rq *RunQueue) Forstage(do func(int, *Node) error) error {
	for i, e := range rq.queue {
		if err := do(i+1, e); err != nil {
			return err
		}
	}
	return nil
}

func (rq *RunQueue) Foreach(do func(int, *Node) error) error {
	for i, e := range rq.queue {
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
	for _, e := range rq.queue {
		for p := e; p != nil; p = p.Parallel {
			n += 1
		}
	}
	return n
}
