package cofunc

import "strings"

const filesuffix = ".flowl"

func IsFlowl(name string) bool {
	return strings.HasSuffix(name, filesuffix)
}

func TruncFlowl(name string) string {
	if IsFlowl(name) {
		return strings.TrimSuffix(name, filesuffix)
	}
	return name
}
