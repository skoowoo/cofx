package parser

import (
	"container/list"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/cofunclabs/cofunc/pkg/enabled"
	"github.com/cofunclabs/cofunc/pkg/eval"
	"github.com/cofunclabs/cofunc/pkg/is"
)

const (
	_condition_expr_var = "x0f1f2f3__"
)

type _var struct {
	sync.Mutex
	v        string
	segments []struct {
		str   string
		isvar bool
	}
	child  []*_var
	cached bool
	asexp  bool
	fields map[string]string

	// for $(v.key)
	field string
	mainv *_var

	// for env
	isenv bool
}

func (v *_var) update(nv *_var) {
	v.Lock()
	defer v.Unlock()

	v.v = nv.v
	v.segments = nv.segments
	v.child = nv.child
	v.cached = nv.cached
	v.asexp = nv.asexp
}

func (v *_var) calc() (string, bool) {
	v.Lock()
	defer v.Unlock()

	if v.mainv != nil && v.field != "" {
		if v.mainv.isenv {
			return os.Getenv(v.field), true
		} else {
			return v.mainv.fields[v.field], false
		}
	}

	if v.cached && !v.asexp {
		return v.v, v.cached
	}

	var (
		vals      []string
		cacheable = true
		vb        strings.Builder
	)
	for _, c := range v.child {
		val, cached := c.calc()
		vals = append(vals, val)
		if !cached {
			cacheable = false
		}
	}
	var seq int
	for _, seg := range v.segments {
		if seg.isvar {
			seg.str = vals[seq]
			seq += 1
		}
		vb.WriteString(seg.str)
	}

	if v.asexp {
		s := vb.String()
		res, err := eval.String(s)
		if err != nil {
			panic(fmt.Errorf("%w: '%s' '%p'", err, s, v))
		}
		v.v = res
		if len(v.child) == 0 {
			v.cached = true
		}
		return v.v, v.cached
	}

	v.v = vb.String()
	if cacheable {
		v.cached = true
	}
	return v.v, v.cached
}

func (v *_var) dfscycle(stack *list.List) error {
	for e := stack.Front(); e != nil; e = e.Next() {
		if e.Value.(*_var) == v {
			// has a cycle
			return ErrVariableHasCycle
		}
	}
	stack.PushBack(v)

	v.Lock()
	defer v.Unlock()

	if len(v.child) == 0 {
		stack.Remove(stack.Back())
		return nil
	}
	for _, c := range v.child {
		if err := c.dfscycle(stack); err != nil {
			return err
		}
	}

	stack.Remove(stack.Back())
	return nil
}

func (v *_var) addField(key, val string) {
	v.Lock()
	defer v.Unlock()
	if v.fields == nil {
		v.fields = make(map[string]string)
	}
	v.fields[key] = val
}

func (v *_var) readField(f string) string {
	v.Lock()
	defer v.Unlock()
	return v.fields[f]
}

func newVarFromToken(t *Token) (*_var, error) {
	v := &_var{
		v:        t.String(),
		segments: t._segments,
		asexp:    t.TypeEqual(_expr_t),
	}
	if !t.hasVar() {
		v.cached = true
	}
	for _, seg := range v.segments {
		if !seg.isvar {
			continue
		}
		var chld *_var
		name := seg.str
		main, field, ok := isFieldVar(name)
		if ok {
			mv, _ := t._b.getVar(main)
			if mv == nil {
				return nil, tokenErrorf(t.ln, ErrVariableNotDefined, "'%s', variable name '%s'", t, main)
			}
			chld = &_var{
				field: field,
				mainv: mv,
			}
		} else {
			chld, _ = t._b.getVar(name)
		}

		if chld != nil {
			v.child = append(v.child, chld)
		} else {
			return nil, tokenErrorf(t.ln, ErrVariableNotDefined, "'%s', variable name '%s'", t, name)
		}
	}
	return v, nil
}

func newVarFromStm(stm *Statement) (*_var, error) {
	var (
		v   *_var
		err error
	)
	if len(stm.tokens) == 2 {
		vt := stm.tokens[1]
		if v, err = newVarFromToken(vt); err != nil {
			return nil, err
		}
	} else {
		v = &_var{
			segments: []struct {
				str   string
				isvar bool
			}{},
			child: []*_var{},
		}
	}
	return v, nil
}

