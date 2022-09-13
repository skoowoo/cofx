package gobuild

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBinOut(t *testing.T) {
	name := "test"
	{
		testingdata := map[string]string{
			"bin/":               "bin/test",
			"bin":                "bin",
			"bin//":              "bin/test",
			"bin/*":              "bin/test",
			"bin/*-hello":        "bin/test-hello",
			"bin/*-hello-":       "bin/test-hello-",
			"bin/a-*":            "bin/a-*",
			"bin/darwin/*.hello": "bin/darwin/test.hello",
			"bin/darwin/hello":   "bin/darwin/hello",
		}
		for k, v := range testingdata {
			bin, err := parseBinFormat(k)
			if err != nil {
				assert.FailNow(t, err.Error())
			}
			assert.Equal(t, v, bin.fullBinPath(name))
		}
	}
	{
		bin, err := parseBinFormat("bin/darwin/*.hello")
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		envs := bin.envs()
		assert.Equal(t, "CGO_ENABLED=0", envs[0])
		assert.Equal(t, "GOOS=darwin", envs[1])
		assert.Equal(t, "GOARCH=amd64", envs[2])
	}
	{
		bin, err := parseBinFormat("bin/*.linux")
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		envs := bin.envs()
		assert.Equal(t, "CGO_ENABLED=0", envs[0])
		assert.Equal(t, "GOOS=linux", envs[1])
		assert.Equal(t, "GOARCH=amd64", envs[2])
	}
	{
		bin, err := parseBinFormat("bin/*.windows.386")
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		envs := bin.envs()
		assert.Equal(t, "CGO_ENABLED=0", envs[0])
		assert.Equal(t, "GOOS=windows", envs[1])
		assert.Equal(t, "GOARCH=386", envs[2])
	}
	{
		bin, err := parseBinFormat("bindarwin/*.hello")
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		envs := bin.envs()
		assert.Len(t, envs, 0)
	}
}
