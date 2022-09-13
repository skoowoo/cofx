package gobuild

import (
	"errors"
	"fmt"
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

func (b binaryOutFormat) fullBinPath(name string) string {
	if b.isDir {
		return filepath.Join(b.dir, name)
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
		s = append(s, "CGO_ENABLED=0")
		s = append(s, "GOOS="+b.os)
		if b.arch != "" {
			s = append(s, "GOARCH="+b.arch)
		}
	}
	return s
}

func parseBinFormats(formats []string) ([]binaryOutFormat, error) {
	var bins []binaryOutFormat
	for _, format := range formats {
		if bin, err := parseBinFormat(format); err != nil {
			return nil, fmt.Errorf("%w: %s", err, format)
		} else {
			bins = append(bins, bin)
		}
	}
	return bins, nil
}

func parseBinFormat(format string) (binaryOutFormat, error) {
	f := binaryOutFormat{}
	if strings.HasSuffix(format, "/") {
		f.dir = format
		f.isDir = true
	} else {
		f.dir = filepath.Dir(format)
		f.name = filepath.Base(format)
		if f.name == "" || f.name == "." || f.name == "/" {
			return binaryOutFormat{}, errors.New("invalid binout format")
		}
		if strings.HasPrefix(f.name, "*") {
			f.name = strings.TrimPrefix(f.name, "*")
			f.addname = true
		}
	}
	ospattern := map[string]string{
		`\Wdarwin\W`: "darwin",
		`^darwin\W`:  "darwin",
		`\Wdarwin$`:  "darwin",

		`\Wlinux\W`: "linux",
		`\Wlinux$`:  "linux",
		`^linux\W`:  "linux",

		`\Wwindows\W`: "windows",
		`^windows\W`:  "windows",
		`\Wwindows$`:  "windows",
	}
	for pattern, os := range ospattern {
		if match, err := regexp.MatchString(pattern, format); err != nil {
			return binaryOutFormat{}, err
		} else if match {
			f.os = os
			break
		}
	}
	archpattern := map[string]string{
		`\Wamd64\W`: "amd64",
		`^amd64\W`:  "amd64",
		`\Wamd64$`:  "amd64",

		`\W386\W`: "386",
		`^386\W`:  "386",
		`\W386$`:  "386",

		`\Warm\W`: "arm",
		`^arm\W`:  "arm",
		`\Warm$`:  "arm",
	}
	for pattern, arch := range archpattern {
		if match, err := regexp.MatchString(pattern, format); err != nil {
			return binaryOutFormat{}, err
		} else if match {
			f.arch = arch
			break
		}
	}
	if f.os != "" && f.arch == "" {
		f.arch = "amd64"
	}
	return f, nil
}
