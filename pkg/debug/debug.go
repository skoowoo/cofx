package debug

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func Open() {
	logrus.SetLevel(logrus.DebugLevel)
}

func Enabled() bool {
	return logrus.IsLevelEnabled(logrus.DebugLevel)
}

func Log(location string, format string, args ...interface{}) {
	if Enabled() {
		format = fmt.Sprintf("In '%s', ", location) + format
		fmt.Printf(format, args...)
	}
}
