package ai

import (
	"fmt"
)

func DBInitError(msg string, err error) error {
	return fmt.Errorf("%s: %q", msg, err)
}
