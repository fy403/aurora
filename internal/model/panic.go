package model

import (
	"errors"
)

func init() {
	ExtantTaskMap["panic"] = PanicTask
}

// PanicTask ...
func PanicTask() (string, error) {
	panic(errors.New("oops"))
}
