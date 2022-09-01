package cofunc

import (
	"path/filepath"
	"strings"

	"github.com/cofunclabs/cofunc/config"
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

// FlowlPath2Name be used to convert the path of a flowl source file to the flow's name.
func FlowlPath2Name(path string, trimpath ...string) string {
	if len(trimpath) > 0 {
		path = strings.TrimPrefix(path, trimpath[0])
	} else {
		path = strings.TrimPrefix(path, config.FlowSourceDir())
	}
	return TruncFlowl(path)
}
