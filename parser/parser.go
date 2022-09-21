//go:generate stringer -type aststate
package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/cofxlabs/cofx/pkg/enabled"
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
	_ast_if_body
	_ast_switch_body
	_ast_case_body
	_ast_default_body
	_ast_event_body
)

var statementPatterns = map[string]struct {
	min     int
	max     int
	types   []TokenType
	values  []string
	uptypes []TokenType
	newbody func() body
}{
	"load": {
		2, 2,
		[]TokenType{_ident_t, _string_t},
		[]string{_kw_load, ""},
		[]TokenType{_keyword_t, _load_t},
		nil,
	},
	"fn": {
		5, 5,
		[]TokenType{_ident_t, _ident_t, _symbol_t, _ident_t, _symbol_t},
		[]string{_kw_fn, "", "=", "", "{"},
		[]TokenType{_keyword_t, _functionname_t, _operator_t, _functionname_t, _symbol_t},
		func() body { return &plainbody{} },
	},
	"co1": {
		2, 2,
		[]TokenType{_ident_t, _ident_t},
		[]string{_kw_co, ""},
		[]TokenType{_keyword_t, _functionname_t},
		nil,
	},
	"co1->": {
		4, 4,
		[]TokenType{_ident_t, _ident_t, _symbol_t, _ident_t},
		[]string{_kw_co, "", "->", ""},
		[]TokenType{_keyword_t, _functionname_t, _operator_t, _varname_t},
		nil,
	},
	"co1+": {
		3, 3,
		[]TokenType{_ident_t, _ident_t, _symbol_t},
		[]string{_kw_co, "", "{"},
		[]TokenType{_keyword_t, _functionname_t, _symbol_t},
		func() body { return &MapBody{} },
	},
	"co1+->": {
		5, 5,
		[]TokenType{_ident_t, _ident_t, _symbol_t, _ident_t, _symbol_t},
		[]string{_kw_co, "", "->", "", "{"},
		[]TokenType{_keyword_t, _functionname_t, _operator_t, _varname_t, _symbol_t},
		func() body { return &MapBody{} },
	},
	"co2": {
		2, 2,
		[]TokenType{_ident_t, _symbol_t},
		[]string{_kw_co, "{"},
		[]TokenType{_keyword_t, _symbol_t},
		func() body { return &ListBody{etype: _functionname_t} },
	},
	"var": {
		2, 4,
		[]TokenType{_ident_t, _ident_t, _symbol_t},
		[]string{_kw_var, "", "="},
		[]TokenType{_keyword_t, _varname_t, _operator_t},
		nil,
	},
	"args": {
		3, 3,
		[]TokenType{_ident_t, _symbol_t, _symbol_t},
		[]string{_kw_args, "=", "{"},
		[]TokenType{_keyword_t, _operator_t, _symbol_t},
		func() body { return &MapBody{} },
	},
	"for1": {
		2, 2,
		[]TokenType{_ident_t, _symbol_t},
		[]string{_kw_for, "{"},
		[]TokenType{_keyword_t, _symbol_t},
		func() body { return &plainbody{} },
	},
	"for2": {
		3, 3,
		[]TokenType{_ident_t, _expr_t, _symbol_t},
		[]string{_kw_for, "", "{"},
		[]TokenType{_keyword_t, _expr_t, _symbol_t},
		func() body { return &plainbody{} },
	},
	"if": {
		3, 3,
		[]TokenType{_ident_t, _expr_t, _symbol_t},
		[]string{_kw_if, "", "{"},
		[]TokenType{_keyword_t, _expr_t, _symbol_t},
		func() body { return &plainbody{} },
	},
	"switch": {
		2, 2,
		[]TokenType{_ident_t, _symbol_t},
		[]string{_kw_switch, "{"},
		[]TokenType{_keyword_t, _symbol_t},
		func() body { return &plainbody{} },
	},
	"case": {
		3, 3,
		[]TokenType{_ident_t, _expr_t, _symbol_t},
		[]string{_kw_case, "", "{"},
		[]TokenType{_keyword_t, _expr_t, _symbol_t},
		func() body { return &plainbody{} },
	},
	"default": {
		2, 2,
		[]TokenType{_ident_t, _symbol_t},
		[]string{_kw_default, "{"},
		[]TokenType{_keyword_t, _symbol_t},
		func() body { return &plainbody{} },
	},
	"closed": {
		1, 1,
		[]TokenType{_symbol_t},
		[]string{"}"},
		[]TokenType{_symbol_t},
		nil,
	},
	"event": {
		2, 2,
		[]TokenType{_ident_t, _symbol_t},
		[]string{_kw_event, "{"},
		[]TokenType{_keyword_t, _symbol_t},
		func() body { return &plainbody{} },
	},
	"builtins": {
		1, 3,
		[]TokenType{_ident_t, _string_t, _string_t},
		[]string{"", "", ""},
		[]TokenType{_ident_t, _string_t, _string_t},
		nil,
	},
}

