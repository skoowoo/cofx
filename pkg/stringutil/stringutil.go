package stringutil

import "strings"

// String2Slice convert a string to a slice, split by the separator, the separator is ',' or '\n'.
func String2Slice(s string) []string {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == '\n'
	})
	var ret []string
	for _, field := range fields {
		tm := strings.TrimSpace(field)
		ret = append(ret, tm)
	}
	return ret
}
