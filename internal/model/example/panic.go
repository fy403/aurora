package example

import "errors"

// PanicTask ...
func PanicTask() (string, error) {
	panic(errors.New("oops"))
}
