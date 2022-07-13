//go:generate stringer -type aststate
package cofunc

import (
	"bufio"
	"errors"
	"io"
	"math"
)

func ParseAST(rd io.Reader) (*AST, error) {
	lx := newLexer()
	buff := bufio.NewReader(rd)
	for n := 1; ; n += 1 {
		line, err := buff.ReadString('\n')
		if err == io.EOF {
			if len(line) != 0 {
				if err := lx.split(line, n); err != nil {
					return nil, err
				}
			}
			break
		}
		if err != nil {
			return nil, err
		}
		if err := lx.split(line, n); err != nil {
			return nil, err
		}
	}

	ast := newAST()
	if err := ast.scan(lx); err != nil {
		return nil, err
	}

	return ast, ast.Foreach(func(b *Block) error {
		if err := doBlockHeader(b); err != nil {
			return err
		}

		if err := doBlockBody(b); err != nil {
			return err
		}
		return nil
	})
}

func doBlockBody(b *Block) error {
	if b.bbody == nil {
		return nil
	}
	lines := b.bbody.List()
	for _, l := range lines {
		// handle tokens
		for _, t := range l.tokens {
			if err := t.extractVar(); err != nil {
				return err
			}
			if err := t.validate(); err != nil {
				return err
			}
		}

		if err := buildVarGraph(b, l); err != nil {
			return err
		}
	}
	return nil
}

func buildVarGraph(b *Block, stm *Statement) error {
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
				}
			}
		}
	}
	if err := b.PutVar(name, v); err != nil {
		return err
	}
	return nil
}

func doBlockHeader(b *Block) error {
	ts := []*Token{
		&b.kind,
		&b.target,
		&b.operator,
		&b.typevalue,
	}
	for _, t := range ts {
		if err := t.extractVar(); err != nil {
			return err
		}
		if err := t.validate(); err != nil {
			return err
		}
	}
	return nil
}

type aststate int

const (
	_ast_unknow aststate = iota
	_ast_identifier
	_ast_global
	_ast_co_body
	_ast_fn_body
	_ast_args_body
	_ast_for_body
)

type _FA struct {
	state    aststate
	prestate aststate
}

func (f *_FA) _goto(s aststate) {
	f.prestate = f.state
	f.state = s
}

func (f *_FA) phase() aststate {
	return f.state
}

// AST store all blocks in the flowl
//
const (
	_kw_load = "load"
	_kw_fn   = "fn"
	_kw_co   = "co"
	_kw_var  = "var"
	_kw_args = "args"
	_kw_for  = "for"
)

type AST struct {
	global Block

	// for parsing
	_FA
}

func newAST() *AST {
	ast := &AST{
		global: Block{
			kind: Token{
				str: "global",
			},
			target:    Token{},
			operator:  Token{},
			typevalue: Token{},
			child:     make([]*Block, 0),
			parent:    nil,
			variable:  vsys{vars: make(map[string]*_var)},
			bbody:     &plainbody{},
		},
		_FA: _FA{
			state: _ast_global,
		},
	}
	return ast
}

func (a *AST) Foreach(do func(*Block) error) error {
	return deepwalk(&a.global, do)
}

func deepwalk(b *Block, do func(*Block) error) error {
	if err := do(b); err != nil {
		return err
	}
	for _, c := range b.child {
		if err := deepwalk(c, do); err != nil {
			return err
		}
	}
	return nil
}

