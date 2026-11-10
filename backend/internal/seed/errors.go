package seed

import "errors"

func seedError(msg, err string) error {
	return errors.New(msg + ": " + err + "\n")
}
