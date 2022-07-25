package cofunc

import (
	"container/list"
	"fmt"
	"strings"
	"sync"

	"github.com/cofunclabs/cofunc/pkg/enabled"
	"github.com/cofunclabs/cofunc/pkg/eval"
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
}

func (v *_var) updateval(nv *_var) {
	v.Lock()
	defer v.Unlock()

	v.v = nv.v
	v.segments = nv.segments
	v.child = nv.child
	v.cached = nv.cached
	v.asexp = nv.asexp
}

func (v *_var) calcvarval() (string, bool) {
	v.Lock()
	defer v.Unlock()

	if v.cached && !v.asexp {
		return v.v, v.cached
	}
	var (
		vals      []string
		cacheable = true
		vb        strings.Builder
	)
	for _, c := range v.child {
		val, cached := c.calcvarval()
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

// vsys defined var table for each block
type vsys struct {
	sync.Mutex
	vars map[string]*_var
}

func (vs *vsys) debug(tab ...string) {
	if !enabled.Debug() {
		return
	}
	indent := strings.Join(tab, "")
	fmt.Println(indent + "variables in block:")
	for k, v := range vs.vars {
		fmt.Printf(indent+"\tname:'%s', value:'%s', exp:'%t', addr:%p, segments:'%+v'\n", k, v.v, v.asexp, v, v.segments)
		for _, c := range v.child {
			fmt.Printf(indent+"\t\taddr:'%p', value:'%s', exp:'%t'\n", c, c.v, c.asexp)
		}
	}
}

func (vs *vsys) putOrUpdate(name string, v *_var) error {
	vs.Lock()
	defer vs.Unlock()

	old, ok := vs.vars[name]
	if ok {
		old.updateval(v)
	} else {
		vs.vars[name] = v
	}
	return nil
}

func (vs *vsys) put(name string, v *_var) error {
	vs.Lock()
	defer vs.Unlock()

	_, ok := vs.vars[name]
	if ok {
		return fmt.Errorf("'%s': %w", name, ErrVariableNameDuplicated)
	}
	vs.vars[name] = v
	return nil
}

func (vs *vsys) get(name string) (*_var, bool) {
	vs.Lock()
	defer vs.Unlock()

	v, ok := vs.vars[name]
	return v, ok
}

func (vs *vsys) calc(name string) (_v interface{}, cached bool) {
	v, ok := vs.get(name)
	if !ok {
		return nil, false
	}
	return v.calcvarval()
}

func (vs *vsys) cyclecheck(names ...string) error {
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

func token2var(t *Token) (*_var, error) {
	v := &_var{
		v:        t.String(),
		segments: t.Segments(),
		asexp:    t.typ == _expr_t,
	}
	if !t.HasVar() {
		v.cached = true
	}
	for _, seg := range v.segments {
		if !seg.isvar {
			continue
		}
		vname := seg.str
		chld, _ := t._b.GetVar(vname)
		if chld != nil {
			v.child = append(v.child, chld)
		} else {
			return nil, TokenErrorf(t.ln, ErrVariableNotDefined, "'%s', variable name '%s'", t, vname)
		}
	}
	return v, nil
}

func statement2var(stm *Statement) (*_var, error) {
	var (
		v   *_var
		err error
	)
	if len(stm.tokens) == 2 {
		vt := stm.tokens[1]
		if v, err = token2var(vt); err != nil {
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
