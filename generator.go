package cofunc

import (
	"errors"
	"io"
	"path"
	"strings"

	"github.com/cofunclabs/cofunc/internal/functiondriver"
	"github.com/sirupsen/logrus"
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
type Node interface {
	String() string
	Name() string
	Exec() error
}

// ForNode stands for the starting of 'for' loop statement
type ForNode struct {
	idx    int
	btfIdx int
}

func (n *ForNode) String() string {
	return "for loop"
}

func (n *ForNode) Name() string {
	return "FOR"
}

func (n *ForNode) Exec() error {
	return nil
}

// btf is an abbreviation for 'back to for'
// BtfNode back to the starting of 'for' statement, start a new cycle
type BtfNode struct {
	idx    int
	forIdx int
}

func (n *BtfNode) String() string {
	return "back to for"
}

func (n *BtfNode) Name() string {
	return "BTF"
}

func (n *BtfNode) Exec() error {
	return nil
}

// FuncNode
type FuncNode struct {
	name     string
	driver   functiondriver.Driver
	parallel *FuncNode
	co       *Block
	fn       *Block
}

func (n *FuncNode) String() string {
	return n.name + "->" + n.driver.FunctionName()
}

func (n *FuncNode) Name() string {
	return n.name
}

func (n *FuncNode) Exec() error {
	return nil
}

func (n *FuncNode) setrb(b *Block) {
	n.co = b
}

func (n *FuncNode) setfb(b *Block) {
	n.fn = b
}

// Args need to be called at running, because it will calcuate variable's value if has variable
func (n *FuncNode) Args() map[string]string {
	var args map[string]string
	if n.co.bbody != nil {
		args = n.co.bbody.(*FMap).ToMap()
		return args
	}
	if n.fn != nil {
		for _, b := range n.fn.child {
			if b.IsArgs() {
				args = b.bbody.(*FMap).ToMap()
				return args
			}
		}
	}
	return nil
}

// SaveReturns need to be called at running, it will create some field var
// Field Var are dynamic var
func (n *FuncNode) SaveReturns(returns map[string]string, filter func(string) bool) bool {
	if n.co.typevalue.IsEmpty() {
		return false
	}
	name := n.co.typevalue.String()
	_, b := n.co.GetVar(name)
	for field, val := range returns {
		if filter != nil && !filter(field) {
			continue
		}
		if err := b.CreateFieldVar(name, field, val); err != nil {
			logrus.Errorln(err)
		}
	}
	return true
}

// RunQ
//
type RunQ struct {
	ast               *AST
	locations         map[string]location
	configuredNodes   map[string]*FuncNode
	stages            []Node
	processingForNode *ForNode
}

func NewRunQ(ast *AST) (*RunQ, error) {
	q := &RunQ{
		locations:       make(map[string]location),
		configuredNodes: make(map[string]*FuncNode),
		stages:          make([]Node, 0),
		ast:             ast,
	}
	if err := q.processLoad(ast); err != nil {
		return nil, err
	}
	if err := q.processFn(ast); err != nil {
		return nil, err
	}
	if err := q.processCoAndFor(ast); err != nil {
		return nil, err
	}
	return q, nil
}

func (rq *RunQ) createFuncNode(nodename, fname string) (*FuncNode, error) {
	loc, ok := rq.locations[fname]
	if !ok {
		return nil, errors.New("not load function: " + fname)
	}
	l := loc.dname + ":" + loc.path
	driver := functiondriver.New(l)
	if driver == nil {
		return nil, errors.New("not found driver: " + l)
	}
	node := &FuncNode{
		name:   nodename,
		driver: driver,
	}
	return node, nil
}

func (rq *RunQ) processLoad(ast *AST) error {
	return ast.Foreach(func(b *Block) error {
		if !b.IsLoad() {
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
		if !b.IsFn() {
			return nil
		}
		nodename, fname := b.target.String(), b.typevalue.String()
		if nodename == fname {
			return errors.New("node and function name are the same: " + nodename)
		}
		node, err := rq.createFuncNode(nodename, fname)
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

func (rq *RunQ) processCoAndFor(ast *AST) error {
	err := ast.Foreach(func(b *Block) error {
		if !b.IsCo() && !b.IsFor() {
			return nil
		}

		// Here is the for statement
		//
		if b.IsFor() {
			if rq.processingForNode != nil {
				// It means that a 'for' loop already exists
				// Before a new 'for' starts, it should mark the end of the previous 'for'
				node := &BtfNode{
					idx:    len(rq.stages),
					forIdx: rq.processingForNode.idx,
				}
				rq.processingForNode.btfIdx = node.idx
				rq.stages = append(rq.stages, node)
			}

			node := &ForNode{
				idx: len(rq.stages), // save the runq's index of 'ForNode'
			}
			rq.stages = append(rq.stages, node)
			rq.processingForNode = node
			return nil
		}

		if rq.processingForNode != nil && !b.parent.IsFor() {
			// It means that a 'for' loop already exists
			// The current 'co' statement is outside the 'for' loop, means that the 'for' loop has ended
			node := &BtfNode{
				idx:    len(rq.stages),
				forIdx: rq.processingForNode.idx,
			}
			rq.processingForNode.btfIdx = node.idx
			rq.stages = append(rq.stages, node)
			rq.processingForNode = nil
		}

		// Here is the serial run function
		//
		if name := b.target.String(); name != "" {
			node, ok := rq.configuredNodes[name]
			if !ok {
				// Not configured function, so run directly with default function name
				var err error
				if node, err = rq.createFuncNode(name, name); err != nil {
					return err
				}
			}
			node.setrb(b)
			rq.stages = append(rq.stages, node)
			return nil
		}

		// Here is the parallel run function
		//
		var last *FuncNode
		names := b.bbody.(*FList).ToSlice()
		for _, name := range names {
			node, ok := rq.configuredNodes[name]
			if !ok {
				// Not configured function, so run directly with default function name
				var err error
				if node, err = rq.createFuncNode(name, name); err != nil {
					return err
				}
			}
			node.setrb(b)
			if last == nil {
				rq.stages = append(rq.stages, node)
			} else {
				last.parallel = node
			}
			last = node
		}
		return nil
	})
	if err != nil {
		return err
	}
	if rq.processingForNode != nil {
		node := &BtfNode{
			idx:    len(rq.stages),
			forIdx: rq.processingForNode.idx,
		}
		rq.processingForNode.btfIdx = node.idx
		rq.stages = append(rq.stages, node)
		rq.processingForNode = nil
	}
	return nil
}

func (rq *RunQ) ForfuncNode(do func(int, *FuncNode) error) error {
	for i, e := range rq.stages {
		if fe, ok := e.(*FuncNode); ok {
			for p := fe; p != nil; p = p.parallel {
				if err := do(i+1, p); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (rq *RunQ) NodeNum() int {
	var n int
	for _, e := range rq.stages {
		if fe, ok := e.(*FuncNode); ok {
			for p := fe; p != nil; p = p.parallel {
				n += 1
			}
		}
	}
	return n
}

// ForstageAndExec is the entry and main program for executing the run queue
func (rq *RunQ) ForstageAndExec(exec func(int, []*FuncNode) error) error {
	i := 0
	for i < len(rq.stages) {
		e := rq.stages[i]

		if n, ok := e.(*ForNode); ok {
			if err := e.Exec(); err != nil {
				i = n.btfIdx + 1
				continue
			}
		}

		if n, ok := e.(*BtfNode); ok {
			i = n.forIdx
			continue
		}

		if n, ok := e.(*FuncNode); ok {
			var batch []*FuncNode
			for p := n; p != nil; p = p.parallel {
				batch = append(batch, p)
			}
			if err := exec(i+1, batch); err != nil {
				return err
			}
		}

		i += 1
	}
	return nil
}
