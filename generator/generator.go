package generator

import (
	"context"
	"fmt"
	"io"

	"github.com/cofunclabs/cofunc/functiondriver"
	"github.com/cofunclabs/cofunc/parser"
	"github.com/sirupsen/logrus"
)

// RunQueue
//
type RunQueue struct {
	locations         functiondriver.LocationStore
	configured        map[string]*FuncNode
	stages            []Node
	global            *parser.Block
	processingForNode *ForNode
}

func New(rd io.Reader) (*RunQueue, *parser.AST, error) {
	ast, err := parser.New(rd)
	if err != nil {
		return nil, nil, err
	}
	r, err := newRunQueue(ast)
	if err != nil {
		return nil, nil, err
	}
	return r, ast, nil
}

func newRunQueue(ast *parser.AST) (*RunQueue, error) {
	r := &RunQueue{
		locations:  functiondriver.NewLocationStore(),
		configured: make(map[string]*FuncNode),
		stages:     make([]Node, 0),
		global:     ast.Global(),
	}
	loads, fns, runs := ast.GetBlocks()
	if err := r.convertLoad(loads); err != nil {
		return nil, err
	}
	if err := r.convertFn(fns); err != nil {
		return nil, err
	}
	if err := r.convertCoAndFor(runs); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *RunQueue) FuncNodeNum() int {
	var n int
	for _, e := range r.stages {
		if fe, ok := e.(*FuncNode); ok {
			for p := fe; p != nil; p = p.parallel {
				n += 1
			}
		}
	}
	return n
}

func (r *RunQueue) ForfuncNode(do func(int, Node) error) error {
	for i, e := range r.stages {
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
func (r *RunQueue) ForstageAndExec(ctx context.Context, exec func(int, []Node) error) error {
	if err := r.beforeExec(ctx); err != nil {
		return err
	}

	stage := 1
	i := 0
	for i < len(r.stages) {
		e := r.stages[i]

		if n, ok := e.(*ForNode); ok {
			if err := n.execCondition(ctx); err != nil {
				if err == ErrConditionIsFalse {
					i = n.btfIdx + 1
					continue
				}
				return err
			}
			if err := n.Exec(ctx); err != nil {
				return err
			}
		}

		if n, ok := e.(*BtfNode); ok {
			i = n.forIdx
			continue
		}

		if n, ok := e.(*FuncNode); ok {
			var batch []Node
			for p := n; p != nil; p = p.parallel {
				if err := p.execCondition(ctx); err != nil {
					if err == ErrConditionIsFalse {
						continue
					}
					return err
				}
				batch = append(batch, p)
			}
			if err := exec(stage, batch); err != nil {
				return err
			}
			stage += 1
		}

		i += 1
	}

	if err := r.afterExec(ctx); err != nil {
		return err
	}
	return nil
}

func (r *RunQueue) beforeExec(ctx context.Context) error {
	// exec 'rewrite variable' statement of global
	for _, stm := range r.global.List() {
		if err := r.global.RewriteVar(stm); err != nil {
			return err
		}
	}
	return nil
}

func (r *RunQueue) afterExec(ctx context.Context) error {
	return nil
}

func (r *RunQueue) createFuncNode(nodename, fname string) (*FuncNode, error) {
	location, ok := r.locations.Get(fname)
	if !ok {
		return nil, wrapErrorf(ErrFunctionNotLoaded, "'%s'", fname)
	}
	driver := functiondriver.New(location)
	if driver == nil {
		return nil, wrapErrorf(ErrDriverNotFound, "'%s'", location)
	}
	node := &FuncNode{
		name:   nodename,
		driver: driver,
	}
	return node, nil
}

func (r *RunQueue) convertLoad(blocks []*parser.Block) error {
	for _, b := range blocks {
		s := b.Target1().String()
		if l, err := r.locations.Add(s); err != nil {
			return wrapErrorf(ErrLoadedFunctionDuplicated, "'%s' in load list", l.FuncName)
		}
	}
	return nil
}

func (r *RunQueue) convertFn(blocks []*parser.Block) error {
	for _, b := range blocks {
		nodename, fname := b.Target1().String(), b.Target2().String()
		node, err := r.createFuncNode(nodename, fname)
		if err != nil {
			return err
		}
		node.fn = b
		r.putConfigured(node)
	}
	return nil
}

func (r *RunQueue) convertCoAndFor(blocks []*parser.Block) error {
	for _, b := range blocks {
		if b.IsFor() {
			node := &ForNode{
				idx: len(r.stages), // save the runq's index of 'ForNode'
				b:   b,
			}
			r.processingForNode = node
			r.stages = append(r.stages, node)
			continue
		}
		if b.IsBtf() {
			node := &BtfNode{
				idx:    len(r.stages),
				forIdx: r.processingForNode.idx,
			}
			r.processingForNode.btfIdx = node.idx
			r.stages = append(r.stages, node)
			r.processingForNode = nil
			continue
		}

		var (
			names []string
			last  *FuncNode
		)
		if !b.Target1().IsEmpty() {
			names = append(names, b.Target1().String()) // only one
		} else {
			names = b.Body().(*parser.ListBody).ToSlice()
		}
		for _, name := range names {
			node := r.getConfigured(name)
			if node == nil {
				// Not configured function, so run directly with default function name
				var err error
				if node, err = r.createFuncNode(name, name); err != nil {
					return wrapErrorf(err, "in parallel run function")
				}
			}
			node.co = b
			node.returnVar = b.Target2().String()
			if last == nil {
				r.stages = append(r.stages, node)
			} else {
				last.parallel = node
			}
			last = node
		}
	}
	return nil
}

func (r *RunQueue) putConfigured(node *FuncNode) {
	r.configured[node.name] = node
}

func (r *RunQueue) getConfigured(nodename string) *FuncNode {
	node, ok := r.configured[nodename]
	if !ok {
		return nil
	}
	return node
}

// Node
//
type Node interface {
	FormatString() string
	Name() string
	Init(context.Context, ...func(context.Context, Node) error) error
	Exec(context.Context) error
}

// ForNode stands for the starting of 'for' loop statement
type ForNode struct {
	idx    int
	btfIdx int
	b      *parser.Block
}

func (n *ForNode) FormatString() string {
	return fmt.Sprintf("for: %d,%d", n.idx, n.btfIdx)
}

func (n *ForNode) Name() string {
	return "FOR"
}

func (n *ForNode) Init(ctx context.Context, with ...func(context.Context, Node) error) error {
	return nil
}

func (n *ForNode) Exec(ctx context.Context) error {
	// exec 'rewrite variable' statement of for block
	for _, stm := range n.b.List() {
		if err := n.b.RewriteVar(stm); err != nil {
			return err
		}
	}
	return nil
}

func (n *ForNode) execCondition(ctx context.Context) error {
	// exec 'for condition' expression
	if !n.b.ExecCondition() {
		return ErrConditionIsFalse
	}
	return nil
}

// btf is an abbreviation for 'back to for'
// BtfNode back to the starting of 'for' statement, start a new cycle
type BtfNode struct {
	idx    int
	forIdx int
}

func (n *BtfNode) FormatString() string {
	return fmt.Sprintf("btf: %d,%d", n.idx, n.forIdx)
}

func (n *BtfNode) Name() string {
	return "BTF"
}
func (n *BtfNode) Init(ctx context.Context, with ...func(context.Context, Node) error) error {
	return nil
}

func (n *BtfNode) Exec(ctx context.Context) error {
	return nil
}

// FuncNode
type FuncNode struct {
	name      string
	driver    functiondriver.Driver
	parallel  *FuncNode
	fn        *parser.Block
	co        *parser.Block
	_args     *parser.MapBody
	returnVar string
}

func (n *FuncNode) FormatString() string {
	return fmt.Sprintf("%s->%s", n.name, n.driver.FunctionName())
}

func (n *FuncNode) Name() string {
	return n.name
}

func (n *FuncNode) Init(ctx context.Context, with ...func(context.Context, Node) error) error {
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
			if err := n.fn.RewriteVar(stm); err != nil {
				return err
			}
		}
	}

	rets, err := n.driver.Run(ctx, n.driver.MergeArgs(n.args()))
	if err != nil {
		return err
	}
	if n.needReturns() {
		n.saveReturns(rets, nil)
	}
	return nil
}

