package example

import (
	"aurora/internal/model"
	"errors"
)

func init() {
	model.ExtantTaskMap["panic"] = PanicTask
}

// PanicTask ...
func PanicTask() (string, error) {
	panic(errors.New("oops"))
}
