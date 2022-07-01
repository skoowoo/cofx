//go:generate stringer -type TokenType
package cofunc

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/cofunclabs/cofunc/pkg/is"
	"github.com/pkg/errors"
)

// Token
//
type TokenType int

const (
	_unknow_t TokenType = iota
	_int_t
	_text_t
	_mapkey_t
	_operator_t
	_functionname_t
	_load_t
	_word_t
)

var tokenPatterns = map[TokenType]*regexp.Regexp{
	_unknow_t:       regexp.MustCompile(`^*$`),
	_int_t:          regexp.MustCompile(`^[1-9][0-9]*$`),
	_text_t:         regexp.MustCompile(`^*$`),
	_mapkey_t:       regexp.MustCompile(`^[^:]+$`), // not contain ":"
	_operator_t:     regexp.MustCompile(`^=$`),
	_load_t:         regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*:.*[a-zA-Z0-9]$`),
	_functionname_t: regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_\-]*$`),
	_word_t:         regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_\-]*$`),
}

type Token struct {
	str        string
	persistent string
	typ        TokenType
	b          *Block
	vars       []*struct {
		n    string // var's name
		v    string // var's value, need to read from others
		s, e int    // S is var start position in 'Token.Value', E is end position
	}
	get func(*Block, string) (string, bool)
}

func newToken(s string, typ TokenType) *Token {
	return &Token{
		str: s,
		typ: typ,
	}
}

func getVarFromEnv(name string) (string, bool) {
	return os.Getenv(name), true
}

func (t *Token) String() string {
	return t.str
}

func (t *Token) Value() string {
	if !t.HasVar() {
		return t.str
	}
	if len(t.persistent) != 0 {
		return t.persistent
	}
	var builder strings.Builder
	pe := 0
	cacheable := true
	for _, v := range t.vars {
		var (
			val string
			ok  bool
		)
		if len(v.v) == 0 {
			//var is not cached
			if t.get != nil && v.n != `\` {
				val, ok = t.get(t.b, v.n)
			} else if v.n != `\` {
				val, ok = getVarFromEnv(v.n)
			} else {
				val, ok = "", true
			}
			if ok {
				v.v = val
			} else {
				cacheable = false
			}
		} else {
			// var is cached
			val = v.v
		}
		builder.WriteString(t.str[pe:v.s])
		builder.WriteString(val)
		pe = v.e
	}
	if l := len(t.str); pe < l {
		builder.WriteString(t.str[pe:l])
	}
	s := builder.String()
	if cacheable {
		// cache the token
		t.persistent = s
	}
	return s
}

func (t *Token) IsEmpty() bool {
	return len(t.str) == 0
}

func (t *Token) HasVar() bool {
	return len(t.vars) != 0
}

func (t *Token) extractVar() error {
	// $(var)
	if t.typ != _text_t {
		return nil
	}
	var (
		start int
		state stateL2
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
		case _l2_unknow:
			// skip
			// transfer
			if c == '$' && next(i) == '(' {
				start = i
				state = _l2_word_started
			}
		case _l2_word_started: // from '$'
			// keep
			if is.Word(c) || c == '(' {
				break
			}
			// transfer
			if c == ')' {
				if j := start - 1; j >= 0 {
					if slash := t.str[j]; slash == '\\' {
						// TODO: drop '\'
						state = _l2_unknow
						t.vars = append(t.vars, &struct {
							n string
							v string
							s int
							e int
						}{n: `\`, v: "", s: j, e: j + 1})
						break
					}
				}
				name := t.str[start+2 : i] // start +2: skip "$("
				if name == "" {
					return errors.New("contain invalid var: " + t.str)
				}
				t.vars = append(t.vars, &struct {
					n string
					v string
					s int
					e int
				}{n: name, s: start, e: i + 1}) // currently i is ')'

				state = _l2_unknow
			}
		}
	}
	return nil
}

func (t *Token) validate() error {
	if pattern := tokenPatterns[t.typ]; !pattern.MatchString(t.str) {
		return errors.Errorf("not match: %s:%s", t.str, pattern)
	}
	return nil
}

// Statement
//
type Statement struct {
	lineNum int
	tokens  []*Token
}

func newStatement(ts ...*Token) *Statement {
	stm := &Statement{}
	stm.tokens = append(stm.tokens, ts...)
	return stm
}

func (s *Statement) LastToken() *Token {
	l := len(s.tokens)
	if l == 0 {
		return nil
	}
	return s.tokens[l-1]
}

// Block
//
type blockBody interface {
	Type() string
	Append(o interface{}) error
	GetStatements() []*Statement
	Len() int
}

type rawbody struct {
	lines []*Statement
}

func (r *rawbody) Len() int {
	return len(r.lines)
}

func (r *rawbody) GetStatements() []*Statement {
	return r.lines
}

func (r *rawbody) Type() string {
	return "raw"
}

func (r *rawbody) append(o interface{}) error {
	stm := o.(*Statement)
	r.lines = append(r.lines, stm)
	return nil
}

func (r *rawbody) LastStatement() *Statement {
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
	state     stateL2
	child     []*Block
	parent    *Block
	vars      map[string]*_Var
	blockBody
}

// TODO:
// calcVar be called at running, not parsing
func (b *Block) calcVar(name string) (string, bool) {
	_ = b.vars
	return "", true
}

func (b *Block) String() string {
	if b.blockBody != nil {
		return fmt.Sprintf(`kind="%s", target="%s", operator="%s", tov="%s", bodylen="%d"`, &b.kind, &b.target, &b.operator, &b.typevalue, b.blockBody.Len())
	} else {
		return fmt.Sprintf(`kind="%s", target="%s", operator="%s", tov="%s"`, &b.kind, &b.target, &b.operator, &b.typevalue)
	}
}