var statementPatterns = map[string]struct {
	min     int
	max     int
	types   []TokenType
	values  []string
	uptypes []TokenType
	newbody func() bbody
}{
	"load": {
		2, 2,
		[]TokenType{_identifier_t, _string_t},
		[]string{"", ""},
		[]TokenType{_keyword_t, _load_t},
		nil,
	},
	"fn": {
		5, 5,
		[]TokenType{_identifier_t, _identifier_t, _symbol_t, _identifier_t, _symbol_t},
		[]string{"", "", "=", "", "{"},
		[]TokenType{_keyword_t, _functionname_t, _operator_t, _functionname_t, _symbol_t},
		func() bbody { return &plainbody{} },
	},
	"co1": {
		2, 2,
		[]TokenType{_identifier_t, _identifier_t},
		[]string{"", ""},
		[]TokenType{_keyword_t, _functionname_t},
		nil,
	},
	"co1->": {
		4, 4,
		[]TokenType{_identifier_t, _identifier_t, _symbol_t, _identifier_t},
		[]string{"", "", "->", ""},
		[]TokenType{_keyword_t, _functionname_t, _operator_t, _varname_t},
		nil,
	},
	"co1+": {
		3, 3,
		[]TokenType{_identifier_t, _identifier_t, _symbol_t},
		[]string{"", "", "{"},
		[]TokenType{_keyword_t, _functionname_t, _symbol_t},
		func() bbody { return &FMap{} },
	},
	"co1+->": {
		5, 5,
		[]TokenType{_identifier_t, _identifier_t, _symbol_t, _identifier_t, _symbol_t},
		[]string{"", "", "->", "", "{"},
		[]TokenType{_keyword_t, _functionname_t, _operator_t, _varname_t, _symbol_t},
		func() bbody { return &FMap{} },
	},
	"co2": {
		2, 2,
		[]TokenType{_identifier_t, _symbol_t},
		[]string{"", "{"},
		[]TokenType{_keyword_t, _symbol_t},
		func() bbody { return &FList{etype: _functionname_t} },
	},
	"var": {
		2, math.MaxInt,
		[]TokenType{_identifier_t, _identifier_t, _symbol_t},
		[]string{"", "", "="},
		[]TokenType{_keyword_t, _varname_t, _operator_t},
		nil,
	},
	"args": {
		3, 3,
		[]TokenType{_identifier_t, _symbol_t, _symbol_t},
		[]string{"", "=", "{"},
		[]TokenType{_keyword_t, _operator_t, _symbol_t},
		func() bbody { return &FMap{} },
	},
	"kv": {
		3, 3,
		[]TokenType{_string_t, _symbol_t, _string_t},
		[]string{"", ":", ""},
		[]TokenType{_mapkey_t, _symbol_t, _string_t},
		nil,
	},
	"element": {
		1, 1,
		[]TokenType{_string_t},
		[]string{""},
		[]TokenType{_string_t},
		nil,
	},
	"for": {
		2, 2,
		[]TokenType{_identifier_t, _symbol_t},
		[]string{"", "{"},
		[]TokenType{_keyword_t, _symbol_t},
		nil,
	},
	"closed": {
		1, 1,
		[]TokenType{_symbol_t},
		[]string{"}"},
		[]TokenType{_symbol_t},
		nil,
	},
}

func (ast *AST) preparse(k string, line []*Token, ln int, b *Block) (bbody, error) {
	pattern := statementPatterns[k]

	if l := len(line); l < pattern.min || l > pattern.max {
		return nil, errors.New("invalid statement(token number): " + k)
	}

	min := len(line)
	if l := len(pattern.types); min > l {
		min = l
	}

	for i := 0; i < min; i++ {
		t := line[i]
		expectTyp := pattern.types[i]
		expectVal := pattern.values[i]

		if expectTyp != t.typ {
			return nil, parseTokenTypeErr().New(line, ln, t, expectTyp)
		}
		if expectVal != "" && expectVal != t.String() {
			return nil, parseTokenValErr().New(line, ln, t, expectVal)
		}
	}

	for i := 0; i < min; i++ {
		t := line[i]
		up := pattern.uptypes[i]
		t.typ = up
	}

	for _, t := range line {
		t._b = b
	}

	var body bbody
	if pattern.newbody != nil {
		body = pattern.newbody()
	}

	return body, nil
}

func (ast *AST) parseVar(line []*Token, ln int, b *Block) error {
	if _, err := ast.preparse("var", line, ln, b); err != nil {
		return err
	}
	name := line[1]
	var val *Token
	if len(line) == 4 {
		val = line[3]
	}

	stm := newstm("var")
	stm.Append(name)
	if val != nil {
		stm.Append(val)
	}
	b.bbody.Append(stm)
	return nil
}

func (ast *AST) parseLoad(line []*Token, ln int, b *Block) error {
	nb := &Block{
		child:    []*Block{},
		parent:   b,
		variable: vsys{vars: make(map[string]*_var)},
	}
	body, err := ast.preparse("load", line, ln, nb)
	if err != nil {
		return err
	}
	nb.bbody = body
	nb.kind = *line[0]
	nb.target = *line[1]

	b.child = append(b.child, nb)
	return nil
}

func (ast *AST) parseFn(line []*Token, ln int, b *Block) (*Block, error) {
	nb := &Block{
		child:    []*Block{},
		parent:   b,
		variable: vsys{vars: make(map[string]*_var)},
		bbody:    &plainbody{},
	}
	body, err := ast.preparse("fn", line, ln, nb)
	if err != nil {
		return nil, err
	}
	nb.bbody = body

	kind, target, op, tv := line[0], line[1], line[2], line[3]

	nb.kind = *kind
	nb.target = *target
	nb.operator = *op
	nb.typevalue = *tv

	b.child = append(b.child, nb)
	return nb, nil
}