var statementInferRules map[string][]inferData = map[string][]inferData{
	"rewrite_var1":     {{_ident_t, ""}, {_symbol_t, "<-"}, {_string_t, ""}},                   // e.g.: v <- "foo"
	"rewrite_var2":     {{_ident_t, ""}, {_symbol_t, "<-"}, {_number_t, ""}},                   // e.g.: v <- 100
	"rewrite_var3":     {{_ident_t, ""}, {_symbol_t, "<-"}, {_refvar_t, ""}},                   // e.g.: v <- $(foo)
	"rewrite_var_exp1": {{_ident_t, ""}, {_symbol_t, "<-"}, {_symbol_t, "-"}, {_number_t, ""}}, // e.g.: v <- -100
	"rewrite_var_exp2": {{_ident_t, ""}, {_symbol_t, "<-"}, {_symbol_t, "("}},                  // e.g.: v <- (1 + 2) * 3
	"rewrite_var_exp3": {{_ident_t, ""}, {_symbol_t, "<-"}, {_string_t, ""}, {_symbol_t, ""}},  // e.g.: v <- "a" > "b"
	"rewrite_var_exp4": {{_ident_t, ""}, {_symbol_t, "<-"}, {_number_t, ""}, {_symbol_t, ""}},  // e.g.: v <- 100 + 1
	"rewrite_var_exp5": {{_ident_t, ""}, {_symbol_t, "<-"}, {_refvar_t, ""}, {_symbol_t, ""}},  // e.g.: v <- $(foo) + 1
}

// AST store all blocks in the flowl
type AST struct {
	// global is the root block of the AST, we can access the whole AST tree through global block
	global Block
	// We use the first line comment in the flowl file as the description of the flow
	desc string

	// for parsing
	_InferTree *inferNode
	_FA
	// for validating
	fns map[string]bool
	cos []string
}

func New(rd io.Reader) (*AST, error) {
	lx := newLexer()
	buff := bufio.NewReader(rd)
	for n := 1; ; n += 1 {
		line, err := buff.ReadString('\n')
		if err == io.EOF {
			if len(line) != 0 {
				if err := lx.split(line, n, true); err != nil {
					return nil, err
				}
			}
			break
		}
		if err != nil {
			return nil, err
		}
		if err := lx.split(line, n, false); err != nil {
			return nil, err
		}
	}

	lx.debug()

	ast := newast()
	if err := ast.scan(lx); err != nil {
		return nil, err
	}

	if enabled.Debug() {
		ast.Foreach(func(b *Block) error {
			b.Debug()
			return nil
		})
	}

	return ast, ast.validate()
}

func newast() *AST {
	ast := &AST{
		global: Block{
			kind: Token{
				str: "global",
			},
			vtbl: vartable{vars: map[string]*_var{"env": newEnvVar()}},
			body: &plainbody{},
		},
		_FA: _FA{
			state: _ast_global,
		},
		_InferTree: buildInferTree(),
		fns:        make(map[string]bool),
	}
	return ast
}

func (ast *AST) Global() *Block {
	return &ast.global
}

func (ast *AST) Desc() string {
	return ast.desc
}

func (ast *AST) GetBlocks() (loads []*Block, fns []*Block, runs []*Block) {
	ast.Foreach(func(b *Block) error {
		if b.IsLoad() {
			loads = append(loads, b)
		}
		if b.IsFn() {
			fns = append(fns, b)
		}
		_, isbuiltin := b.IsBuiltinDirective()
		if b.IsFor() || b.IsBtf() || b.IsCo() || isbuiltin {
			runs = append(runs, b)
		}
		return nil
	})
	return
}

