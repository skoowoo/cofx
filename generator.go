package cofunc

import (
	"context"
	"io"
	"path"
	"strings"

	"github.com/cofunclabs/cofunc/internal/functiondriver"
	"github.com/sirupsen/logrus"
)

func ParseFlowl(rd io.Reader) (*RunQ, *AST, error) {
	ast, err := ParseAST(rd)
	if err != nil {
		return nil, nil, err
	}
	r, err := NewRunQ(ast)
	if err != nil {
		return nil, nil, err
	}
	return r, ast, nil
}

// location
//
type location struct {
	dname string
	fname string
	path  string
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
	r := &RunQ{
		locations:       make(map[string]location),
		configuredNodes: make(map[string]*FuncNode),
		stages:          make([]Node, 0),
		g:               &ast.global,
	}
	if err := r.convertLoad(ast); err != nil {
		return nil, err
	}
	if err := r.convertFn(ast); err != nil {
		return nil, err
	}
	if err := r.convertCoAndFor(ast); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *RunQ) FuncNodeNum() int {
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

func (r *RunQ) ForfuncNode(do func(int, Node) error) error {
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
func (r *RunQ) ForstageAndExec(ctx context.Context, exec func(int, []Node) error) error {
	if err := r.beforeExec(ctx); err != nil {
		return err
	}

	stage := 1
	i := 0
	for i < len(r.stages) {
		e := r.stages[i]

		if n, ok := e.(*ForNode); ok {
			if err := n.ConditionExec(ctx); err != nil {
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
				if err := p.ConditionExec(ctx); err != nil {
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

func (r *RunQ) beforeExec(ctx context.Context) error {
	// exec 'rewrite variable' statement of global
	for _, stm := range r.g.List() {
		if err := r.g.rewriteVar(stm); err != nil {
			return err
		}
	}
	return nil
}

func (r *RunQ) afterExec(ctx context.Context) error {
	return nil
}

func (r *RunQ) createFuncNode(nodename, fname string) (*FuncNode, error) {
	loc, ok := r.locations[fname]
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

func (r *RunQ) convertLoad(ast *AST) error {
	return ast.Foreach(func(b *Block) error {
		if !b.IsLoad() {
			return nil
		}
		s := b.target1.String()
		fields := strings.Split(s, ":")
		dname, p, fname := fields[0], fields[1], path.Base(fields[1])
		if _, ok := r.locations[fname]; ok {
			return GeneratorErrorf(ErrLoadedFunctionDuplicated, "'%s' in load list", fname)
		}
		r.locations[fname] = location{
			dname: dname,
			path:  p,
			fname: fname,
		}
		return nil
	})
}

func (r *RunQ) convertFn(ast *AST) error {
	return ast.Foreach(func(b *Block) error {
		if !b.IsFn() {
			return nil
		}
		nodename, fname := b.target1.String(), b.target2.String()
		if nodename == fname {
			return GeneratorErrorf(ErrNameConflict, "node and function name are the same '%s'", nodename)
		}
		node, err := r.createFuncNode(nodename, fname)
		if err != nil {
			return err
		}
		node.fn = b
		if _, ok := r.configuredNodes[node.name]; ok {
			return GeneratorErrorf(ErrConfigedFunctionDuplicated, "node name '%s', function name '%s'", node.name, fname)
		}
		r.configuredNodes[node.name] = node
		return nil
	})
}

func (r *RunQ) convertCoAndFor(ast *AST) error {
	err := ast.Foreach(func(b *Block) error {
		if !b.IsCo() && !b.IsFor() {
			return nil
		}

		// Here is the for statement
		//
		if b.IsFor() {
			if r.processingForNode != nil {
				// It means that a 'for' loop already exists
				// Before a new 'for' starts, it should mark the end of the previous 'for'
				node := &BtfNode{
					idx:    len(r.stages),
					forIdx: r.processingForNode.idx,
				}
				r.processingForNode.btfIdx = node.idx
				r.stages = append(r.stages, node)
			}

			node := &ForNode{
				idx: len(r.stages), // save the runq's index of 'ForNode'
				b:   b,
			}
			r.stages = append(r.stages, node)
			r.processingForNode = node

			if !b.target1.IsEmpty() && !b.target2.IsEmpty() {
				stm := newstm("var").Append(&b.target1).Append(&b.target2)
				node.condition = stm
				if err := b.initVar(stm); err != nil {
					return err
				}
			}
			return nil
		}

		if r.processingForNode != nil && !b.InFor() {
			// It means that a 'for' loop already exists
			// The current 'co' statement is outside the 'for' loop, means that the 'for' loop has ended
			node := &BtfNode{
				idx:    len(r.stages),
				forIdx: r.processingForNode.idx,
			}
			r.processingForNode.btfIdx = node.idx
			r.stages = append(r.stages, node)
			r.processingForNode = nil
		}

		if b.IsCo() && (b.parent.IsCase() || b.parent.IsDefault()) {
			stm := newstm("var").Append(&b.parent.target1).Append(&b.parent.target2)
			if err := b.initVar(stm); err != nil {
				return err
			}
		}

		// Here is the serial run function
		//
		if name := b.target1.String(); name != "" {
			node, ok := r.configuredNodes[name]
			if !ok {
				// Not configured function, so run directly with default function name
				var err error
				if node, err = r.createFuncNode(name, name); err != nil {
					return GeneratorErrorf(err, "in serial run function")
				}
			}
			node.co = b
			node.retVarName = b.target2.String()
			r.stages = append(r.stages, node)
			return nil
		}

		// Here is the parallel run function
		//
		var last *FuncNode
		names := b.bbody.(*FList).ToSlice()
		for _, name := range names {
			node, ok := r.configuredNodes[name]
			if !ok {
				// Not configured function, so run directly with default function name
				var err error
				if node, err = r.createFuncNode(name, name); err != nil {
					return GeneratorErrorf(err, "in parallel run function")
				}
			}
			node.co = b
			node.retVarName = b.target2.String()
			if last == nil {
				r.stages = append(r.stages, node)
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
	if r.processingForNode != nil {
		node := &BtfNode{
			idx:    len(r.stages),
			forIdx: r.processingForNode.idx,
		}
		r.processingForNode.btfIdx = node.idx
		r.stages = append(r.stages, node)
		r.processingForNode = nil
	}
	return nil
}

// Node
//
type Node interface {
	String() string
	Name() string
	Init(context.Context, ...func(context.Context, Node) error) error
	ConditionExec(context.Context) error
	Exec(context.Context) error
}

// ForNode stands for the starting of 'for' loop statement
type ForNode struct {
	idx       int
	btfIdx    int
	b         *Block
	condition *Statement
}

func (n *ForNode) String() string {
	return "for loop"
}

func (n *ForNode) Name() string {
	return "FOR"
}

func (n *ForNode) Init(ctx context.Context, with ...func(context.Context, Node) error) error {
	return nil
}

func (n *ForNode) ConditionExec(ctx context.Context) error {
	// exec 'for condition' expression
	if n.condition != nil {
		if !n.b.CalcConditionTrue() {
			return ErrConditionIsFalse
		}
	}
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
func (n *BtfNode) Init(ctx context.Context, with ...func(context.Context, Node) error) error {
	return nil
}

func (n *BtfNode) ConditionExec(ctx context.Context) error {
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

func (n *FuncNode) ConditionExec(ctx context.Context) error {
	if n.co.InSwitch() {
		if !n.co.CalcConditionTrue() {
			return ErrConditionIsFalse
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
		if err := b.addField2Var(name, field, val); err != nil {
			logrus.Errorln(err)
		}
	}
	return true
}

func (n *FuncNode) needSaveReturns() bool {
	return n.retVarName != ""
}

func withArgs(ctx context.Context, n Node) error {
	funcnode, ok := n.(*FuncNode)
	if !ok {
		return nil
	}
	if funcnode.co.bbody != nil {
		m, ok := funcnode.co.bbody.(*FMap)
		if ok {
			funcnode.args = m
			return nil
		}
	}

	if funcnode.fn != nil {
		for _, b := range funcnode.fn.child {
			if b.IsArgs() {
				funcnode.args = b.bbody.(*FMap)
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
