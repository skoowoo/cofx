//go:generate stringer -type aststate
package cofunc

import (
	"bufio"
	"io"
	"math"
	"strings"

	"github.com/cofunclabs/cofunc/pkg/enabled"
)

func init() {
	infertree = _buildInferTree()
}

var infertree *_InferNode

const (
	_kw_comment = "//"
	_kw_load    = "load"
	_kw_fn      = "fn"
	_kw_co      = "co"
	_kw_var     = "var"
	_kw_args    = "args"
	_kw_for     = "for"
)

type aststate int

const (
	_ast_unknow aststate = iota
	_ast_ident
	_ast_global
	_ast_co_body
	_ast_fn_body
	_ast_args_body
	_ast_for_body
)

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
		[]TokenType{_ident_t, _string_t},
		[]string{"", ""},
		[]TokenType{_keyword_t, _load_t},
		nil,
	},
	"fn": {
		5, 5,
		[]TokenType{_ident_t, _ident_t, _symbol_t, _ident_t, _symbol_t},
		[]string{"", "", "=", "", "{"},
		[]TokenType{_keyword_t, _functionname_t, _operator_t, _functionname_t, _symbol_t},
		func() bbody { return &plainbody{} },
	},
	"co1": {
		2, 2,
		[]TokenType{_ident_t, _ident_t},
		[]string{"", ""},
		[]TokenType{_keyword_t, _functionname_t},
		nil,
	},
	"co1->": {
		4, 4,
		[]TokenType{_ident_t, _ident_t, _symbol_t, _ident_t},
		[]string{"", "", "->", ""},
		[]TokenType{_keyword_t, _functionname_t, _operator_t, _varname_t},
		nil,
	},
	"co1+": {
		3, 3,
		[]TokenType{_ident_t, _ident_t, _symbol_t},
		[]string{"", "", "{"},
		[]TokenType{_keyword_t, _functionname_t, _symbol_t},
		func() bbody { return &FMap{} },
	},
	"co1+->": {
		5, 5,
		[]TokenType{_ident_t, _ident_t, _symbol_t, _ident_t, _symbol_t},
		[]string{"", "", "->", "", "{"},
		[]TokenType{_keyword_t, _functionname_t, _operator_t, _varname_t, _symbol_t},
		func() bbody { return &FMap{} },
	},
	"co2": {
		2, 2,
		[]TokenType{_ident_t, _symbol_t},
		[]string{"", "{"},
		[]TokenType{_keyword_t, _symbol_t},
		func() bbody { return &FList{etype: _functionname_t} },
	},
	"var": {
		2, math.MaxInt,
		[]TokenType{_ident_t, _ident_t, _symbol_t},
		[]string{"", "", "="},
		[]TokenType{_keyword_t, _varname_t, _operator_t},
		nil,
	},
	"args": {
		3, 3,
		[]TokenType{_ident_t, _symbol_t, _symbol_t},
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
		[]TokenType{_ident_t, _symbol_t},
		[]string{"", "{"},
		[]TokenType{_keyword_t, _symbol_t},
		func() bbody { return &plainbody{} },
	},
	"closed": {
		1, 1,
		[]TokenType{_symbol_t},
		[]string{"}"},
		[]TokenType{_symbol_t},
		nil,
	},
}

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

	lx.debug()

	ast := newAST()
	if err := ast.scan(lx); err != nil {
		return nil, err
	}

	if enabled.Debug() {
		ast.Foreach(func(b *Block) error {
			b.debug()
			return nil
		})
	}

	return ast, ast.Foreach(func(b *Block) error {
		if err := b.validate(); err != nil {
			return err
		}
		if err := b.variables.cyclecheck(); err != nil {
			return err
		}
		return nil
	})
}

// AST store all blocks in the flowl
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
			variables: vsys{vars: make(map[string]*_var)},
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