func (ast *AST) Foreach(do func(*Block) error) error {
	return deepwalk(&ast.global, do)
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

func (ast *AST) scan(lx *lexer) error {
	var parsingblock = &ast.global

	err := lx.foreachLine(func(ln int, line []*Token) error {
		if len(line) == 0 {
			return nil
		}
		// the line is a commment
		if line[0].String() == _kw_comment {
			// save the first line comment as the description of the flow
			if ast.desc == "" && len(ast.global.child) == 0 && len(line) > 1 {
				ast.desc = line[1].String()
			}
			// discard the other line comments
			return nil
		}

		switch ast.phase() {
		case _ast_global:
			kind := line[0]
			switch kind.String() {
			case _kw_load:
				return ast.parseLoad(line, ln, parsingblock)
			case _kw_fn:
				block, err := ast.parseFn(line, ln, parsingblock)
				if err != nil {
					return err
				}
				parsingblock = block
				ast._goto(_ast_fn_body)
			case _kw_co:
				block, err := ast.parseCo(line, ln, parsingblock)
				if err != nil {
					return err
				}
				if block.body != nil {
					parsingblock = block
					ast._goto(_ast_co_body)
				}
			case _kw_var:
				return ast.parseVar(line, ln, parsingblock)
			case _kw_for:
				block, err := ast.parseFor(line, ln, parsingblock)
				if err != nil {
					return err
				}
				parsingblock = block
				ast._goto(_ast_for_body)
			case _kw_if:
				block, err := ast.parseIf(line, ln, parsingblock)
				if err != nil {
					return err
				}
				parsingblock = block
				ast._goto(_ast_if_body)
			case _kw_switch:
				block, err := ast.parseSwitch(line, ln, parsingblock)
				if err != nil {
					return err
				}
				parsingblock = block
				ast._goto(_ast_switch_body)
			case _kw_event:
				block, err := ast.parseEvent(line, ln, parsingblock)
				if err != nil {
					return err
				}
				parsingblock = block
				ast._goto(_ast_event_body)
			case _di_exit, _di_sleep, _di_println, _di_if_none_exit:
				if err := ast.parseBuiltDirective(line, ln, parsingblock); err != nil {
					return err
				}
			default:
				if _parse, err := ast._InferTree.lookup(line); err == nil {
					if err := _parse(parsingblock, line, ln); err != nil {
						return statementTokensErrorf(err, line)
					}
					return nil
				}
				return statementTokensErrorf(ErrStatementUnknow, line)
			}
		case _ast_fn_body:
			block, err := ast.parseFnBody(line, ln, parsingblock)
			if err != nil {
				return err
			}
			if block == nil {
				panic("block is nil")
			}
			parsingblock = block
		case _ast_args_body:
			block, err := ast.parseArgsBody(line, ln, parsingblock)
			if err != nil {
				return err
			}
			if block == nil {
				panic("block is nil")
			}
			parsingblock = block
		case _ast_co_body:
			block, err := ast.parseCoBody(line, ln, parsingblock)
			if err != nil {
				return err
			}
			if block == nil {
				panic("block is nil")
			}
			parsingblock = block
		case _ast_for_body:
			block, err := ast.parseForBody(line, ln, parsingblock)
			if err != nil {
				return err
			}
			if block == nil {
				panic("block is nil")
			}
			parsingblock = block
		case _ast_if_body:
			block, err := ast.parseIfBody(line, ln, parsingblock)
			if err != nil {
				return err
			}
			if block == nil {
				panic("block is nil")
			}
			parsingblock = block
		case _ast_switch_body:
			block, err := ast.parseSwitchBody(line, ln, parsingblock)
			if err != nil {
				return err
			}
			if block == nil {
				panic("block is nil")
			}
			parsingblock = block
		case _ast_case_body, _ast_default_body:
			block, err := ast.parseCaseDeafultBody(line, ln, parsingblock)
			if err != nil {
				return err
			}
			if block == nil {
				panic("block is nil")
			}
			parsingblock = block
		case _ast_event_body:
			block, err := ast.parseEventBody(line, ln, parsingblock)
			if err != nil {
				return err
			}
			if block == nil {
				panic("block is nil")
			}
			parsingblock = block
		}
		return nil
	})
	if err != nil {
		return err
	}
	if ast.phase() != _ast_global {
		return errors.New("incomplete source file, possible missing terminator")
	}
	return nil
}

func (ast *AST) validate() error {
	for _, s := range ast.cos {
		if ok, found := ast.fns[s]; !found {
			continue
		} else {
			if ok {
				ok = false
				continue
			}
			return wrapErrorf(ErrIdentConflict, "duplicate calling the fn '%s'", s)
		}
	}
	ast.cos = nil
	ast.fns = nil

	return ast.Foreach(func(b *Block) error {
		if err := b.validate(); err != nil {
			return err
		}
		if err := b.vtbl.cyclecheck(); err != nil {
			return err
		}
		return nil
	})
}

func (ast *AST) preparse(k string, line []*Token, ln int, b *Block) (body, error) {
	pattern := statementPatterns[k]

	if l := len(line); l < pattern.min || l > pattern.max {
		return nil, tokenErrorf(ln, ErrTokenNumInLine, "actual %d, expect [%d,%d]", l, pattern.min, pattern.max)
	}

	min := len(line)
	if l := len(pattern.types); min > l {
		min = l
	}

	for i := 0; i < min; i++ {
		t := line[i]
		expectTyp := pattern.types[i]
		expectVal := pattern.values[i]

		if !t.TypeEqual(expectTyp) {
			return nil, tokenTypeErrorf(t, expectTyp)
		}
		if expectVal != "" && expectVal != t.String() {
			return nil, tokenValueErrorf(t, expectVal)
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

	var body body
	if pattern.newbody != nil {
		body = pattern.newbody()
	}

	return body, nil
}

func (ast *AST) parseVar(line []*Token, ln int, current *Block) error {
	var composed []*Token
	if l := len(line); l > 4 {
		composed = append(composed, line[0:3]...)
		// Compose all intermediate tokens to expresssion
		composed = append(composed, newExpression(line[3:]).ToToken())
	} else {
		composed = line
	}
	if _, err := ast.preparse("var", composed, ln, current); err != nil {
		return err
	}

	var (
		name = composed[1]
		val  *Token
	)
	if len(composed) == 4 {
		// e.g.:
		// 		var v = "foo"
		// 		var v = 100
		// 		var v = $(a)

		// 		var v = 1 + 1
		// 		var v = -1
		// 		var v = 1 + $(foo)
		// 		var v = "a" > "b"
		// the value is a expression
		val = composed[3]
		if !val.TypeEqual(_string_t, _number_t, _refvar_t, _expr_t) {
			return varErrorf(val.ln, ErrVariableValueType, "variable '%s' value '%s' type '%s'", name, val.String(), val.typ)
		}
	}

	var stm *Statement
	if val != nil {
		stm = NewStatement("var").Append(name).Append(val)
	} else {
		stm = NewStatement("var").Append(name)
	}
	if err := current.initVar(stm); err != nil {
		return err
	}
	return nil
}

func (ast *AST) parseLoad(line []*Token, ln int, parent *Block) error {
	b := &Block{
		child:  []*Block{},
		parent: parent,
		vtbl:   vartable{vars: make(map[string]*_var)},
	}
	body, err := ast.preparse("load", line, ln, b)
	if err != nil {
		return err
	}
	b.body = body
	b.kind = *line[0]
	b.target1 = *line[1]

	parent.child = append(parent.child, b)
	return nil
}

func (ast *AST) parseFn(line []*Token, ln int, parent *Block) (*Block, error) {
	b := &Block{
		child:  []*Block{},
		parent: parent,
		vtbl:   vartable{vars: make(map[string]*_var)},
	}
	body, err := ast.preparse("fn", line, ln, b)
	if err != nil {
		return nil, err
	}
	b.body = body

	kind, target, op, tv := line[0], line[1], line[2], line[3]

	b.kind = *kind
	b.target1 = *target
	b.operator = *op
	b.target2 = *tv

	// validate grammar
	if b.Target1().StringEqual(&b.target2) {
		s1 := b.target1.String()
		s2 := b.target2.String()
		return nil, parseErrorf(ln, ErrIdentConflict, "'%s', '%s'", s1, s2)
	}

	if s, ok := ast.fns[b.Target1().String()]; ok {
		return nil, parseErrorf(ln, ErrIdentConflict, "duplicate definition of fn '%s'", s)
	} else {
		ast.fns[b.Target1().String()] = true
	}
	parent.child = append(parent.child, b)
	return b, nil
}

func (ast *AST) parseFnBody(line []*Token, ln int, current *Block) (*Block, error) {
	if _, err := ast.preparse("closed", line, ln, current); err == nil {
		ast._goto(_ast_global)
		return &ast.global, nil
	}

	kind := line[0]
	switch kind.String() {
	case _kw_args:
		block, err := ast.parseArgs(line, ln, current)
		if err != nil {
			return nil, err
		}
		ast._goto(_ast_args_body)
		return block, nil
	case _kw_var:
		if err := ast.parseVar(line, ln, current); err != nil {
			return nil, err
		}
	default:
		if _parse, err := ast._InferTree.lookup(line); err == nil {
			if err := _parse(current, line, ln); err != nil {
				return nil, statementTokensErrorf(err, line)
			}
			return current, nil
		}
		return nil, statementErrorf(ln, ErrStatementUnknow, "%s", kind)
	}
	return current, nil
}

func (ast *AST) parseCo(line []*Token, ln int, parent *Block) (*Block, error) {
	b := &Block{
		parent: parent,
		vtbl:   vartable{vars: make(map[string]*_var)},
	}

	var (
		body body
		err  error
	)
	keys := []string{"co1", "co1+", "co2", "co1->", "co1+->"}
	for _, k := range keys {
		body, err = ast.preparse(k, line, ln, b)
		if err == nil {
			b.kind = *line[0]
			b.body = body
			switch k {
			case "co1": // co sleep
				b.target1 = *line[1]
			case "co1+": // co sleep {
				b.target1 = *line[1]
			case "co1->": // co sleep -> out
				b.target1 = *line[1]
				b.operator = *line[2]
				b.target2 = *line[3]
			case "co1+->": // co sleep -> out {
				b.target1 = *line[1]
				b.operator = *line[2]
				b.target2 = *line[3]
			case "co2": // co {
			}
			break
		}
	}
	if err != nil {
		return nil, err
	}

	// check grammar
	if b.Target1().StringEqual(&b.target2) {
		s1 := b.target1.String()
		s2 := b.target2.String()
		return nil, parseErrorf(ln, ErrIdentConflict, "'%s','%s'", s1, s2)
	}

	// check return value variable
	if !b.target2.IsEmpty() {
		name := b.target2.String()
		if v, _ := b.getVar(name); v == nil {
			return nil, varErrorf(b.target2.ln, ErrVariableNotDefined, "'%s'", name)
		}
	}

	// when co is in switch/if, add the condition var statement
	if b.IsCo() && (b.InSwitch() || b.InIf()) {
		stm := NewStatement("var").Append(b.Parent().Target1()).Append(b.Parent().Target2())
		if err := b.initVar(stm); err != nil {
			return nil, err
		}
	}

	if !b.Target1().IsEmpty() {
		ast.cos = append(ast.cos, b.Target1().String())
	}

	parent.child = append(parent.child, b)
	return b, nil
}

func (ast *AST) parseCoBody(line []*Token, ln int, current *Block) (*Block, error) {
	if _, err := ast.preparse("closed", line, ln, current); err == nil {
		parent := current.parent
		if parent.IsFor() {
			ast._goto(_ast_for_body)
		} else if parent.IsIf() {
			ast._goto(_ast_if_body)
		} else if parent.IsCase() {
			ast._goto(_ast_case_body)
		} else if parent.IsDefault() {
			ast._goto(_ast_default_body)
		} else if parent.IsEvent() {
			ast._goto(_ast_event_body)
		} else {
			ast._goto(_ast_global)
		}
		return parent, nil
	}

	for _, t := range line {
		t.ln = ln
		t._b = current
		if err := t.extractVar(); err != nil {
			return nil, statementTokensErrorf(err, line)
		}

		if !t.IsEmpty() {
			ast.cos = append(ast.cos, t.String())
		}
	}
	if err := current.body.Append(line); err != nil {
		return nil, err
	}
	return current, nil
}

func (ast *AST) parseArgs(line []*Token, ln int, parent *Block) (*Block, error) {
	b := &Block{
		child:  []*Block{},
		parent: parent,
		vtbl:   vartable{vars: make(map[string]*_var)},
	}
	body, err := ast.preparse("args", line, ln, b)
	if err != nil {
		return nil, err
	}
	b.body = body
	b.kind = *line[0]

	parent.child = append(parent.child, b)
	return b, nil
}

func (ast *AST) parseArgsBody(line []*Token, ln int, current *Block) (*Block, error) {
	if _, err := ast.preparse("closed", line, ln, current); err == nil {
		parent := current.parent
		ast._goto(_ast_fn_body)
		return parent, nil
	}
	for _, t := range line {
		t.ln = ln
		t._b = current
		if err := t.extractVar(); err != nil {
			return nil, statementTokensErrorf(err, line)
		}
	}
	if err := current.body.Append(line); err != nil {
		return nil, err
	}
	return current, nil
}

func (ast *AST) parseFor(line []*Token, ln int, parent *Block) (*Block, error) {
	var composed []*Token
	l := len(line)
	if l > 2 {
		// first
		composed = append(composed, line[0])
		// Compose all intermediate tokens into a condition expression
		composed = append(composed, newExpression(line[1:l-1]).ToToken())
		// last
		composed = append(composed, line[l-1])
	} else {
		composed = line
	}

	b := &Block{
		child:  []*Block{},
		parent: parent,
		vtbl:   vartable{vars: make(map[string]*_var)},
	}
	if len(composed) == 2 {
		body, err := ast.preparse("for1", composed, ln, b)
		if err != nil {
			return nil, err
		}
		b.body = body
		b.kind = *composed[0]
	} else {
		body, err := ast.preparse("for2", composed, ln, b)
		if err != nil {
			return nil, err
		}
		b.body = body
		b.kind = *composed[0]
		b.target1 = Token{
			ln:  ln,
			_b:  b,
			str: _condition_expr_var,
			typ: _varname_t,
		}
		b.target2 = *composed[1]

		// add the condition var statement
		if !b.Target1().IsEmpty() && !b.Target2().IsEmpty() {
			stm := NewStatement("var").Append(b.Target1()).Append(b.Target2())
			if err := b.initVar(stm); err != nil {
				return nil, err
			}
		}
	}

	parent.child = append(parent.child, b)
	return b, nil
}

func (ast *AST) parseForBody(line []*Token, ln int, current *Block) (*Block, error) {
	if _, err := ast.preparse("closed", line, ln, current); err == nil {
		// add a 'btf' block into the child of 'for', it represents the end of the loop
		btf := &Block{
			kind: Token{
				str: "btf",
				typ: _keyword_t,
			},
			parent: current,
		}
		current.child = append(current.child, btf)

		// back to global
		ast._goto(_ast_global)
		return current.parent, nil
	}

	kind := line[0]
	switch kind.String() {
	case _kw_co:
		block, err := ast.parseCo(line, ln, current)
		if err != nil {
			return nil, err
		}
		if block.body != nil {
			ast._goto(_ast_co_body)
			return block, nil
		}
	case _kw_if:
		block, err := ast.parseIf(line, ln, current)
		if err != nil {
			return nil, err
		}
		ast._goto(_ast_if_body)
		return block, nil
	case _kw_switch:
		block, err := ast.parseSwitch(line, ln, current)
		if err != nil {
			return nil, err
		}
		ast._goto(_ast_switch_body)
		return block, nil
	case _di_exit, _di_sleep, _di_println, _di_if_none_exit:
		if err := ast.parseBuiltDirective(line, ln, current); err != nil {
			return nil, err
		}
	default:
		if _parse, err := ast._InferTree.lookup(line); err == nil {
			if err := _parse(current, line, ln); err != nil {
				return nil, statementTokensErrorf(err, line)
			}
			return current, nil
		}
		return nil, statementErrorf(ln, ErrStatementUnknow, "%s", kind)
	}
	return current, nil
}

func (ast *AST) parseIf(line []*Token, ln int, parent *Block) (*Block, error) {
	var (
		composed []*Token
	)
	if l := len(line); l > 3 {
		// first
		composed = append(composed, line[0])
		// Compose all intermediate tokens to expression
		composed = append(composed, newExpression(line[1:l-1]).ToToken())
		// last
		composed = append(composed, line[l-1])
	} else {
		composed = line
	}

	b := &Block{
		child:  []*Block{},
		parent: parent,
		vtbl:   vartable{vars: make(map[string]*_var)},
	}
	body, err := ast.preparse("if", composed, ln, b)
	if err != nil {
		return nil, err
	}
	b.body = body
	b.kind = *composed[0]
	b.target1 = Token{
		ln:  ln,
		_b:  b,
		str: _condition_expr_var,
		typ: _varname_t,
	}
	b.target2 = *composed[1]

	parent.child = append(parent.child, b)
	return b, nil
}

func (ast *AST) parseIfBody(line []*Token, ln int, current *Block) (*Block, error) {
	if _, err := ast.preparse("closed", line, ln, current); err == nil {
		parent := current.parent
		if parent.IsFor() {
			ast._goto(_ast_for_body)
		} else {
			ast._goto(_ast_global)
		}
		return parent, nil
	}

	kind := line[0]
	switch kind.String() {
	case _kw_co:
		block, err := ast.parseCo(line, ln, current)
		if err != nil {
			return nil, err
		}
		if block.body != nil {
			ast._goto(_ast_co_body)
			return block, nil
		}
	case _di_exit, _di_sleep, _di_println, _di_if_none_exit:
		if err := ast.parseBuiltDirective(line, ln, current); err != nil {
			return nil, err
		}
	default:
		return nil, statementErrorf(ln, ErrStatementUnknow, "%s", kind)
	}
	return current, nil
}

func (ast *AST) parseSwitch(line []*Token, ln int, parent *Block) (*Block, error) {
	b := &Block{
		child:  []*Block{},
		parent: parent,
		vtbl:   vartable{vars: make(map[string]*_var)},
	}
	body, err := ast.preparse("switch", line, ln, b)
	if err != nil {
		return nil, err
	}
	b.body = body
	b.kind = *line[0]

	parent.child = append(parent.child, b)
	return b, nil
}

func (ast *AST) parseSwitchBody(line []*Token, ln int, current *Block) (*Block, error) {
	if _, err := ast.preparse("closed", line, ln, current); err == nil {
		parent := current.parent
		if parent.IsFor() {
			ast._goto(_ast_for_body)
		} else {
			ast._goto(_ast_global)
		}
		return parent, nil
	}

	kind := line[0]
	switch kind.String() {
	case _kw_case:
		block, err := ast.parseCase(line, ln, current)
		if err != nil {
			return nil, err
		}
		ast._goto(_ast_case_body)
		return block, nil
	case _kw_default:
		block, err := ast.parseDefault(line, ln, current)
		if err != nil {
			return nil, err
		}
		ast._goto(_ast_default_body)
		return block, nil
	default:
		return nil, statementErrorf(ln, ErrStatementUnknow, "%s", kind)
	}
}

func (ast *AST) parseCase(line []*Token, ln int, parent *Block) (*Block, error) {
	var composed []*Token
	if l := len(line); l > 3 {
		// first
		composed = append(composed, line[0])
		// Compose all intermediate tokens to expresssion
		composed = append(composed, newExpression(line[1:l-1]).ToToken())
		// last
		composed = append(composed, line[l-1])
	} else {
		composed = line
	}

	b := &Block{
		child:  []*Block{},
		parent: parent,
		vtbl:   vartable{vars: make(map[string]*_var)},
	}
	body, err := ast.preparse("case", composed, ln, b)
	if err != nil {
		return nil, err
	}
	b.body = body
	b.kind = *composed[0]
	b.target1 = Token{
		ln:  ln,
		_b:  b,
		str: _condition_expr_var,
		typ: _varname_t,
	}
	b.target2 = *composed[1]

	parent.child = append(parent.child, b)
	return b, nil
}

func (ast *AST) parseDefault(line []*Token, ln int, parent *Block) (*Block, error) {
	// Only one defaul statement inside a switch, so check it
	for _, c := range parent.child {
		if c.IsDefault() {
			return nil, statementErrorf(ln, ErrStatementTooMany, "default in swith")
		}
	}

	b := &Block{
		child:  []*Block{},
		parent: parent,
		vtbl:   vartable{vars: make(map[string]*_var)},
	}
	body, err := ast.preparse("default", line, ln, b)
	if err != nil {
		return nil, err
	}
	b.body = body
	b.kind = *line[0]

	b.target1 = Token{
		ln:  ln,
		_b:  b,
		str: _condition_expr_var,
		typ: _varname_t,
	}

	// generate condition expression of the 'default'
	var cases []string
	swb := parent
	for _, c := range swb.child {
		if c.IsCase() {
			cases = append(cases, fmt.Sprintf("(!(%s))", c.target2.String()))
		}
	}
	b.target2 = Token{
		ln:  ln,
		_b:  b,
		str: strings.Join(cases, "&&"),
		typ: _expr_t,
	}
	if err := b.target2.extractVar(); err != nil {
		return nil, err
	}

	parent.child = append(parent.child, b)
	return b, nil
}

func (ast *AST) parseCaseDeafultBody(line []*Token, ln int, current *Block) (*Block, error) {
	if _, err := ast.preparse("closed", line, ln, current); err == nil {
		parent := current.parent
		ast._goto(_ast_switch_body)
		return parent, nil
	}

	kind := line[0]
	switch kind.String() {
	case _kw_co:
		block, err := ast.parseCo(line, ln, current)
		if err != nil {
			return nil, err
		}
		if block.body != nil {
			ast._goto(_ast_co_body)
			return block, nil
		}
	case _di_exit, _di_sleep, _di_println, _di_if_none_exit:
		if err := ast.parseBuiltDirective(line, ln, current); err != nil {
			return nil, err
		}
	default:
		return nil, statementErrorf(ln, ErrStatementUnknow, "%s", kind)
	}
	return current, nil
}

func (ast *AST) parseEvent(line []*Token, ln int, parent *Block) (*Block, error) {
	b := &Block{
		parent: parent,
		vtbl:   vartable{vars: make(map[string]*_var)},
	}
	body, err := ast.preparse("event", line, ln, b)
	if err != nil {
		return nil, err
	}
	b.body = body
	b.kind = *line[0]

	parent.child = append(parent.child, b)
	return b, nil
}

func (ast *AST) parseEventBody(line []*Token, ln int, current *Block) (*Block, error) {
	if _, err := ast.preparse("closed", line, ln, current); err == nil {
		ast._goto(_ast_global)
		return &ast.global, nil
	}

	kind := line[0]
	switch kind.String() {
	case _kw_co:
		block, err := ast.parseCo(line, ln, current)
		if err != nil {
			return nil, err
		}
		// need to goto and parse the body of 'co'
		if block.body != nil {
			ast._goto(_ast_co_body)
			return block, nil
		}
	default:
		return nil, statementErrorf(ln, ErrStatementUnknow, "%s", kind)
	}
	// don't goto, keep parsing the body of 'event'
	return current, nil
}

func (ast *AST) parseBuiltDirective(line []*Token, ln int, parent *Block) error {
	b := &Block{
		parent: parent,
		vtbl:   vartable{vars: make(map[string]*_var)},
	}
	body, err := ast.preparse("builtins", line, ln, b)
	if err != nil {
		return err
	}
	b.body = body
	b.kind = *line[0]
	if len(line) == 2 {
		b.target1 = *line[1]
	}
	if len(line) == 3 {
		b.target1 = *line[1]
		b.target2 = *line[2]
	}

	// when in switch/if, add the condition var statement
	if b.InSwitch() || b.InIf() {
		stm := NewStatement("var").Append(b.Parent().Target1()).Append(b.Parent().Target2())
		if err := b.initVar(stm); err != nil {
			return err
		}
	}

	parent.child = append(parent.child, b)
	return nil
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

type inferNode struct {
	data   inferData
	childs []inferNode
	parse  func(*Block, []*Token, int) error
}

type inferData struct {
	tt TokenType
	tv string
}

func buildInferTree() *inferNode {
	root := &inferNode{}
	p := root
	for k, rule := range statementInferRules {
		for _, e := range rule {
			p = p.insert(e)
		}
		if strings.HasPrefix(k, "rewrite_var_exp") {
			p.parse = parseRewriteVarWithExp
		} else if strings.HasPrefix(k, "rewrite_var") {
			p.parse = parseRewriteVar
		}

		p = root
	}
	return root
}

func (in *inferNode) insert(n inferData) *inferNode {
	p := in
	for i, child := range p.childs {
		if child.data.tt == n.tt && child.data.tv == n.tv {
			return &p.childs[i]
		}
	}
	p.childs = append(p.childs, inferNode{
		data: n,
	})
	l := len(p.childs)
	return &p.childs[l-1]
}

func (in *inferNode) lookup(line []*Token) (func(*Block, []*Token, int) error, error) {
	var found = false
	p := in
	for _, t := range line {
		if len(p.childs) == 0 && p.parse != nil {
			return p.parse, nil
		}

		for i, child := range p.childs {
			if !t.TypeEqual(child.data.tt) {
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
			return nil, statementTokensErrorf(ErrStatementInferFailed, line)
		}
	}
	return p.parse, nil
}

func parseRewriteVar(b *Block, line []*Token, ln int) error {
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
	if v, _ := b.getVar(name); v == nil {
		return wrapErrorf(ErrVariableNotDefined, "variable name '%s'", name)
	}
	stm := NewStatement("rewrite_var").Append(t1).Append(t2)
	// if err := b.rewriteVar(stm); err != nil {
	// 	return err
	// }
	return b.body.Append(stm)
}

func parseRewriteVarWithExp(b *Block, line []*Token, ln int) error {
	t1 := line[0]
	t1.typ = _varname_t
	t1.ln = ln
	t1._b = b

	t2 := newExpression(line[2:]).ToToken()
	t2.ln = ln
	t2._b = b
	if err := t2.extractVar(); err != nil {
		return err
	}

	name := t1.String()
	if v, _ := b.getVar(name); v == nil {
		return wrapErrorf(ErrVariableNotDefined, "variable name '%s'", name)
	}
	stm := NewStatement("rewrite_var").Append(t1).Append(t2)
	//if err := b.rewriteVar(stm); err != nil {
	//	return err
	//}
	return b.body.Append(stm)
}
