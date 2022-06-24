package flowl

import "strings"

const filesuffix = ".flowl"

func IsFlowl(name string) bool {
	return strings.HasSuffix(name, filesuffix)
}
