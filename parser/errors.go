package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func wrapErrorf(err error, format string, args ...interface{}) error {
	var builder strings.Builder
	builder.WriteString(err.Error())
	builder.WriteString(": ")
	return fmt.Errorf(builder.String()+format, args...)
}

func parseErrorf(ln int, err error, format string, args ...interface{}) error {
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

func tokenErrorf(ln int, err error, format string, args ...interface{}) error {
	return parseErrorf(ln, err, format, args...)
}

func tokenTypeErrorf(t *Token, expect TokenType) error {
	return tokenErrorf(t.ln, ErrTokenType, "'%s', actual '%s', expect '%s'", t, t.typ, expect)
}

func tokenValueErrorf(t *Token, expect string) error {
	return tokenErrorf(t.ln, ErrTokenValue, "actual '%s', expect '%s'", t, expect)
}

var (
	ErrStatementUnknow      error = errors.New("unknow statement")
	ErrMapKVIllegal         error = errors.New("map kv format illegal")
	ErrListElemIllegal      error = errors.New("list element format illegal")
	ErrStatementInferFailed error = errors.New("statement infer failed")
	ErrStatementTooMany     error = errors.New("statement too many")
)

func statementErrorf(ln int, err error, format string, args ...interface{}) error {
	return parseErrorf(ln, err, format, args...)
}

func statementTokensErrorf(err error, tokens []*Token) error {
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
	return parseErrorf(ln, err, "%s", builder.String())
}

var (
	ErrVariableFormat         error = errors.New("variable format illegal")
	ErrVariableNameEmpty      error = errors.New("variable name is empty")
	ErrVariableNameDuplicated error = errors.New("variable name is duplicated")
	ErrVariableNotDefined     error = errors.New("variable not defined")
	ErrVariableHasCycle       error = errors.New("variable has cycle")
	ErrVariableValueType      error = errors.New("variable's value type illegal")
)

func varErrorf(ln int, err error, format string, args ...interface{}) error {
	return parseErrorf(ln, err, format, args...)
}
