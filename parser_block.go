//go:generate stringer -type TokenType
package cofunc

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cofunclabs/cofunc/pkg/enabled"
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
	_refvar_t
	_mapkey_t
	_operator_t
	_functionname_t
	_load_t
	_keyword_t
	_varname_t
	_expr_t
)

var tokenPatterns = map[TokenType]*regexp.Regexp{
	_unknow_t:       regexp.MustCompile(`^*$`),
	_string_t:       regexp.MustCompile(`^*$`),
	_refvar_t:       regexp.MustCompile(`^\$\([a-zA-Z0-9_\.]*\)$`),
	_ident_t:        regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_\.]*$`),
	_number_t:       regexp.MustCompile(`^[0-9\.]+$`),
	_mapkey_t:       regexp.MustCompile(`^[^:]+$`), // not contain ":"
	_operator_t:     regexp.MustCompile(`^(=|->)$`),
	_load_t:         regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*:.*[a-zA-Z0-9]$`),
	_functionname_t: regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`),
	_keyword_t:      regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`),
	_varname_t:      regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`),
}

type Token struct {
	str       string
	typ       TokenType
	ln        int
	_b        *Block
	_segments []struct {
		str   string
		isvar bool
	}
	_get func(*Block, string) (string, bool)
}

func _lookupVar(b *Block, name string) (string, bool) {
	return b.CalcVar(name)
}

func (t *Token) Segments() []struct {
	str   string
	isvar bool
} {
	return t._segments
}

func (t *Token) CopySegments() []struct {
	str   string
	isvar bool
} {
	var segments []struct {
		str   string
		isvar bool
	}

	segments = append(segments, t._segments...)
	return segments
}

func (t *Token) IsEmpty() bool {
	return len(t.str) == 0
}

func (t *Token) String() string {
	return t.str
}

func (t *Token) FormatString() string {
	return fmt.Sprintf("['%s','%s']", t.str, t.typ)
}

func (t *Token) validate() error {
	if pattern, ok := tokenPatterns[t.typ]; ok {
		if !pattern.MatchString(t.str) {
			return TokenErrorf(t.ln, ErrTokenRegex, "actual '%s', expect '%s'", t, pattern)
		}
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

// Value will calcuate the variable's value, if the token contain some variables
func (t *Token) Value() string {
	if !t.HasVar() {
		return t.str
	}
	if t._get == nil {
		t._get = _lookupVar
	}
	var bd strings.Builder
	for _, seg := range t._segments {
		if seg.isvar {
			val, cached := t._get(t._b, seg.str)
			if !cached {
				_ = cached
			}
			bd.WriteString(val)
		} else {
			bd.WriteString(seg.str)
		}
	}
	return bd.String()
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
	if t.typ != _string_t && t.typ != _expr_t && t.typ != _refvar_t {
		return nil
	}
	// Avoid repeated to extract the variable
	if len(t._segments) != 0 {
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
				state = _ast_ident
			}
		case _ast_ident: // from '$'
			// keep
			if is.Ident(c) || c == '(' {
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

func (s *Statement) Copy() *Statement {
	stm := &Statement{
		desc: s.desc,
	}
	for _, t := range s.tokens {
		nt := &Token{
			str:       t.str,
			typ:       t.typ,
			ln:        t.ln,
			_b:        t._b,
			_segments: t.CopySegments(),
			_get:      t._get,
		}
		stm.tokens = append(stm.tokens, nt)
	}
	return stm
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
	target1   Token
	operator  Token
	target2   Token
	child     []*Block
	parent    *Block
	variables vsys
	bbody
}

// Getvar lookup variable by name in map
func (b *Block) GetVar(name string) (*_var, *Block) {
	for p := b; p != nil; p = p.parent {
		v, ok := p.variables.get(name)
		if !ok {
			continue
		}
		return v, p
	}
	return nil, nil
}

// PutVar insert a variable into map
func (b *Block) PutVar(name string, v *_var) error {
	return b.variables.put(name, v)
}

// UpdateVar insert or update a variable into map
func (b *Block) UpdateVar(name string, v *_var) error {
	return b.variables.putOrUpdate(name, v)
}

// CalcVar calcuate the variable's value
func (b *Block) CalcVar(name string) (string, bool) {
	if b == nil {
		panic(fmt.Sprintf("var name '%s': block is nil", name))
	}

	main, field, ok := isFieldVar(name)
	if ok {
		v, _ := b.GetVar(main)
		if v == nil {
			return "", false
		}
		return v.readField(field), false
	}

	var _debug_ strings.Builder
	for p := b; p != nil; p = p.parent {
		if enabled.Debug() {
			_debug_.WriteByte('\t')
			_debug_.WriteString(p.String())
			_debug_.WriteByte('\n')
		}

		v, cached := p.variables.calc(name)
		if v == nil {
			continue
		}
		return v.(string), cached
	}

	panic("not found variable: " + name)
}

func (b *Block) CalcConditionTrue() bool {
	s, _ := b.CalcVar(_condition_expr_var)
	return s == "true"
}

func (b *Block) validate() error {
	ts := []*Token{
		&b.kind,
		&b.target1,
		&b.operator,
		&b.target2,
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

func (b *Block) initVar(stm *Statement) error {
	if stm.desc != "var" {
		return nil
	}
	name := stm.tokens[0].String()
	v, err := statement2var(stm)
	if err != nil {
		return err
	}
	if err := b.PutVar(name, v); err != nil {
		return StatementTokensErrorf(err, stm.tokens)
	}
	if err := b.variables.cyclecheck(name); err != nil {
		return err
	}
	return nil
}

func (b *Block) rewriteVar(stm *Statement) error {
	stm = stm.Copy()
	if stm.desc != "rewrite_var" {
		return nil
	}
	name := stm.tokens[0].String()

	// Eliminate the circular dependency of the variable itself to itself
	s, _ := b.CalcVar(name)
	segments := stm.tokens[1].Segments()
	for i, seg := range segments {
		if seg.isvar && seg.str == name {
			segments[i].str = s
			segments[i].isvar = false
		}
	}

	v, err := statement2var(stm)
	if err != nil {
		return err
	}
	if _, inblock := b.GetVar(name); inblock != nil {
		if err := inblock.UpdateVar(name, v); err != nil {
			return StatementTokensErrorf(err, stm.tokens)
		}
	} else {
		return fmt.Errorf("%w: rewrite var '%s'", ErrVariableNotDefined, name)
	}

	if err := b.variables.cyclecheck(name); err != nil {
		return err
	}
	return nil
}

func (b *Block) addField2Var(name, field, val string) error {
	v, _ := b.GetVar(name)
	if v == nil {
		return fmt.Errorf("%w: variable '%s'", ErrVariableNotDefined, name)
	}
	v.addField(field, val)
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

func (b *Block) IsIf() bool {
	return b.Iskind(_kw_if)
}

func (b *Block) IsSwitch() bool {
	return b.Iskind(_kw_switch)
}

func (b *Block) IsCase() bool {
	return b.Iskind(_kw_case)
}

func (b *Block) IsDefault() bool {
	return b.Iskind(_kw_default)
}

func (b *Block) InFor() bool {
	var p *Block
	for p = b.parent; p != nil; p = p.parent {
		if p.IsFor() {
			return true
		}
	}
	return false
}

func (b *Block) InSwitch() bool {
	return b.parent.IsCase() || b.parent.IsDefault()
}

func (b *Block) String() string {
	var builder strings.Builder

	builder.WriteString(b.kind.String())

	if !b.target1.IsEmpty() {
		builder.WriteString(" ")
		builder.WriteString(b.target1.String())
	}

	if !b.operator.IsEmpty() {
		builder.WriteString(" ")
		builder.WriteString(b.operator.String())
	}

	if !b.target2.IsEmpty() {
		builder.WriteString(" ")
		builder.WriteString(b.target2.String())
	}

	if b.bbody != nil {
		builder.WriteString("{}")
	}
	return builder.String()
}

func (b *Block) debug() {
	if !enabled.Debug() {
		return
	}

	fmt.Printf("---> block: '%s'\n", b.String())
	if b.parent != nil {
		fmt.Printf("\tparent: '%s'\n", b.parent.String())
	} else {
		fmt.Println("\tparent: 'nil'")
	}

	fmt.Println("\tchild:")
	for _, c := range b.child {
		fmt.Printf("\t\t'%s'\n", c.String())
	}

	b.variables.debug("\t")
}
