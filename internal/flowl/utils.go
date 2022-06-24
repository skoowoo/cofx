package flowl

import (
	"errors"
	"strings"
)

func ValidateFileName(name string) error {
	if strings.HasSuffix(name, filesuffix) {
		return nil
	}
	return errors.New("invalid flowl filename, miss suffix '.flowl': " + name)
}
