package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

type getdir func() string

func Init() error {
	dirs := []getdir{
		LogDir,
		LogBucketDir,
		FlowSourceDir,
	}
	for _, dir := range dirs {
		_, err := os.Stat(dir())
		if err == nil {
			continue
		}
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir(), 0755); err != nil {
				return err
			}
			continue
		}
		return err
	}
	return nil
}

func LogDir() string {
	v := os.Getenv("CO_LOG_DIR")
	if v == "" {
		v = ".cofunc/logs"
	}
	return prettyDirPath(v)
}

func LogBucketDir() string {
	return path.Join(LogDir(), "buckets")
}

// LogFunctionDir returns the directory path of the function's logger
func LogFunctionDir(flowID string, seq int) (string, error) {
	dir := path.Join(LogBucketDir(), fmt.Sprintf("%s/%d", flowID, seq))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// LogFunctionFile returns the name of function's log file, the argument is the directory where the log file is located
func LogFunctionFile(dir string) string {
	return path.Join(dir, "logfile")
}

func FlowSourceDir() string {
	v := os.Getenv("CO_FLOW_SOURCE_DIR")
	if v == "" {
		v = ".cofunc/flows"
	}
	return prettyDirPath(v)
}

func prettyDirPath(p string) string {
	return filepath.Clean(p) + "/"
}
