package model

import (
	"errors"
)

func init() {
	// ExtantTaskMap["panic"] = &request.Handler{
	// 	Usage: "无参数, 直接返回string, error",
	// 	Fn:    PanicTask,
	// }
}

// PanicTask ...
func PanicTask() (string, error) {
	panic(errors.New("oops"))
}