func (ast *AST) parseCo(line []*Token, ln int, b *Block) (*Block, error) {
	nb := &Block{
		child:    []*Block{},
		parent:   b,
		variable: vsys{vars: make(map[string]*_var)},
		bbody:    nil,
	}

	var (
		body bbody
		err  error
	)
	keys := []string{"co1", "co1+", "co2", "co1->", "co1+->"}
	for _, k := range keys {
		body, err = ast.preparse(k, line, ln, nb)
		if err == nil {
			nb.kind = *line[0]
			nb.bbody = body
			switch k {
			case "co1": // co sleep
				nb.target = *line[1]
			case "co1+": // co sleep {
				nb.target = *line[1]
			case "co1->": // co sleep -> out
				nb.target = *line[1]
				nb.operator = *line[2]
				nb.typevalue = *line[3]
			case "co1+->": // co sleep -> out {
				nb.target = *line[1]
				nb.operator = *line[2]
				nb.typevalue = *line[3]
			case "co2": // co {
			}
			break
		}
	}
	if err != nil {
		return nil, err
	}

	b.child = append(b.child, nb)
	return nb, nil
}

func (ast *AST) parseArgs(line []*Token, ln int, b *Block) (*Block, error) {
	nb := &Block{
		child:    []*Block{},
		parent:   b,
		variable: vsys{vars: make(map[string]*_var)},
	}
	body, err := ast.preparse("args", line, ln, nb)
	if err != nil {
		return nil, err
	}
	nb.bbody = body
	nb.kind = *line[0]

	b.child = append(b.child, nb)
	return nb, nil
}

func (ast *AST) parseFor(line []*Token, ln int, b *Block) (*Block, error) {
	nb := &Block{
		child:    []*Block{},
		parent:   b,
		variable: vsys{vars: make(map[string]*_var)},
	}
	body, err := ast.preparse("for", line, ln, nb)
	if err != nil {
		return nil, err
	}
	nb.bbody = body
	nb.kind = *line[0]

	b.child = append(b.child, nb)
	return nb, nil
}

func (ast *AST) scan(lx *lexer) error {
	var parsingblock = &ast.global

	return lx.foreachLine(func(ln int, line []*Token) error {
		if len(line) == 0 {
			return nil
		}
		switch ast.phase() {
		case _ast_global:
			kind := line[0]
			switch kind.String() {
			case "//":
				return nil
			case _kw_load:
				return ast.parseLoad(line, ln, parsingblock)
			case _kw_fn:
				fnblock, err := ast.parseFn(line, ln, parsingblock)
				if err != nil {
					return err
				}
				parsingblock = fnblock
				ast._goto(_ast_fn_body)
			case _kw_co:
				coblock, err := ast.parseCo(line, ln, parsingblock)
				if err != nil {
					return err
				}
				if coblock.bbody != nil {
					parsingblock = coblock
					ast._goto(_ast_co_body)
				}
			case _kw_var:
				return ast.parseVar(line, ln, parsingblock)
			case _kw_for:
				forblock, err := ast.parseFor(line, ln, parsingblock)
				if err != nil {
					return err
				}
				parsingblock = forblock
				ast._goto(_ast_for_body)
			default:
				return errors.New("invalid block define: " + kind.String())
			}
		case _ast_fn_body:
			if _, err := ast.preparse("closed", line, ln, parsingblock); err == nil {
				parsingblock = parsingblock.parent
				ast._goto(_ast_global)
				break
			}

			kind := line[0]
			switch kind.String() {
			case _kw_args:
				argsblock, err := ast.parseArgs(line, ln, parsingblock)
				if err != nil {
					return err
				}
				parsingblock = argsblock
				ast._goto(_ast_args_body)
			case _kw_var:
				return ast.parseVar(line, ln, parsingblock)
			default:
				return errors.New("invalid statement: " + kind.String())
			}
		case _ast_args_body:
			if _, err := ast.preparse("closed", line, ln, parsingblock); err == nil {
				parsingblock = parsingblock.parent
				ast._goto(_ast_fn_body)
				break
			}
			for _, t := range line {
				t._b = parsingblock
			}
			if err := parsingblock.bbody.Append(line); err != nil {
				return err
			}
		case _ast_co_body:
			if _, err := ast.preparse("closed", line, ln, parsingblock); err == nil {
				parsingblock = parsingblock.parent
				if parsingblock.IsFor() {
					ast._goto(_ast_for_body)
				} else {
					ast._goto(_ast_global)
				}
				break
			}

			if err := parsingblock.bbody.Append(line); err != nil {
				return err
			}
		case _ast_for_body:
			if _, err := ast.preparse("closed", line, ln, parsingblock); err == nil {
				parsingblock = parsingblock.parent
				ast._goto(_ast_global)
				break
			}

			kind := line[0]
			switch kind.String() {
			case _kw_co:
				coblock, err := ast.parseCo(line, ln, parsingblock)
				if err != nil {
					return err
				}
				if coblock.bbody != nil {
					parsingblock = coblock
					ast._goto(_ast_co_body)
				}
			default:
				return errors.New("invalid statement: " + kind.String())
			}
		}

		return nil
	})
}
