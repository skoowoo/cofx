package cofunc

import (
	"errors"
	"io"
	"path"
	"strings"

	"github.com/cofunclabs/cofunc/internal/functiondriver"
)

func ParseFlowl(rd io.Reader) (rq *RunQ, ast *AST, err error) {
	if ast, err = ParseAST(rd); err != nil {
		return
	}
	rq, err = NewRunQ(ast)
	if err != nil {
		return
	}
	return
}

// location
//
type location struct {
	dname string
	fname string
	path  string
}

// Node
//
type Node struct {
	name     string
	driver   functiondriver.Driver
	parallel *Node
	rb       *Block
	fb       *Block
}

func (n *Node) String() string {
	return n.name + "->" + n.driver.FunctionName()
}

func (n *Node) Parallel() *Node {
	return n.parallel
}

func (n *Node) setrb(b *Block) {
	n.rb = b
}

func (n *Node) setfb(b *Block) {
	n.fb = b
}

// Args need to be called at running, because it will calcuate variable's value if has variable
func (n *Node) Args() map[string]string {
	var args map[string]string
	if n.rb.bbody != nil {
		args = n.rb.bbody.(*FMap).ToMap()
		return args
	}
	if n.fb != nil {
		for _, b := range n.fb.child {
			if b.Iskind("args") {
				args = b.bbody.(*FMap).ToMap()
				return args
			}
		}
	}
	return nil
}

// RunQ
//
type RunQ struct {
	locations       map[string]location
	configuredNodes map[string]*Node
	stage           []*Node
	ast             *AST
}

func NewRunQ(ast *AST) (*RunQ, error) {
	q := &RunQ{
		locations:       make(map[string]location),
		configuredNodes: make(map[string]*Node),
		stage:           make([]*Node, 0),
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

func (rq *RunQ) createNode(nodename, fname string) (*Node, error) {
	loc, ok := rq.locations[fname]
	if !ok {
		return nil, errors.New("not load function: " + fname)
	}
	l := loc.dname + ":" + loc.path
	driver := functiondriver.New(l)
	if driver == nil {
		return nil, errors.New("not found driver: " + l)
	}
	node := &Node{
		name:   nodename,
		driver: driver,
	}
	return node, nil
}

func (rq *RunQ) processLoad(ast *AST) error {
	return ast.Foreach(func(b *Block) error {
		if b.kind.String() != "load" {
			return nil
		}
		s := b.target.String()
		fields := strings.Split(s, ":")
		dname, p, fname := fields[0], fields[1], path.Base(fields[1])
		if _, ok := rq.locations[fname]; ok {
			return errors.New("repeat to load function: " + fname)
		}
		rq.locations[fname] = location{
			dname: dname,
			path:  p,
			fname: fname,
		}
		return nil
	})
}

func (rq *RunQ) processFn(ast *AST) error {
	return ast.Foreach(func(b *Block) error {
		if b.kind.String() != "fn" {
			return nil
		}
		nodename, fname := b.target.String(), b.typevalue.String()
		if nodename == fname {
			return errors.New("node and function name are the same: " + nodename)
		}
		node, err := rq.createNode(nodename, fname)
		if err != nil {
			return err
		}
		node.setfb(b)
		if _, ok := rq.configuredNodes[node.name]; ok {
			return errors.New("repeat to configure function:" + node.name)
		}
		rq.configuredNodes[node.name] = node
		return nil
	})
}

func (rq *RunQ) processRun(ast *AST) error {
	return ast.Foreach(func(b *Block) error {
		if b.kind.String() != "run" {
			return nil
		}
		if name := b.target.String(); name != "" {
			// here is the serial run function
			//
			node, ok := rq.configuredNodes[name]
			if !ok {
				// not configured function, so run directly with default function name
				var err error
				if node, err = rq.createNode(name, name); err != nil {
					return err
				}
			}
			node.setrb(b)
			rq.stage = append(rq.stage, node)
			return nil
		}

		// Here is the parallel run function
		//
		var last *Node
		names := b.bbody.(*FList).ToSlice()
		for _, name := range names {
			node, ok := rq.configuredNodes[name]
			if !ok {
				// not configured function, so run directly with default function name
				var err error
				if node, err = rq.createNode(name, name); err != nil {
					return err
				}
			}
			node.setrb(b)
			if last == nil {
				rq.stage = append(rq.stage, node)
			} else {
				last.parallel = node
			}
			last = node
		}
		return nil
	})
}

func (rq *RunQ) Forstage(do func(int, *Node) error) error {
	for i, e := range rq.stage {
		if err := do(i+1, e); err != nil {
			return err
		}
	}
	return nil
}

func (rq *RunQ) Foreach(do func(int, *Node) error) error {
	for i, e := range rq.stage {
		for p := e; p != nil; p = p.parallel {
			if err := do(i+1, p); err != nil {
				return err
			}
		}
	}
	return nil
}

func (rq *RunQ) NodeNum() int {
	var n int
	for _, e := range rq.stage {
		for p := e; p != nil; p = p.parallel {
			n += 1
		}
	}
	return n
}