func (ast *AST) preparse(k string, line []*Token, ln int, b *Block) (bbody, error) {
	pattern := statementPatterns[k]

	if l := len(line); l < pattern.min || l > pattern.max {
		return nil, TokenErrorf(ln, ErrTokenNumInLine, "actual %d, expect [%d,%d]", l, pattern.min, pattern.max)
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
			return nil, TokenTypeErrorf(t, expectTyp)
		}
		if expectVal != "" && expectVal != t.String() {
			return nil, TokenValueErrorf(t, expectVal)
		}
	}

	for i := 0; i < min; i++ {
		t := line[i]
		up := pattern.uptypes[i]
		t.typ = up
	}

	for _, t := range line {
		t._b = b
		t.ln = ln
		if err := t.extractVar(); err != nil {
			return nil, err
		}
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
		// e.g.:
		// 		var v = "foo"
		// 		var v = 100
		val = line[3]
		if val.typ != _string_t && val.typ != _number_t {
			return VarErrorf(val.ln, ErrVariableValueType, "variable '%s' value '%s' type '%s'", name, val.String(), val.typ)
		}
	} else if len(line) > 4 {
		// e.g.:
		// 		var v = 1 + 1
		// 		var v = -1
		// 		var v = 1 + $(foo)
		// the value is a expression
		var builder strings.Builder
		for _, t := range line[3:] {
			builder.WriteString(t.String())
		}
		exp := builder.String()
		val = &Token{
			str: exp,
			typ: _exp_t,
			ln:  ln,
			_b:  b,
		}
		if err := val.extractVar(); err != nil {
			return err
		}
	}

	var stm *Statement
	if val != nil {
		stm = newstm("var").Append(name).Append(val)
	} else {
		stm = newstm("var").Append(name)
	}
	if err := b.initVar(stm); err != nil {
		return err
	}
	return nil
}

