package cofunc

import (
	"errors"
	"fmt"
)

func errInvalidChar(c byte, line string) error {
	return errors.New("contain invalid character: " + fmt.Sprintf("%c, %s", c, line))
}
