package cofunc

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func WrapErrorf(err error, format string, args ...interface{}) error {
	var builder strings.Builder
	builder.WriteString(err.Error())
	builder.WriteString(": ")
	return fmt.Errorf(builder.String()+format, args...)
}

func ParseErrorf(ln int, err error, format string, args ...interface{}) error {
	var builder strings.Builder
	builder.WriteString(strconv.Itoa(ln))
	builder.WriteString(": ")
	builder.WriteString(err.Error())
	builder.WriteString(": ")
	return fmt.Errorf(builder.String()+format, args...)
}

var (
	ErrTokenNumInLine        error = errors.New("token number not match")
	ErrTokenType             error = errors.New("token type not match")
	ErrTokenValue            error = errors.New("token value not match")
	ErrTokenRegex            error = errors.New("token regex not match")
	ErrTokenCharacterIllegal error = errors.New("token character illegal")
)

func TokenErrorf(ln int, err error, format string, args ...interface{}) error {
	return ParseErrorf(ln, err, format, args...)
}

func TokenTypeErrorf(t *Token, expect TokenType) error {
	return TokenErrorf(t.ln, ErrTokenType, "'%s', actual '%s', expect '%s'", t, t.typ, expect)
}

func TokenValueErrorf(t *Token, expect string) error {
	return TokenErrorf(t.ln, ErrTokenValue, "actual '%s', expect '%s'", t, expect)
}

var (
	ErrStatementUnknow      error = errors.New("unknow statement")
	ErrMapKVIllegal         error = errors.New("map kv format illegal")
	ErrListElemIllegal      error = errors.New("list element format illegal")
	ErrStatementInferFailed error = errors.New("statement infer failed")
)

func StatementErrorf(ln int, err error, format string, args ...interface{}) error {
	return ParseErrorf(ln, err, format, args...)
}

func StatementTokensErrorf(err error, tokens []*Token) error {
	if len(tokens) == 0 {
		return err
	}
	var builder strings.Builder
	ln := 0
	for _, t := range tokens {
		ln = t.ln
		builder.WriteString("'" + t.String() + "'")
		builder.WriteString(" ")
	}
	return ParseErrorf(ln, err, "%s", builder.String())
}

var (
	ErrVariableFormat         error = errors.New("variable format illegal")
	ErrVariableNameEmpty      error = errors.New("variable name is empty")
	ErrVariableNameDuplicated error = errors.New("variable name is duplicated")
	ErrVariableNotDefined     error = errors.New("variable not defined")
	ErrVariableHasCycle       error = errors.New("variable has cycle")
	ErrVariableValueType      error = errors.New("variable's value type illegal")
)

func VarErrorf(ln int, err error, format string, args ...interface{}) error {
	return ParseErrorf(ln, err, format, args...)
}
