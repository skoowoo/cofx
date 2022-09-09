package gobuild

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"
)

type binaryOutFormat struct {
	os      string
	arch    string
	dir     string
	name    string
	addname bool
	isDir   bool
}

func (b binaryOutFormat) bin(name string) string {
	if b.isDir {
		return b.dir
	}
	if b.addname {
		return filepath.Join(b.dir, name+b.name)
	} else {
		return filepath.Join(b.dir, b.name)
	}
}

func (b binaryOutFormat) envs() []string {
	var s []string
	if b.os != "" {
		s = append(s, "GOOS="+b.os)
	}
	if b.arch != "" {
		s = append(s, "GOARCH="+b.arch)
	}
	return s
}

func parseBinout(binout string) (binaryOutFormat, error) {
	f := binaryOutFormat{}
	if strings.HasSuffix(binout, "/") {
		f.dir = binout
		f.isDir = true
	} else {
		f.dir = filepath.Dir(binout)
		f.name = filepath.Base(binout)
		if f.name == "" || f.name == "." || f.name == "/" {
			return binaryOutFormat{}, errors.New("invalid binout format")
		}
		if strings.HasPrefix(f.name, "*") {
			f.name = strings.TrimPrefix(f.name, "*")
			f.addname = true
		}
	}
	ospattern := map[string]string{
		`\Wdarwin\W`:  "darwin",
		`\Wlinux\W`:   "linux",
		`\Wwindows\W`: "windows",
	}
	for pattern, os := range ospattern {
		if match, err := regexp.MatchString(pattern, binout); err != nil {
			return binaryOutFormat{}, err
		} else if match {
			f.os = os
			break
		}
	}
	archpattern := map[string]string{
		`\Wamd64\W`: "amd64",
		`\W386\W`:   "386",
		`\Warm\W`:   "arm",
	}
	for pattern, arch := range archpattern {
		if match, err := regexp.MatchString(pattern, binout); err != nil {
			return binaryOutFormat{}, err
		} else if match {
			f.arch = arch
			break
		}
	}
	return f, nil
}
