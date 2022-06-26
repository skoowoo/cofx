package is

import "unicode"

func Space(x rune) bool {
	return x == ' ' || x == '\t'
}

func SpaceOrEOL(x rune) bool {
	return unicode.IsSpace(x)
}

func EOL(x rune) bool {
	return x == '\n'
}

func LeftBracket(x rune) bool {
	return x == '{'
}

func RightBracket(x rune) bool {
	return x == '}'
}

func Eq(x rune) bool {
	return x == '='
}

func Word(x rune) bool {
	if x >= 'a' && x <= 'z' {
		return true
	}
	if x >= 'A' && x <= 'Z' {
		return true
	}
	if x >= '0' && x <= '9' {
		return true
	}
	if x == '_' || x == '-' {
		return true
	}
	return false
}
