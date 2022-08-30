//go:generate stringer -type TokenType
package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cofunclabs/cofunc/pkg/is"
)

const (
	_kw_comment = "//"
	_kw_load    = "load"
	_kw_fn      = "fn"
	_kw_co      = "co"
	_kw_var     = "var"
	_kw_args    = "args"
	_kw_for     = "for"
	_kw_if      = "if"
	_kw_switch  = "switch"
	_kw_case    = "case"
	_kw_default = "default"
	_kw_event   = "event"
)

var keywordTable = map[string]struct{}{
	_kw_args:    {},
	_kw_case:    {},
	_kw_co:      {},
	_kw_comment: {},
	_kw_default: {},
	_kw_fn:      {},
	_kw_for:     {},
	_kw_if:      {},
	_kw_load:    {},
	_kw_switch:  {},
	_kw_var:     {},
	_kw_event:   {},
}

func iskeyword(ss ...string) (string, bool) {
	for _, s := range ss {
		_, ok := keywordTable[s]
		if ok {
			return s, true
		}
	}
	return "", false
}

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

type TokenType int

var tokenPatterns = map[TokenType]*regexp.Regexp{
	_unknow_t:       regexp.MustCompile(`^*$`),
	_string_t:       regexp.MustCompile(`^*$`),
	_refvar_t:       regexp.MustCompile(`^\$\([a-zA-Z0-9_\.]*\)$`),
	_ident_t:        regexp.MustCompile(`^[a-zA-Z0-9_\.]*$`),
	_number_t:       regexp.MustCompile(`^[0-9\.]+$`),
	_mapkey_t:       regexp.MustCompile(`^[^:]+$`), // not contain ":"
	_operator_t:     regexp.MustCompile(`^(=|->)$`),
	_load_t:         regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*:.*[a-zA-Z0-9]$`),
	_functionname_t: regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`),
	_keyword_t:      regexp.MustCompile(`^[a-z]*$`),
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

func (t *Token) IsEmpty() bool {
	return len(t.str) == 0
}

func (t *Token) StringEqual(t1 *Token) bool {
	if !t.IsEmpty() && !t1.IsEmpty() {
		if t.String() == t1.String() {
			return true
		}
	}
	return false
}

func (t *Token) TypeEqual(ts ...TokenType) bool {
	for _, t1 := range ts {
		if t.typ == t1 {
			return true
		}
	}
	return false
}

func (t *Token) String() string {
	return t.str
}

func (t *Token) FormatString() string {
	return fmt.Sprintf("['%s','%s']", t.str, t.typ)
}

func _lookupVar(b *Block, name string) (string, bool) {
	return b.calcVar(name)
}

func (t *Token) copySegments() []struct {
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

func (t *Token) validate() error {
	if pattern, ok := tokenPatterns[t.typ]; ok {
		if !pattern.MatchString(t.str) {
			return tokenErrorf(t.ln, ErrTokenRegex, "actual '%s', expect '%s'", t, pattern)
		}
	}

	if t.TypeEqual(_functionname_t, _varname_t, _ident_t) {
		if s, ok := iskeyword(t.String()); ok {
			return tokenErrorf(t.ln, ErrIsKeyword, "'%s'", s)
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
				return varErrorf(t.ln, ErrVariableFormat, "'%s' in token '%s'", name, t)
			}
			f1, f2 := fields[0], fields[1]
			if f1 == "" || f2 == "" {
				return varErrorf(t.ln, ErrVariableFormat, "'%s' in token '%s'", name, t)
			}
			name = f1
		}
		if v, _ := t._b.getVar(name); v == nil {
			return varErrorf(t.ln, ErrVariableNotDefined, "'%s' in token '%s'", name, t)
		}
	}
	return nil
}

// value will calcuate the variable's value, if the token contain some variables
func (t *Token) value() string {
	if !t.hasVar() {
		return t.str
	}
	if t._get == nil {
		t._get = _lookupVar
	}
	var bd strings.Builder
	for _, seg := range t._segments {
		if seg.isvar {
			val, _ := t._get(t._b, seg.str)
			bd.WriteString(val)
		} else {
			bd.WriteString(seg.str)
		}
	}
	return bd.String()
}

func (t *Token) hasVar() bool {
	for _, seg := range t._segments {
		if seg.isvar {
			return true
		}
	}
	return false
}

func (t *Token) extractVar() error {
	// $(var)
	if !t.TypeEqual(_string_t, _expr_t, _refvar_t) {
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
					return varErrorf(t.ln, ErrVariableNameEmpty, "token '%s'", t)
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
