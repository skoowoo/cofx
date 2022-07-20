//go:generate stringer -type TokenType
package cofunc

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cofunclabs/cofunc/pkg/debug"
	"github.com/cofunclabs/cofunc/pkg/is"
)

// Token
//
type TokenType int

const (
	_unknow_t TokenType = iota
	_ident_t
	_symbol_t
	_number_t
	_string_t
	_mapkey_t
	_operator_t
	_functionname_t
	_load_t
	_keyword_t
	_varname_t
)

var tokenPatterns = map[TokenType]*regexp.Regexp{
	_unknow_t:       regexp.MustCompile(`^*$`),
	_string_t:       regexp.MustCompile(`^*$`),
	_ident_t:        regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_\.]*$`),
	_number_t:       regexp.MustCompile(`^[1-9][0-9]*$`),
	_mapkey_t:       regexp.MustCompile(`^[^:]+$`), // not contain ":"
	_operator_t:     regexp.MustCompile(`^(=|->)$`),
	_load_t:         regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*:.*[a-zA-Z0-9]$`),
	_functionname_t: regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`),
	_keyword_t:      regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`),
	_varname_t:      regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`),
}

type Token struct {
	str         string
	typ         TokenType
	ln          int
	_b          *Block
	_persistent string
	_segments   []struct {
		str   string
		isvar bool
	}
	_get func(*Block, string) (string, bool)
}

// _lookupVar be called at running, not parsing
func _lookupVar(b *Block, name string) (string, bool) {
	return b.CalcVar(name)
}

func (t *Token) Segments() []struct {
	str   string
	isvar bool
} {
	return t._segments
}

func (t *Token) IsEmpty() bool {
	return len(t.str) == 0
}

func (t *Token) String() string {
	return t.str
}

func (t *Token) validate() error {
	if pattern := tokenPatterns[t.typ]; !pattern.MatchString(t.str) {
		return TokenErrorf(t.ln, ErrTokenRegex, "actual '%s', expect '%s'", t, pattern)
	}

	// check var
	for _, seg := range t._segments {
		if !seg.isvar {
			continue
		}
		name := seg.str
		if strings.Contains(seg.str, ".") {
			fields := strings.Split(seg.str, ".")
			if len(fields) != 2 {
				return VarErrorf(t.ln, ErrVariableFormat, "'%s' in token '%s'", name, t)
			}
			f1, f2 := fields[0], fields[1]
			if f1 == "" || f2 == "" {
				return VarErrorf(t.ln, ErrVariableFormat, "'%s' in token '%s'", name, t)
			}
			name = f1
		}
		if v, _ := t._b.GetVar(name); v == nil {
			return VarErrorf(t.ln, ErrVariableNotDefined, "'%s' in token '%s'", name, t)
		}
	}
	return nil
}

// @running
// Value will calcuate the variable's value, if the token contain some variables
func (t *Token) Value() string {
	if !t.HasVar() {
		return t.str
	}
	if len(t._persistent) != 0 {
		return t._persistent
	}
	if t._get == nil {
		t._get = _lookupVar
	}
	var bd strings.Builder
	cacheable := true
	for _, seg := range t._segments {
		if seg.isvar {
			val, cached := t._get(t._b, seg.str)
			if !cached {
				cacheable = false
			}
			bd.WriteString(val)
		} else {
			bd.WriteString(seg.str)
		}
	}
	s := bd.String()
	if cacheable {
		// cache the token
		t._persistent = s
	}
	return s
}

func (t *Token) HasVar() bool {
	for _, seg := range t._segments {
		if seg.isvar {
			return true
		}
	}
	return false
}

func (t *Token) extractVar() error {
	// $(var)
	if t.typ != _string_t {
		return nil
	}
	var (
		start  int
		vstart int
		state  aststate
	)
	l := len(t.str)
	next := func(i int) byte {
		i += 1
		if i >= l {
			return 'x'
		}
		return t.str[i]
	}
	for i, c := range t.str {
		switch state {
		case _ast_unknow:
			// skip
			// transfer
			if c == '$' && next(i) == '(' {
				vstart = i
				state = _ast_identifier
			}
		case _ast_identifier: // from '$'
			// keep
			if is.Identifier(c) || c == '(' {
				break
			}
			// transfer
			if c == ')' {
				if j := vstart - 1; j >= 0 {
					if slash := t.str[j]; slash == '\\' {
						// drop '\'
						state = _ast_unknow
						if start < j {
							t._segments = append(t._segments, struct {
								str   string
								isvar bool
							}{t.str[start:j], false})
						}
						start = j + 1 //  j+1 = vstart
						break
					}
				}
				name := t.str[vstart+2 : i] // start +2: skip "$("
				if name == "" {
					return VarErrorf(t.ln, ErrVariableNameEmpty, "token '%s'", t)
				}

				if start < vstart {
					t._segments = append(t._segments, struct {
						str   string
						isvar bool
					}{t.str[start:vstart], false})
				}

				t._segments = append(t._segments, struct {
					str   string
					isvar bool
				}{name, true})

				start = i + 1 // currently i is ')'
				state = _ast_unknow
			}
		}
	}
	if start < len(t.str) {
		t._segments = append(t._segments, struct {
			str   string
			isvar bool
		}{t.str[start:], false})
	}
	return nil
}

// Statement
//
type Statement struct {
	desc   string
	tokens []*Token
}

func newstm(desc string) *Statement {
	stm := &Statement{desc: desc}
	return stm
}

func (s *Statement) LastToken() *Token {
	l := len(s.tokens)
	if l == 0 {
		return nil
	}
	return s.tokens[l-1]
}

func (s *Statement) Append(t *Token) *Statement {
	s.tokens = append(s.tokens, t)
	return s
}

// Block
//
type bbody interface {
	Type() string
	Append(o interface{}) error
	List() []*Statement
	Len() int
}

type plainbody struct {
	lines []*Statement
}

func (r *plainbody) Len() int {
	return len(r.lines)
}

func (r *plainbody) List() []*Statement {
	return r.lines
}

func (r *plainbody) Type() string {
	return "raw"
}

func (r *plainbody) Append(o interface{}) error {
	stm := o.(*Statement)
	r.lines = append(r.lines, stm)
	return nil
}

func (r *plainbody) Laststm() *Statement {
	l := len(r.lines)
	if l == 0 {
		panic("not found statement")
	}
	return r.lines[l-1]
}

type Block struct {
	kind      Token
	target    Token
	operator  Token
	typevalue Token
	child     []*Block
	parent    *Block
	variable  vsys
	bbody
}

// Getvar lookup variable by name in map
func (b *Block) GetVar(name string) (*_var, *Block) {
	for p := b; p != nil; p = p.parent {
		v, ok := p.variable.get(name)
		if !ok {
			continue
		}
		return v, p
	}
	return nil, nil
}

// PutVar insert a variable into map
func (b *Block) PutVar(name string, v *_var) error {
	return b.variable.put(name, v)
}

// UpdateVar insert or update a variable into map
func (b *Block) UpdateVar(name string, v *_var) error {
	return b.variable.putOrUpdate(name, v)
}

// CreateFieldVar TODO:
func (b *Block) CreateFieldVar(name, field, val string) error {
	s := name + "." + field
	v := &_var{
		v:      val,
		cached: false,
	}
	return b.UpdateVar(s, v)
}

// CalcVar calcuate the variable's value
func (b *Block) CalcVar(name string) (string, bool) {
	if b == nil {
		panic(fmt.Sprintf("var name '%s': block is nil", name))
	}
	var _debug_ strings.Builder
	for p := b; p != nil; p = p.parent {
		if debug.Enabled() {
			_debug_.WriteByte('\t')
			_debug_.WriteString(p.String())
			_debug_.WriteByte('\n')
		}

		v, cached := p.variable.calc(name)
		if v == nil {
			continue
		}
		debug.Log("*Block.CalcVar()", "calcute variable succeed: '%s', query path:\n%s\n", name, _debug_.String())
		return v.(string), cached
	}

	debug.Log("*Block.CalcVar()", "calcute variable failed: '%s', query path:\n%s\n", name, _debug_.String())
	if strings.Contains(name, ".") {
		return "", false
	}
	panic("not found variable: " + name)
}

func (b *Block) validate() error {
	ts := []*Token{
		&b.kind,
		&b.target,
		&b.operator,
		&b.typevalue,
	}
	for _, t := range ts {
		if err := t.validate(); err != nil {
			return err
		}
	}
	if b.bbody == nil {
		return nil
	}
	lines := b.bbody.List()
	for _, l := range lines {
		// handle tokens
		for _, t := range l.tokens {
			if err := t.validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *Block) insertVar(stm *Statement) error {
	if stm.desc != "var" {
		return nil
	}
	name := stm.tokens[0].String()
	v := &_var{
		segments: []struct {
			str   string
			isvar bool
		}{},
		child: []*_var{},
	}
	if len(stm.tokens) == 2 {
		vt := stm.tokens[1]
		if !vt.HasVar() {
			v.v = vt.String()
			v.cached = true
		} else {
			v.segments = vt.Segments()
			for _, seg := range v.segments {
				if !seg.isvar {
					continue
				}
				vname := seg.str
				chld, _ := b.GetVar(vname)
				if chld != nil {
					v.child = append(v.child, chld)
				} else {
					return TokenErrorf(vt.ln, ErrVariableNotDefined, "'%s', variable name '%s'", vt, vname)
				}
			}
		}
	}
	if err := b.PutVar(name, v); err != nil {
		return StatementTokensErrorf(err, stm.tokens)
	}
	b.CalcVar(name)
	return nil
}

func (b *Block) Iskind(s string) bool {
	return b.kind.String() == s
}

func (b *Block) IsArgs() bool {
	return b.Iskind(_kw_args)
}

func (b *Block) IsCo() bool {
	return b.Iskind(_kw_co)
}

func (b *Block) IsVar() bool {
	return b.Iskind(_kw_var)
}

func (b *Block) IsFn() bool {
	return b.Iskind(_kw_fn)
}

func (b *Block) IsLoad() bool {
	return b.Iskind(_kw_load)
}

func (b *Block) IsGlobal() bool {
	return b.Iskind("global")
}

func (b *Block) IsFor() bool {
	return b.Iskind(_kw_for)
}

func (b *Block) String() string {
	if b.bbody != nil {
		return fmt.Sprintf("%s %s %s %s {}", &b.kind, &b.target, &b.operator, &b.typevalue)
	} else {
		return fmt.Sprintf("%s %s %s %s", &b.kind, &b.target, &b.operator, &b.typevalue)
	}
}
