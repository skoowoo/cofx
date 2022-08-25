package cofunc

import (
	"path/filepath"
	"strings"
)

const filesuffix = ".flowl"

func IsFlowl(name string) bool {
	return filepath.Ext(name) == filesuffix
}

func TruncFlowl(name string) string {
	if IsFlowl(name) {
		return strings.TrimSuffix(name, filesuffix)
	}
	return name
}
