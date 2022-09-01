package actuator

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/cofunclabs/cofunc/functiondriver"
	"github.com/cofunclabs/cofunc/parser"
	"github.com/sirupsen/logrus"
)

// RunQueue
type RunQueue struct {
	locations         functiondriver.LocationStore
	configured        map[string]*TaskNode
	steps             []Node
	triggers          []Trigger
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
		configured: make(map[string]*TaskNode),
		steps:      make([]Node, 0),
		global:     ast.Global(),
	}
	loads, fns, runs := ast.GetBlocks()
	if err := r.generateLocations(loads); err != nil {
		return nil, err
	}
	if err := r.generateConfiguredFn(fns); err != nil {
		return nil, err
	}
	if err := r.generateSteps(runs); err != nil {
		return nil, err
	}
	if err := r.generateEventTriggers(runs); err != nil {
		return nil, err
	}
	return r, nil
}

// GetTriggers returns all event triggers
func (r *RunQueue) GetTriggers() []Trigger {
	return r.triggers
}

// WalkNode traverses all task nodes in order
func (r *RunQueue) WalkNode(do func(Node) error) error {
	for _, e := range r.steps {
		if fe, ok := e.(*TaskNode); ok {
			for p := fe; p != nil; p = p.parallel {
				if err := do(p); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// WalkAndExec is the entry and main program for executing the run queue
func (r *RunQueue) WalkAndExec(ctx context.Context, exec func([]Node) error) error {
	if err := r.beforeExec(ctx); err != nil {
		return err
	}

	var (
		i = 0
	)
	for i < len(r.steps) {
		e := r.steps[i]

		// Execute for node
		if n, ok := e.(*ForNode); ok {
			if err := n.Exec(ctx); err != nil {
				if err == ErrConditionIsFalse {
					i = n.btfIdx + 1
					continue
				}
				return err
			}
		}

		// Execute btf node
		if n, ok := e.(*BtfNode); ok {
			i = n.forIdx
			continue
		}

		// Execute function node
		if n, ok := e.(*TaskNode); ok {
			var batch []Node
			for p := n; p != nil; p = p.parallel {
				batch = append(batch, p)
			}
			if err := exec(batch); err != nil {
				return err
			}
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

func (r *RunQueue) createNode(nodename, fname string) (*TaskNode, error) {
	location, ok := r.locations.Get(fname)
	if !ok {
		return nil, wrapErrorf(ErrFunctionNotLoaded, "'%s'", fname)
	}
	driver := functiondriver.New(location)
	if driver == nil {
		return nil, wrapErrorf(ErrDriverNotFound, "'%s'", location)
	}
	node := &TaskNode{
		name:   nodename,
		driver: driver,
	}
	return node, nil
}

func (r *RunQueue) generateLocations(blocks []*parser.Block) error {
	for _, b := range blocks {
		s := b.Target1().String()
		if l, err := r.locations.Add(s); err != nil {
			return wrapErrorf(ErrLoadedFunctionDuplicated, "'%s' in load list", l.FuncName)
		}
	}
	return nil
}

func (r *RunQueue) generateConfiguredFn(blocks []*parser.Block) error {
	for _, b := range blocks {
		nodename, fname := b.Target1().String(), b.Target2().String()
		node, err := r.createNode(nodename, fname)
		if err != nil {
			return err
		}
		node.fn = b
		r.putConfigured(node)
	}
	return nil
}

func (r *RunQueue) generateEventTriggers(blocks []*parser.Block) error {
	// the seq number of trigger node start from 10000, '< 10000' is reserved for function node.
	seq := 10000
	for _, b := range blocks {
		if b.IsCo() && b.Parent().IsEvent() {
			name := b.Target1().String()
			node := r.getConfigured(name)
			if node == nil {
				// Not configured function, so run directly with default function name
				var err error
				if node, err = r.createNode(name, name); err != nil {
					return wrapErrorf(err, "in event trigger")
				}
			}
			node.co = b
			node.returnVar = b.Target2().String()
			node.seq = seq
			seq++

			r.triggers = append(r.triggers, node)
		}
	}
	return nil
}

func (r *RunQueue) generateSteps(blocks []*parser.Block) error {
	var (
		step = 0
		// the seq number of function node start from 1000, the choice is only for the seq number to have
		// the same length.
		seq = 1000
	)
	for _, b := range blocks {
		// filter out the 'co' blocks in 'event' block
		if b.IsCo() && b.Parent().IsEvent() {
			continue
		}

		if b.IsFor() {
			node := &ForNode{
				idx: len(r.steps), // save the runq's index of 'ForNode'
				b:   b,
			}
			r.processingForNode = node
			r.steps = append(r.steps, node)
			continue
		}
		if b.IsBtf() {
			node := &BtfNode{
				idx:    len(r.steps),
				forIdx: r.processingForNode.idx,
			}
			r.processingForNode.btfIdx = node.idx
			r.steps = append(r.steps, node)
			r.processingForNode = nil
			continue
		}

		var (
			names []string
			last  *TaskNode
		)

		step += 1

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
				if node, err = r.createNode(name, name); err != nil {
					return wrapErrorf(err, " use co to run function")
				}
			}
			node.co = b
			node.returnVar = b.Target2().String()
			node.step = step
			node.seq = seq
			seq += 1

			if last == nil {
				r.steps = append(r.steps, node)
			} else {
				last.parallel = node
			}
			last = node
		}
	}
	return nil
}

func (r *RunQueue) putConfigured(node *TaskNode) {
	r.configured[node.name] = node
}

func (r *RunQueue) getConfigured(nodename string) *TaskNode {
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

type Task interface {
	Step() int
	Seq() int
	Driver() functiondriver.Driver
	IgnoreFailure() bool
	RetryOnFailure() int
}

type Trigger interface {
	Node
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
	if err := n.execCondition(ctx); err != nil {
		return err
	}
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

// TaskNode is the unit of a flow, be used to connect driver and execute the function through the driver
type TaskNode struct {
	// name of the node
	name string
	// driver connected by the node
	driver functiondriver.Driver
	// 'fn' configuration of the function connected by the node
	fn *parser.Block
	// starting definition of the function connected by the node
	co *parser.Block
	// returnVar is a variable name, used to save the function's return values
	returnVar string
	// The execution step that the node will be in, Steps are counted from 1
	step int
	// The sequence number of Node in the run queue
	seq int
	// parallel is a node running in parallel with this node at the same time,
	// We use 'parallel' field to implement the parallel execution of nodes

	_args    *parser.MapBody
	parallel *TaskNode
}

func (n *TaskNode) Step() int {
	return n.step
}

func (n *TaskNode) Seq() int {
	return n.seq
}

func (n *TaskNode) Driver() functiondriver.Driver {
	return n.driver
}

func (n *TaskNode) IgnoreFailure() bool {
	ignore := n.driver.Manifest().IgnoreFailure
	if n.fn != nil {
		if v := n.fn.GetVarValue("ignore_failure"); strings.ToLower(v) == "true" {
			ignore = true
		}
	}
	return ignore
}

func (n *TaskNode) RetryOnFailure() int {
	retries := n.driver.Manifest().RetryOnFailure
	if n.fn != nil {
		if v := n.fn.GetVarValue("retry_on_failure"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				retries = n
			}
		}
	}
	return retries
}

func (n *TaskNode) FormatString() string {
	return fmt.Sprintf("%s->%s", n.name, n.driver.FunctionName())
}

func (n *TaskNode) Name() string {
	return n.name
}

func (n *TaskNode) Init(ctx context.Context, with ...func(context.Context, Node) error) error {
	with = append(with, withArgs())
	for _, f := range with {
		if err := f(ctx, n); err != nil {
			return err
		}
	}
	return nil
}

func (n *TaskNode) Exec(ctx context.Context) error {
	if err := n.execCondition(ctx); err != nil {
		return err
	}
	// exec 'rewrite variable' statement of fn block
	if n.fn != nil {
		for _, stm := range n.fn.List() {
			if err := n.fn.RewriteVar(stm); err != nil {
				return err
			}
		}
	}

	rets, err := n.driver.Run(ctx, n.args())
	if err != nil {
		return err
	}
	if n.needReturns() {
		n.saveReturns(rets, nil)
	}
	return nil
}

func (n *TaskNode) execCondition(ctx context.Context) error {
	if n.co.InSwitch() {
		if !n.co.ExecCondition() {
			return ErrConditionIsFalse
		}
	}
	return nil
}

func (n *TaskNode) args() map[string]string {
	if n._args == nil {
		return map[string]string{}
	}
	return n._args.ToMap()
}

// saveReturns will create some field var
// Field Var are dynamic var
func (n *TaskNode) saveReturns(retkvs map[string]string, filter func(string) bool) bool {
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

func (n *TaskNode) needReturns() bool {
	return len(n.returnVar) != 0
}

func withArgs() func(context.Context, Node) error {
	return func(ctx context.Context, n Node) error {
		funcnode, ok := n.(*TaskNode)
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
}

func WithLoad(logger io.Writer) func(context.Context, Node) error {
	return func(ctx context.Context, n Node) error {
		funcnode, ok := n.(*TaskNode)
		if !ok {
			return nil
		}
		return funcnode.driver.Load(ctx, logger)
	}
}
