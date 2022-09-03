package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

type getdir func() string

func Init() error {
	dirs := []getdir{
		HomeDir,
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

func HomeDir() string {
	v := os.Getenv("COFUNC_HOME")
	if len(v) == 0 {
		home, err := homedir.Dir()
		if err != nil {
			panic(err)
		}
		v = filepath.Join(home, ".cofunc")
	}
	return v
}

func FlowSourceDir() string {
	v := filepath.Join(HomeDir(), "flowls")
	return prettyDirPath(v)
}

func LogDir() string {
	v := filepath.Join(HomeDir(), "logs")
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

func prettyDirPath(p string) string {
	return filepath.Clean(p) + "/"
}