func (ast *AST) parseLoad(line []*Token, ln int, b *Block) error {
	nb := &Block{
		child:     []*Block{},
		parent:    b,
		variables: vsys{vars: make(map[string]*_var)},
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
		child:     []*Block{},
		parent:    b,
		variables: vsys{vars: make(map[string]*_var)},
		bbody:     &plainbody{},
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
		child:     []*Block{},
		parent:    b,
		variables: vsys{vars: make(map[string]*_var)},
		bbody:     nil,
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

	if !nb.typevalue.IsEmpty() {
		name := nb.typevalue.String()
		if v, _ := nb.GetVar(name); v == nil {
			return nil, VarErrorf(nb.typevalue.ln, ErrVariableNotDefined, "'%s'", name)
		}
	}

	b.child = append(b.child, nb)
	return nb, nil
}

func (ast *AST) parseArgs(line []*Token, ln int, b *Block) (*Block, error) {
	nb := &Block{
		child:     []*Block{},
		parent:    b,
		variables: vsys{vars: make(map[string]*_var)},
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
		child:     []*Block{},
		parent:    b,
		variables: vsys{vars: make(map[string]*_var)},
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
		// discard the commments
		if line[0].String() == _kw_comment {
			return nil
		}

		switch ast.phase() {
		case _ast_global:
			kind := line[0]
			switch kind.String() {
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
				if _parse, err := _lookupInferTree(infertree, line); err == nil {
					if err := _parse(parsingblock, line, ln); err != nil {
						return StatementTokensErrorf(err, line)
					}
					return nil
				}
				return StatementTokensErrorf(ErrStatementUnknow, line)
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
				if _parse, err := _lookupInferTree(infertree, line); err == nil {
					if err := _parse(parsingblock, line, ln); err != nil {
						return StatementTokensErrorf(err, line)
					}
					return nil
				}
				return StatementErrorf(ln, ErrStatementUnknow, "%s", kind)
			}
		case _ast_args_body:
			if _, err := ast.preparse("closed", line, ln, parsingblock); err == nil {
				parsingblock = parsingblock.parent
				ast._goto(_ast_fn_body)
				break
			}
			for _, t := range line {
				t.ln = ln
				t._b = parsingblock
				if err := t.extractVar(); err != nil {
					return StatementTokensErrorf(err, line)
				}
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

			for _, t := range line {
				t.ln = ln
				t._b = parsingblock
				if err := t.extractVar(); err != nil {
					return StatementTokensErrorf(err, line)
				}
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
				if _parse, err := _lookupInferTree(infertree, line); err == nil {
					if err := _parse(parsingblock, line, ln); err != nil {
						return StatementTokensErrorf(err, line)
					}
					return nil
				}
				return StatementErrorf(ln, ErrStatementUnknow, "%s", kind)
			}
		}

		return nil
	})
}

type _FA struct {
	state aststate
}

func (f *_FA) _goto(s aststate) {
	f.state = s
}

func (f *_FA) phase() aststate {
	return f.state
}

//
//
type _InferData struct {
	tt TokenType
	tv string
}

type _InferNode struct {
	data   _InferData
	childs []_InferNode
	_parse func(*Block, []*Token, int) error
}

func _lookupInferTree(root *_InferNode, line []*Token) (func(*Block, []*Token, int) error, error) {
	var found = false
	p := root
	for _, t := range line {
		if len(p.childs) == 0 && p._parse != nil {
			return p._parse, nil
		}

		for i, child := range p.childs {
			if child.data.tt != t.typ {
				continue
			}
			if child.data.tv != "" && child.data.tv != t.String() {
				continue
			}
			p = &p.childs[i]
			found = true
		}
		if found {
			found = false
		} else {
			return nil, StatementTokensErrorf(ErrStatementInferFailed, line)
		}
	}
	return p._parse, nil
}

func _buildInferTree() *_InferNode {
	var rules map[string][]_InferData = map[string][]_InferData{
		"rewrite_var1":     {{_ident_t, ""}, {_symbol_t, "<-"}, {_string_t, ""}},                   // e.g.: v <- "foo"
		"rewrite_var2":     {{_ident_t, ""}, {_symbol_t, "<-"}, {_number_t, ""}},                   // e.g.: v <- 100
		"rewrite_var_exp1": {{_ident_t, ""}, {_symbol_t, "<-"}, {_symbol_t, "-"}, {_number_t, ""}}, // e.g.: v <- -100
		"rewrite_var_exp2": {{_ident_t, ""}, {_symbol_t, "<-"}, {_symbol_t, "("}},                  // e.g.: v <- (1 + 2) * 3
		"rewrite_var_exp3": {{_ident_t, ""}, {_symbol_t, "<-"}, {_string_t, ""}, {_symbol_t, ""}},  // e.g.: v <- $(foo) + 1
		"rewrite_var_exp4": {{_ident_t, ""}, {_symbol_t, "<-"}, {_number_t, ""}, {_symbol_t, ""}},  // e.g.: v <- 100 + 1
	}

	root := &_InferNode{}
	p := root
	for k, rule := range rules {
		for _, e := range rule {
			p = _insertInferTree(p, e)
		}
		if strings.HasPrefix(k, "rewrite_var_exp") {
			p._parse = _parseRewriteVarWithExp
		} else if strings.HasPrefix(k, "rewrite_var") {
			p._parse = _parseRewriteVar
		}

		p = root
	}
	return root
}

func _insertInferTree(p *_InferNode, n _InferData) *_InferNode {
	for i, child := range p.childs {
		if child.data.tt == n.tt && child.data.tv == n.tv {
			return &p.childs[i]
		}
	}
	p.childs = append(p.childs, _InferNode{
		data: n,
	})
	l := len(p.childs)
	return &p.childs[l-1]
}

func _parseRewriteVar(b *Block, line []*Token, ln int) error {
	t1, t2 := line[0], line[2]

	t1.typ = _varname_t
	t1._b = b
	t1.ln = ln

	t2._b = b
	t2.ln = ln
	if err := t2.extractVar(); err != nil {
		return err
	}

	name := t1.String()
	if v, _ := b.GetVar(name); v == nil {
		return WrapErrorf(ErrVariableNotDefined, "variable name '%s'", name)
	}
	stm := newstm("rewrite_var").Append(t1).Append(t2)
	// if err := b.rewriteVar(stm); err != nil {
	// 	return err
	// }
	return b.bbody.Append(stm)
}

func _parseRewriteVarWithExp(b *Block, line []*Token, ln int) error {
	t1 := line[0]

	t1.typ = _varname_t
	t1._b = b
	t1.ln = ln

	var builder strings.Builder
	for _, t := range line[2:] {
		builder.WriteString(t.String())
	}
	exp := builder.String()
	t2 := &Token{
		str: exp,
		typ: _exp_t,
		ln:  ln,
		_b:  b,
	}
	if err := t2.extractVar(); err != nil {
		return err
	}

	name := t1.String()
	if v, _ := b.GetVar(name); v == nil {
		return WrapErrorf(ErrVariableNotDefined, "variable name '%s'", name)
	}
	stm := newstm("rewrite_var").Append(t1).Append(t2)
	//if err := b.rewriteVar(stm); err != nil {
	//	return err
	//}
	return b.bbody.Append(stm)
}