func newEnvVar() *_var {
	return &_var{
		isenv: true,
	}
}

func isFieldVar(name string) (string, string, bool) {
	fields := strings.SplitN(name, ".", 2)
	if len(fields) != 2 {
		return "", "", false
	}
	return fields[0], fields[1], true
}

type expression struct {
	s string
}

func newExpression(tokens []*Token) *expression {
	var (
		hasString bool
		hasArith  bool
		hasNumber bool
		builder   strings.Builder
		subtokens []*Token
	)

	convert := func() {
		for _, t := range subtokens {
			switch t.typ {
			case _string_t:
				builder.WriteString("\"")
				builder.WriteString(t.String())
				builder.WriteString("\"")
			case _refvar_t:
				if hasString {
					builder.WriteString("\"")
					builder.WriteString(t.String())
					builder.WriteString("\"")
				} else if hasNumber {
					builder.WriteString(t.String())
				} else if !hasArith {
					builder.WriteString("\"")
					builder.WriteString(t.String())
					builder.WriteString("\"")
				} else {
					builder.WriteString(t.String())
				}
			default:
				builder.WriteString(t.String())
			}
		}

	}

	for _, t := range tokens {
		subtokens = append(subtokens, t)

		if t.String() == "||" || t.String() == "&&" {
			convert()
			hasArith = false
			hasString = false
			subtokens = nil
			continue
		}

		if is.Arithmetic(t.String()) {
			hasArith = true
		}
		if t.TypeEqual(_string_t) {
			hasString = true
		}
		if t.TypeEqual(_number_t) {
			hasNumber = true
		}
	}
	if subtokens != nil {
		convert()
	}

	return &expression{
		s: builder.String(),
	}
}

func (e *expression) ToToken() *Token {
	return &Token{
		str: e.s,
		typ: _expr_t,
	}
}

// vartable defined var table for each block
type vartable struct {
	sync.Mutex
	vars map[string]*_var
}

func (vs *vartable) debug(tab ...string) {
	if !enabled.Debug() {
		return
	}

	vs.Lock()
	defer vs.Unlock()

	indent := strings.Join(tab, "")
	fmt.Println(indent + "variables in block:")
	for k, v := range vs.vars {
		fmt.Printf(indent+"\tname:'%s', value:'%s', exp:'%t', addr:%p, segments:'%+v'\n", k, v.v, v.asexp, v, v.segments)
		for _, c := range v.child {
			fmt.Printf(indent+"\t\taddr:'%p', value:'%s', exp:'%t'\n", c, c.v, c.asexp)
		}
	}
}

func (vs *vartable) put(name string, v *_var) {
	vs.Lock()
	defer vs.Unlock()

	old, ok := vs.vars[name]
	if ok {
		old.update(v)
	} else {
		vs.vars[name] = v
	}
}

func (vs *vartable) add(name string, v *_var) error {
	vs.Lock()
	defer vs.Unlock()

	_, ok := vs.vars[name]
	if ok {
		return fmt.Errorf("'%s': %w", name, ErrVariableNameDuplicated)
	}
	vs.vars[name] = v
	return nil
}

func (vs *vartable) get(name string) (*_var, bool) {
	vs.Lock()
	defer vs.Unlock()

	v, ok := vs.vars[name]
	return v, ok
}

func (vs *vartable) calc(name string) (_v interface{}, cached bool) {
	main, field, ok := isFieldVar(name)
	if ok {
		if main == "env" {
			return os.Getenv(field), true
		}
		v, ok := vs.get(main)
		if !ok {
			return nil, false
		}
		return v.readField(field), false
	}

	v, ok := vs.get(name)
	if !ok {
		return nil, false
	}
	return v.calc()
}

func (vs *vartable) cyclecheck(names ...string) error {
	vs.Lock()
	defer vs.Unlock()

	stack := list.New()

	if len(names) != 0 {
		for _, name := range names {
			v, ok := vs.vars[name]
			if ok {
				if err := v.dfscycle(stack); err != nil {
					return fmt.Errorf("%w: start variable '%s'", err, name)
				}
			}
		}
		return nil
	}

	for name, v := range vs.vars {
		if err := v.dfscycle(stack); err != nil {
			return fmt.Errorf("%w: start variable '%s'", err, name)
		}
	}
	return nil
}
