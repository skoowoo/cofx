package parser

import (
	"fmt"
	"strings"

	"github.com/cofxlabs/cofx/pkg/enabled"
)

type Block struct {
	kind     Token
	target1  Token
	operator Token
	target2  Token
	child    []*Block
	parent   *Block
	vtbl     vartable
	body
}

func (b *Block) Child() []*Block {
	return b.child
}

func (b *Block) Body() body {
	return b.body
}

func (b *Block) Parent() *Block {
	return b.parent
}

func (b *Block) Target1() *Token {
	return &b.target1
}

func (b *Block) Target2() *Token {
	return &b.target2
}

func (b *Block) RewriteVar(stm *Statement) error {
	stm = stm.Copy()
	if stm.desc != "rewrite_var" {
		return nil
	}
	name := stm.tokens[0].String()

	// Eliminate the circular dependency of the variable itself to itself
	s, _ := b.calcVar(name)
	segments := stm.tokens[1]._segments
	for i, seg := range segments {
		if seg.isvar && seg.str == name {
			segments[i].str = s
			segments[i].isvar = false
		}
	}

	v, err := newVarFromStm(stm)
	if err != nil {
		return err
	}
	if _, inblock := b.getVar(name); inblock != nil {
		inblock.putVar(name, v)
	} else {
		return fmt.Errorf("%w: rewrite var '%s'", ErrVariableNotDefined, name)
	}

	if err := b.vtbl.cyclecheck(name); err != nil {
		return err
	}
	return nil
}

func (b *Block) AddField2Var(name, field, val string) error {
	v, _ := b.getVar(name)
	if v == nil {
		return fmt.Errorf("%w: variable '%s'", ErrVariableNotDefined, name)
	}
	v.addField(field, val)
	return nil
}

func (b *Block) ExecCondition() bool {
	_, ok := b.vtbl.get(_condition_expr_var)
	if !ok {
		// not found condition var in the block
		return true
	}
	s, _ := b.calcVar(_condition_expr_var)
	return s == "true"
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

func (b *Block) IsBtf() bool {
	return b.Iskind("btf")
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

func (b *Block) IsEvent() bool {
	return b.Iskind(_kw_event)
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

	if b.body != nil {
		builder.WriteString("{}")
	}
	return builder.String()
}

func (b *Block) Debug() {
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

	b.vtbl.debug("\t")
}

// GetVarValue returns the value of the variable, the argument is the variable name
func (b *Block) GetVarValue(name string) string {
	defer func() {
		recover()
	}()
	v, _ := b.calcVar(name)
	return v
}

// GetVar lookup variable by name in map
func (b *Block) getVar(name string) (*_var, *Block) {
	for p := b; p != nil; p = p.parent {
		v, ok := p.vtbl.get(name)
		if !ok {
			continue
		}
		return v, p
	}
	return nil, nil
}

// addVar insert a variable into map
func (b *Block) addVar(name string, v *_var) error {
	return b.vtbl.add(name, v)
}

// putVar insert or update a variable into map
func (b *Block) putVar(name string, v *_var) {
	b.vtbl.put(name, v)
}

// calcVar calcuate the variable's value
func (b *Block) calcVar(name string) (string, bool) {
	if b == nil {
		panic(fmt.Sprintf("var name '%s': block is nil", name))
	}

	for p := b; p != nil; p = p.parent {
		v, cached := p.vtbl.calc(name)
		if v == nil {
			continue
		}
		return v.(string), cached
	}

	panic("not found variable: " + name)
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
	if b.body == nil {
		return nil
	}
	lines := b.body.List()
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
	v, err := newVarFromStm(stm)
	if err != nil {
		return err
	}
	if err := b.addVar(name, v); err != nil {
		return statementTokensErrorf(err, stm.tokens)
	}
	if err := b.vtbl.cyclecheck(name); err != nil {
		return err
	}
	return nil
}

type body interface {
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

type MapBody struct {
	plainbody
}

func (m *MapBody) ToMap() map[string]string {
	ret := make(map[string]string)
	for _, ln := range m.lines {
		k, v := ln.tokens[0].value(), ln.tokens[1].value()
		ret[k] = v
	}
	return ret
}

func (m *MapBody) Append(o interface{}) error {
	ts := o.([]*Token)
	if len(ts) != 3 {
		return statementTokensErrorf(ErrMapKVIllegal, ts)
	}
	k, delim, v := ts[0], ts[1], ts[2]
	if !k.TypeEqual(_string_t) {
		return tokenTypeErrorf(k, _string_t)
	}
	if !delim.TypeEqual(_symbol_t) {
		return tokenTypeErrorf(delim, _symbol_t)
	}
	if delim.String() != ":" {
		return tokenValueErrorf(delim, ":")
	}
	if !v.TypeEqual(_string_t) {
		return tokenTypeErrorf(k, _string_t)
	}
	m.lines = append(m.lines, NewStatement("kv").Append(k).Append(v))

	return nil
}

type ListBody struct {
	plainbody
	etype TokenType
}

func (l *ListBody) ToSlice() []string {
	var ret []string
	for _, ln := range l.lines {
		v := ln.tokens[0].value()
		ret = append(ret, v)
	}
	return ret
}

func (l *ListBody) Append(o interface{}) error {
	ts := o.([]*Token)
	if len(ts) != 1 {
		return statementTokensErrorf(ErrListElemIllegal, ts)
	}
	t := ts[0]
	t.typ = l.etype
	l.lines = append(l.lines, NewStatement("element").Append(t))
	return nil
}