func (n *FuncNode) execCondition(ctx context.Context) error {
	if n.co.InSwitch() {
		if !n.co.ExecCondition() {
			return ErrConditionIsFalse
		}
	}
	return nil
}

func (n *FuncNode) args() map[string]string {
	if n._args == nil {
		return map[string]string{}
	}
	return n._args.ToMap()
}

// saveReturns will create some field var
// Field Var are dynamic var
func (n *FuncNode) saveReturns(retkvs map[string]string, filter func(string) bool) bool {
	name := n.returnVar
	for field, val := range retkvs {
		if filter != nil && !filter(field) {
			continue
		}
		if err := n.co.AddField2Var(name, field, val); err != nil {
			logrus.Errorln(err)
		}
	}
	return true
}

func (n *FuncNode) needReturns() bool {
	return len(n.returnVar) != 0
}

func withArgs(ctx context.Context, n Node) error {
	funcnode, ok := n.(*FuncNode)
	if !ok {
		return nil
	}
	if funcnode.co.Body() != nil {
		m, ok := funcnode.co.Body().(*parser.MapBody)
		if ok {
			funcnode._args = m
			return nil
		}
	}

	if funcnode.fn != nil {
		for _, b := range funcnode.fn.Child() {
			if b.IsArgs() {
				funcnode._args = b.Body().(*parser.MapBody)
				return nil
			}
		}
	}
	return nil
}

func withLoad(ctx context.Context, n Node) error {
	funcnode, ok := n.(*FuncNode)
	if !ok {
		return nil
	}
	return funcnode.driver.Load(ctx)
}
