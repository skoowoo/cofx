package config

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

type getdir func() string

func Init() error {
	dirs := []getdir{
		HomeDir,
		LogDir,
		PrivateFlowlDir,
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
	v := os.Getenv("COFX_HOME")
	if len(v) == 0 {
		home, err := homedir.Dir()
		if err != nil {
			panic(err)
		}
		v = filepath.Join(home, ".cofx")
	}
	return v
}

func PrivateFlowlDir() string {
	v := filepath.Join(HomeDir(), "flowls")
	return prettyDirPath(v)
}

func BaseFlowlDir() string {
	v := filepath.Join(GetProgramAbsDir(), "flowls")
	return prettyDirPath(v)
}

func LogDir() string {
	v := filepath.Join(HomeDir(), "logs")
	return prettyDirPath(v)
}

// PrivateShellDir store all functions that's based on shell driver.
func PrivateShellDir() string {
	v := filepath.Join(HomeDir(), "shell")
	return prettyDirPath(v)
}

func BaseShellDir() string {
	v := filepath.Join(GetProgramAbsDir(), "shell")
	return prettyDirPath(v)
}

func prettyDirPath(p string) string {
	return filepath.Clean(p) + "/"
}

// GetProgramAbsDir returns the absolute path of the program, it's also the installed path of the program.
func GetProgramAbsDir() string {
	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return path
}
