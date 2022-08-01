package textline

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	ErrNotfound = errors.New("not found")
)

func FindFileLine(filepath string, split func(rune) bool, find func([]string) (string, bool)) (string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("%w: file '%s'", err, filepath)
	}
	defer f.Close()

	buf := bufio.NewScanner(f)
	for {
		if !buf.Scan() {
			break
		}
		line := buf.Text()
		fields := strings.FieldsFunc(line, split)
		if s, ok := find(fields); ok {
			return s, nil
		}
	}
	return "", ErrNotfound
}
