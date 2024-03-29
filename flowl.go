package cofx

import (
	"path/filepath"
	"strings"

	"github.com/skoowoo/cofx/config"
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
		path = strings.TrimPrefix(path, config.PrivateFlowlDir())
		path = strings.TrimPrefix(path, config.BaseFlowlDir())
	}
	return TruncFlowl(path)
}
