package cofunc

import (
	"context"
	"fmt"
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
	Init(context.Context, ...func(context.Context, *FuncNode) error) error
	Exec(context.Context) error
}

// ForNode stands for the starting of 'for' loop statement
type ForNode struct {
	idx    int
	btfIdx int
	b      *Block
}

func (n *ForNode) String() string {
	return "for loop"
}

func (n *ForNode) Name() string {
	return "FOR"
}

func (n *ForNode) Init(ctx context.Context, with ...func(context.Context, *FuncNode) error) error {
	return nil
}

func (n *ForNode) Exec(ctx context.Context) error {
	// exec 'rewrite variable' statement of for block
	for _, stm := range n.b.List() {
		if err := n.b.rewriteVar(stm); err != nil {
			return err
		}
	}
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
func (n *BtfNode) Init(ctx context.Context, with ...func(context.Context, *FuncNode) error) error {
	return nil
}

func (n *BtfNode) Exec(ctx context.Context) error {
	return nil
}

// FuncNode
type FuncNode struct {
	name       string
	driver     functiondriver.Driver
	parallel   *FuncNode
	co         *Block
	fn         *Block
	args       *FMap
	retVarName string
}

func (n *FuncNode) String() string {
	return n.name + "->" + n.driver.FunctionName()
}

func (n *FuncNode) Name() string {
	return n.name
}

func withArgs(ctx context.Context, n *FuncNode) error {
	if n.co.bbody != nil {
		m, ok := n.co.bbody.(*FMap)
		if ok {
			n.args = m
			return nil
		}
	}

	if n.fn != nil {
		for _, b := range n.fn.child {
			if b.IsArgs() {
				n.args = b.bbody.(*FMap)
				return nil
			}
		}
	}
	return nil
}

func withLoad(ctx context.Context, n *FuncNode) error {
	return n.driver.Load(ctx)
}

func (n *FuncNode) Init(ctx context.Context, with ...func(context.Context, *FuncNode) error) error {
	if len(with) == 0 {
		with = append(with, withArgs, withLoad)
	}
	for _, f := range with {
		if err := f(ctx, n); err != nil {
			return err
		}
	}
	return nil
}

func (n *FuncNode) Exec(ctx context.Context) error {
	// exec 'rewrite variable' statement of fn block
	if n.fn != nil {
		for _, stm := range n.fn.List() {
			if err := n.fn.rewriteVar(stm); err != nil {
				return err
			}
		}
	}

	if err := n.driver.MergeArgs(n._args()); err != nil {
		return err
	}
	rets, err := n.driver.Run(ctx)
	if err != nil {
		return err
	}
	if n.needSaveReturns() {
		n._saveReturns(rets, nil)
	}
	return nil
}

func (n *FuncNode) _args() map[string]string {
	if n.args == nil {
		return map[string]string{}
	}
	return n.args.ToMap()
}

// _saveReturns will create some field var
// Field Var are dynamic var
func (n *FuncNode) _saveReturns(retkvs map[string]string, filter func(string) bool) bool {
	name := n.retVarName
	_, b := n.co.GetVar(name)
	for field, val := range retkvs {
		if filter != nil && !filter(field) {
			continue
		}
		if err := b.CreateFieldVar(name, field, val); err != nil {
			logrus.Errorln(err)
		}
	}
	return true
}

func (n *FuncNode) needSaveReturns() bool {
	return n.retVarName != ""
}

// RunQ
//
type RunQ struct {
	locations         map[string]location
	configuredNodes   map[string]*FuncNode
	stages            []Node
	g                 *Block
	processingForNode *ForNode
}

func NewRunQ(ast *AST) (*RunQ, error) {
	q := &RunQ{
		locations:       make(map[string]location),
		configuredNodes: make(map[string]*FuncNode),
		stages:          make([]Node, 0),
		g:               &ast.global,
	}
	if err := q.convertLoad(ast); err != nil {
		return nil, err
	}
	if err := q.convertFn(ast); err != nil {
		return nil, err
	}
	if err := q.convertCoAndFor(ast); err != nil {
		return nil, err
	}
	return q, nil
}

func (rq *RunQ) createFuncNode(nodename, fname string) (*FuncNode, error) {
	loc, ok := rq.locations[fname]
	if !ok {
		return nil, GeneratorErrorf(ErrFunctionNotLoaded, "'%s'", fname)
	}
	l := loc.dname + ":" + loc.path
	driver := functiondriver.New(l)
	if driver == nil {
		return nil, GeneratorErrorf(ErrDriverNotFound, "'%s'", l)
	}
	node := &FuncNode{
		name:   nodename,
		driver: driver,
	}
	return node, nil
}

func (rq *RunQ) convertLoad(ast *AST) error {
	return ast.Foreach(func(b *Block) error {
		if !b.IsLoad() {
			return nil
		}
		s := b.target.String()
		fields := strings.Split(s, ":")
		dname, p, fname := fields[0], fields[1], path.Base(fields[1])
		if _, ok := rq.locations[fname]; ok {
			return GeneratorErrorf(ErrLoadedFunctionDuplicated, "'%s' in load list", fname)
		}
		rq.locations[fname] = location{
			dname: dname,
			path:  p,
			fname: fname,
		}
		return nil
	})
}

func (rq *RunQ) convertFn(ast *AST) error {
	return ast.Foreach(func(b *Block) error {
		if !b.IsFn() {
			return nil
		}
		nodename, fname := b.target.String(), b.typevalue.String()
		if nodename == fname {
			return GeneratorErrorf(ErrNameConflict, "node and function name are the same '%s'", nodename)
		}
		node, err := rq.createFuncNode(nodename, fname)
		if err != nil {
			return err
		}
		node.fn = b
		if _, ok := rq.configuredNodes[node.name]; ok {
			return GeneratorErrorf(ErrConfigedFunctionDuplicated, "node name '%s', function name '%s'", node.name, fname)
		}
		rq.configuredNodes[node.name] = node
		return nil
	})
}

func (rq *RunQ) convertCoAndFor(ast *AST) error {
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
				b:   b,
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
					return GeneratorErrorf(err, "in serial run function")
				}
			}
			node.co = b
			node.retVarName = b.typevalue.String()
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
					return GeneratorErrorf(err, "in parallel run function")
				}
			}
			node.co = b
			node.retVarName = b.typevalue.String()
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

func (rq *RunQ) ForfuncNode(do func(int, Node) error) error {
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

// ForstageAndExec is the entry and main program for executing the run queue
func (rq *RunQ) ForstageAndExec(ctx context.Context, exec func(int, []Node) error) (err1 error) {
	defer func() {
		if r := recover(); r != nil {
			err1 = fmt.Errorf("PANIC: %v", r)
		}
	}()

	// exec 'rewrite variable' statement of global
	for _, stm := range rq.g.List() {
		if err := rq.g.rewriteVar(stm); err != nil {
			return err
		}
	}

	stage := 1
	i := 0
	for i < len(rq.stages) {
		e := rq.stages[i]

		if n, ok := e.(*ForNode); ok {
			if err := e.Exec(ctx); err != nil {
				i = n.btfIdx + 1
				continue
			}
		}

		if n, ok := e.(*BtfNode); ok {
			i = n.forIdx
			continue
		}

		if n, ok := e.(*FuncNode); ok {
			var batch []Node
			for p := n; p != nil; p = p.parallel {
				batch = append(batch, p)
			}
			if err := exec(stage, batch); err != nil {
				return err
			}
			stage += 1
		}

		i += 1
	}
	return nil
}
