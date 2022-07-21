package cofunc

import (
	"container/list"
	"fmt"
	"strings"
	"sync"
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
}

func (v *_var) updateval(nv *_var) {
	v.Lock()
	defer v.Unlock()

	v.v = nv.v
	v.segments = nv.segments
	v.child = nv.child
	v.cached = nv.cached
}

func (v *_var) calcvarval() (string, bool) {
	v.Lock()
	defer v.Unlock()

	if v.cached || len(v.child) == 0 {
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
	if cacheable {
		v.cached = true
	}
	v.v = vb.String()
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

func (vs *vsys) cyclecheck() error {
	vs.Lock()
	defer vs.Unlock()

	stack := list.New()
	for name, v := range vs.vars {
		if err := v.dfscycle(stack); err != nil {
			return fmt.Errorf("%w: start variable '%s'", err, name)
		}
	}
	return nil
}

func token2var(t *Token) (*_var, error) {
	v := &_var{
		segments: []struct {
			str   string
			isvar bool
		}{},
		child: []*_var{},
	}
	if !t.HasVar() {
		v.v = t.String()
		v.cached = true
	} else {
		v.segments = t.Segments()
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
