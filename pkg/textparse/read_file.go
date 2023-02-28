package textparse

import (
	"os"
	"strconv"
	"strings"
)

type FileResult struct {
	value []byte
	err   error
	path  string
}

// ReadFile reads the file content and save it into the FileResult.
// If filepath not exist, try to find a exist file in 'files'.
func ReadFile(filepath string, files ...string) FileResult {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		for _, f := range files {
			if _, err := os.Stat(f); os.IsNotExist(err) {
				continue
			}
			filepath = f
			break
		}
	}

	b, err := os.ReadFile(filepath)
	if err != nil {
		return FileResult{
			err:  err,
			path: filepath,
		}
	}
	return FileResult{
		value: b,
		path:  filepath,
	}
}

// String returns the file content as string, returns the file path at the same time.
func (r FileResult) String() (string, string, error) {
	if r.err != nil && !os.IsNotExist(r.err) {
		return "", "", r.err
	}
	return r.path, strings.TrimSpace(string(r.value)), nil
}

// Int returns the file content as int, returns the file path at the same time.
func (r FileResult) Int() (string, int, error) {
	if r.err != nil {
		return "", 0, r.err
	}
	s := string(r.value)
	n, err := strconv.Atoi(s)
	if err != nil {
		return "", 0, err
	}
	return r.path, n, nil
}
