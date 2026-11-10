package data

import (
	"errors"
	"fmt"
)

var ErrOnScan = errors.New("problem on scan")
var ErrOnRowScan = errors.New("problem on scan row")

func SQLError(msg string, err error) error {
	return fmt.Errorf("%s: %w", msg, err)
}
