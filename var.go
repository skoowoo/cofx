package cofunc

import (
	"errors"
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

// vsys defined var table for each block
type vsys struct {
	vars map[string]*_var
}

func (vs *vsys) put(name string, v *_var) error {
	_, ok := vs.vars[name]
	if ok {
		return errors.New("variable name is same: " + name)
	}
	vs.vars[name] = v
	return nil
}

func (vs *vsys) get(name string) (*_var, bool) {
	v, ok := vs.vars[name]
	return v, ok
}

func (vs *vsys) calc(name string) (_v interface{}, cached bool) {
	v, ok := vs.vars[name]
	if !ok {
		return nil, false
	}
	return calcvar(v)
}

func calcvar(v *_var) (string, bool) {
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
		val, cached := calcvar(c)
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
